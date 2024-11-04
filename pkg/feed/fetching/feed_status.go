// Copyright (c) VirtualTam
// SPDX-License-Identifier: MIT

package fetching

import (
	"time"

	"github.com/mmcdole/gofeed"
)

// FeedStatus represents the status of a remote feed after fetching.
//
// The ETag and LastModified fields are populated from the headers sent by the
// remote server, and should be saved to leverage server caching in future
// requests.
//
// If the remote server responded with `200 OK`, the Feed will be populated with
// data parsed from the response.
//
// If the remote server responded with '304 Not Modified', the Feed will be nil.
type FeedStatus struct {
	StatusCode   int
	ETag         string
	LastModified time.Time
	Feed         *gofeed.Feed
}
