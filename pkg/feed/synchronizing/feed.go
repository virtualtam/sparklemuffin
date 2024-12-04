// Copyright (c) VirtualTam
// SPDX-License-Identifier: MIT

package synchronizing

import "time"

type FeedFetchMetadata struct {
	UUID string

	ETag         string
	LastModified time.Time

	UpdatedAt time.Time
	FetchedAt time.Time
}

type FeedMetadata struct {
	UUID string

	Title       string
	Description string

	Hash uint64

	UpdatedAt time.Time
}
