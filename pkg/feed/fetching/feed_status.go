// Copyright (c) VirtualTam
// SPDX-License-Identifier: MIT

package fetching

import "github.com/mmcdole/gofeed"

type FeedStatus struct {
	StatusCode int
	ETag       string
	Feed       *gofeed.Feed
}
