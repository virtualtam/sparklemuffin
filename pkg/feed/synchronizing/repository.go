// Copyright (c) VirtualTam
// SPDX-License-Identifier: MIT

package synchronizing

import (
	"context"
	"time"

	"github.com/virtualtam/sparklemuffin/pkg/feed"
)

// Repository provides access to feed data for synchronizing.
type Repository interface {
	// FeedGetNByLastSynchronizationTime returns at most n feeds that have last been synchronized before
	// a given time.Time.
	//
	// This method must only return feeds with at least one active user Subscription.
	FeedGetNByLastSynchronizationTime(ctx context.Context, n uint, before time.Time) ([]feed.Feed, error)

	// FeedUpdateFetchMetadata updates fetch metadata (ETag, FetchedAt, UpdatedAt) for a given feed.Feed.
	FeedUpdateFetchMetadata(ctx context.Context, feedFetchMetadata FeedFetchMetadata) error

	// FeedUpdateMetadata updates metadata (Title, Description) for a given feed.Feed.
	FeedUpdateMetadata(ctx context.Context, feedMetadata FeedMetadata) error

	// FeedEntryUpsertMany adds a collection of new entries and updates existing entries.
	FeedEntryUpsertMany(ctx context.Context, entries []feed.Entry) (int64, error)
}
