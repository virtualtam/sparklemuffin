// Copyright (c) VirtualTam
// SPDX-License-Identifier: MIT

package synchronizing

import "time"

// FeedFetchMetadata represents the metadata for feed fetch information.
type FeedFetchMetadata struct {
	UUID string

	ETag         string
	LastModified time.Time

	UpdatedAt time.Time
	FetchedAt time.Time
}

// FeedMetadata represents the metadata for a feed and its content.
type FeedMetadata struct {
	UUID string

	Title       string
	Description string

	Hash uint64

	UpdatedAt time.Time
}
