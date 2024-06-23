// Copyright (c) VirtualTam
// SPDX-License-Identifier: MIT

package synchronizing

import (
	"net/http"
	"time"

	"github.com/mmcdole/gofeed"
	"github.com/rs/zerolog/log"
	"github.com/sourcegraph/conc/pool"
	"github.com/virtualtam/sparklemuffin/pkg/feed"
)

const (
	feedsToSynchronize uint          = 20
	minFeedAge         time.Duration = 6 * time.Hour

	nWorkers int = 5
)

// Service handles feed synchronization operations.
type Service struct {
	r          Repository
	feedParser *gofeed.Parser
}

func NewService(r Repository, httpClient *http.Client) *Service {
	feedParser := gofeed.NewParser()
	feedParser.Client = httpClient

	return &Service{
		r:          r,
		feedParser: feedParser,
	}
}

// Run synchronizes syndication feeds for all users.
func (s *Service) Run() error {
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
			return s.synchronizeFeed(workerFeed)
		})
	}

	if err := workerPool.Wait(); err != nil {
		log.Error().Err(err).Msg("feeds: failed to synchronize some feeds")
		return err
	}

	return nil
}

func (s *Service) synchronizeFeed(feed feed.Feed) error {
	log.Info().Str("feed_url", feed.FeedURL).Msg("feeds: synchronizing")

	now := time.Now().UTC()

	syndicationFeed, err := s.feedParser.ParseURL(feed.FeedURL)
	if err != nil {
		return err
	}

	rowsAffected, err := s.createOrUpdateEntries(feed, now, syndicationFeed.Items)
	if err != nil {
		return err
	}

	log.Info().
		Str("feed_url", feed.FeedURL).
		Int64("n_entries", rowsAffected).
		Msg("feeds: entries created or updated")

	feed.FetchedAt = now

	return s.r.FeedUpdateFetchedAt(feed)
}

func (s *Service) createOrUpdateEntries(f feed.Feed, now time.Time, items []*gofeed.Item) (int64, error) {
	var entries []feed.Entry

	for _, item := range items {
		entry := feed.NewEntryFromItem(f.UUID, now, item)

		if err := entry.ValidateForAddition(); err != nil {
			log.Warn().Err(err).Msg("skipping invalid entry")
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
