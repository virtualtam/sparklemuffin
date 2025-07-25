// Copyright (c) VirtualTam
// SPDX-License-Identifier: MIT

package synchronizing

import (
	"time"

	"github.com/virtualtam/sparklemuffin/pkg/feed"
)

// Repository provides access to feed data for synchronizing.
type Repository interface {
	// FeedGetNByLastSynchronizationTime returns at most n feeds that have last been synchronized before
	// a given time.Time.
	//
	// This method must return only feeds with at least one active user Subscription.
	FeedGetNByLastSynchronizationTime(n uint, before time.Time) ([]feed.Feed, error)

	// FeedUpdateFetchMetadata updates fetch metadata (ETag, FetchedAt, UpdatedAt) for a given feed.Feed.
	FeedUpdateFetchMetadata(feedFetchMetadata FeedFetchMetadata) error

	// FeedUpdateMetadata updates metadata (Title, Description) for a given feed.Feed.
	FeedUpdateMetadata(feedMetadata FeedMetadata) error

	// FeedEntryUpsertMany adds a collection of new entries and updates existing entries.
	FeedEntryUpsertMany(entries []feed.Entry) (int64, error)
}
