// Copyright (c) VirtualTam
// SPDX-License-Identifier: MIT

package feed

import (
	"time"
)

// Entry represents an entry of a syndication feed (Atom or RSS).
type Entry struct {
	UID      string
	FeedUUID string

	URL   string
	Title string

	PublishedAt time.Time
	UpdatedAt   time.Time
}
