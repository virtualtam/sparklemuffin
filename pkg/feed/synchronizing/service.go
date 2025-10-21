// Copyright (c) VirtualTam
// SPDX-License-Identifier: MIT

package synchronizing

import (
	"net/http"
	"time"

	"github.com/mmcdole/gofeed"
	"github.com/rs/zerolog/log"
	"github.com/sourcegraph/conc/pool"

	"github.com/virtualtam/sparklemuffin/internal/textkit"
	"github.com/virtualtam/sparklemuffin/pkg/feed"
	"github.com/virtualtam/sparklemuffin/pkg/feed/fetching"
)

const (
	feedsToSynchronize uint = 20
	minFeedAge              = 6 * time.Hour

	nWorkers int = 5
)

// Service handles feed synchronization operations.
type Service struct {
	r Repository

	client *fetching.Client

	textRanker       *textkit.TextRanker
	textRankMaxTerms int
}

// NewService initializes and returns a new feed synchronization service.
func NewService(r Repository, client *fetching.Client) *Service {
	return &Service{
		r:                r,
		client:           client,
		textRanker:       textkit.NewTextRanker(),
		textRankMaxTerms: feed.EntryTextRankMaxTerms,
	}
}

// Synchronize synchronizes syndication feeds for all users.
func (s *Service) Synchronize(jobID string) error {
	lastSyncBefore := time.Now().UTC().Add(-minFeedAge)

	// 1. List all feeds that have last been synchronized before a given time.Time
	feeds, err := s.r.FeedGetNByLastSynchronizationTime(feedsToSynchronize, lastSyncBefore)
	if err != nil {
		log.
			Error().
			Err(err).
			Str("job_id", jobID).
			Msg("feeds: failed to list feeds to synchronize")
		return err
	}

	if len(feeds) == 0 {
		log.Info().Msg("feeds: nothing to synchronize")
		return nil
	}

	// 2. Start a concurrent worker pool
	workerPool := pool.New().WithErrors().WithMaxGoroutines(nWorkers)
	log.Debug().
		Int("n_workers", nWorkers).
		Msg("feeds: synchronization worker pool started")

	// 3. For each feed:
	for _, workerFeed := range feeds {
		workerPool.Go(func() error {
			// 3.1. Fetch entries
			// 3.2. Upsert entries
			// 3.3. Update FetchedAt date
			return s.synchronizeFeed(workerFeed, jobID)
		})
	}

	if err := workerPool.Wait(); err != nil {
		log.
			Error().
			Err(err).
			Str("job_id", jobID).
			Msg("feeds: failed to synchronize some feeds")
		return err
	}

	return nil
}

func (s *Service) synchronizeFeed(feed feed.Feed, jobID string) error {
	log.
		Info().
		Str("feed_url", feed.FeedURL).
		Str("job_id", jobID).
		Msg("feeds: synchronizing")

	feedStatus, err := s.client.Fetch(feed.FeedURL, feed.ETag, feed.LastModified)
	if err != nil {
		log.
			Error().
			Err(err).
			Str("feed_url", feed.FeedURL).
			Str("job_id", jobID).
			Msg("feeds: failed to fetch feed")
		return err
	}

	now := time.Now().UTC()

	feedFetchMetadata := FeedFetchMetadata{
		UUID:         feed.UUID,
		ETag:         feedStatus.ETag,
		LastModified: feedStatus.LastModified,
		UpdatedAt:    now,
		FetchedAt:    now,
	}

	if err := s.r.FeedUpdateFetchMetadata(feedFetchMetadata); err != nil {
		log.
			Error().
			Err(err).
			Str("feed_url", feed.FeedURL).
			Str("job_id", jobID).
			Msg("feeds: failed to update fetch metadata")
		return err
	}

	if feedStatus.StatusCode == http.StatusNotModified {
		// The remote server responds with a '304 Not Modified' status, indicating that
		// we already have the latest version of the feed.

		log.Info().
			Str("feed_url", feed.FeedURL).
			Str("job_id", jobID).
			Str("reason", "304 Not Modified").
			Msg("feeds: skipping update, remote content not modified")

		return nil
	}

	if feedStatus.Hash == feed.Hash {
		// The feed data returned by the remote server is the same as the one we already have,
		// or the remote server does not support HTTP conditional requests
		// (no ETag nor Last-Modified headers are set).
		//
		// See https://inessential.com/2024/08/03/netnewswire_and_conditional_get_issues.html
		log.Info().
			Str("feed_url", feed.FeedURL).
			Str("job_id", jobID).
			Str("reason", "hashes match").
			Msg("feeds: skipping update, remote content not modified")
		return nil
	}

	if feedStatus.Feed.Title != feed.Title || feedStatus.Feed.Description != feed.Description || feedStatus.Hash != feed.Hash {
		feedMetadata := FeedMetadata{
			UUID:        feed.UUID,
			Title:       feedStatus.Feed.Title,
			Description: feedStatus.Feed.Description,
			Hash:        feedStatus.Hash,
			UpdatedAt:   now,
		}

		if err := s.r.FeedUpdateMetadata(feedMetadata); err != nil {
			log.
				Error().
				Err(err).
				Str("feed_url", feed.FeedURL).
				Str("job_id", jobID).
				Msg("feeds: failed to update metadata")
			return err
		}
	}

	rowsAffected, err := s.createOrUpdateEntries(feed, now, feedStatus.Feed.Items)
	if err != nil {
		log.
			Error().
			Err(err).
			Str("feed_url", feed.FeedURL).
			Str("job_id", jobID).
			Msg("feeds: failed to create or update entries")
		return err
	}

	log.Info().
		Str("feed_url", feed.FeedURL).
		Str("job_id", jobID).
		Int64("n_entries", rowsAffected).
		Msg("feeds: entries created or updated")

	return nil
}

func (s *Service) createOrUpdateEntries(f feed.Feed, now time.Time, items []*gofeed.Item) (int64, error) {
	var entries []feed.Entry

	for _, item := range items {
		entry := feed.NewEntryFromItem(f.UUID, now, item)
		entry.ExtractTextRankTerms(s.textRanker, s.textRankMaxTerms)

		if err := entry.ValidateForAddition(now); err != nil {
			log.
				Warn().
				Err(err).
				Str("feed_uuid", entry.FeedUUID).
				Str("entry_url", entry.URL).
				Msg("feeds: skipping invalid entry")
			continue
		}

		entries = append(entries, entry)
	}

	rowsAffected, err := s.r.FeedEntryUpsertMany(entries)
	if err != nil {
		log.
			Error().
			Err(err).
			Str("feed_uuid", f.UUID).
			Msg("feeds: failed to create or update entries")
		return 0, err
	}

	return rowsAffected, nil
}
