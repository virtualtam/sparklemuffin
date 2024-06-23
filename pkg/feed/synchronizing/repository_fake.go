// Copyright (c) VirtualTam
// SPDX-License-Identifier: MIT

package synchronizing

import (
	"time"

	"github.com/virtualtam/sparklemuffin/pkg/feed"
)

var _ Repository = &fakeRepository{}

type fakeRepository struct {
	Feeds   []feed.Feed
	Entries []feed.Entry
}

func (r *fakeRepository) FeedGetNByLastSynchronizationTime(n uint, lastSyncBefore time.Time) ([]feed.Feed, error) {
	feedsToSync := []feed.Feed{}

	for _, f := range r.Feeds {
		if f.FetchedAt.After(lastSyncBefore) {
			continue
		}

		feedsToSync = append(feedsToSync, f)
	}

	return feedsToSync, nil
}

func (r *fakeRepository) FeedUpdateFetchedAt(updatedFeed feed.Feed) error {
	for index, f := range r.Feeds {
		if f.UUID == updatedFeed.UUID {
			r.Feeds[index] = updatedFeed
			return nil
		}
	}

	return feed.ErrFeedNotFound
}

func (r *fakeRepository) FeedEntryUpsertMany(newEntries []feed.Entry) (int64, error) {
	uniqueURLs := map[string]int{}
	for index, entry := range r.Entries {
		uniqueURLs[entry.URL] = index
	}

	var newOrUpdated int64

	for _, newEntry := range newEntries {
		if index, ok := uniqueURLs[newEntry.URL]; ok {
			// entry already exists
			r.Entries[index].Title = newEntry.Title
			r.Entries[index].UpdatedAt = newEntry.UpdatedAt

			newOrUpdated++

			continue
		}

		r.Entries = append(r.Entries, newEntry)
		uniqueURLs[newEntry.URL] = len(r.Entries) - 1
		newOrUpdated++
	}

	return newOrUpdated, nil
}
