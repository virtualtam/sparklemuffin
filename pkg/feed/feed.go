// Copyright (c) VirtualTam
// SPDX-License-Identifier: MIT

package feed

import (
	"time"
)

// Feed represents a Web syndication feed (Atom or RSS).
type Feed struct {
	UUID string

	FeedURL string
	Title   string
	Slug    string

	CreatedAt      time.Time
	UpdatedAt      time.Time
	SynchronizedAt time.Time
}
