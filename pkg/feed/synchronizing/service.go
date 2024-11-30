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
	feedsToSynchronize uint          = 20
	minFeedAge         time.Duration = 6 * time.Hour

	nWorkers int = 5
)

// Service handles feed synchronization operations.
type Service struct {
	r Repository

	client *fetching.Client

	textRanker       *textkit.TextRanker
	textRankMaxTerms int
}

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
			// 3.1 Fetch entries
			// 3.2 Upsert entries
			// 3.3 Update FetchedAt date
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
		return err
	}

	if feedStatus.StatusCode == http.StatusNotModified {
		// the remote server responds with a '304 Not Modified' status, indicating that
		// we already have the latest version of the feed

		log.Info().
			Str("feed_url", feed.FeedURL).
			Str("job_id", jobID).
			Msg("feeds: already up-to-date, nothing to do")

		return nil
	}

	if feedStatus.Feed.Title != feed.Title || feedStatus.Feed.Description != feed.Description {
		feedMetadata := FeedMetadata{
			UUID:        feed.UUID,
			Title:       feedStatus.Feed.Title,
			Description: feedStatus.Feed.Description,
			UpdatedAt:   now,
		}

		if err := s.r.FeedUpdateMetadata(feedMetadata); err != nil {
			return err
		}
	}

	rowsAffected, err := s.createOrUpdateEntries(feed, now, feedStatus.Feed.Items)
	if err != nil {
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

		if err := entry.ValidateForAddition(); err != nil {
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
		return 0, err
	}

	return rowsAffected, nil
}
