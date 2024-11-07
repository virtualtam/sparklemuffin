// Copyright (c) VirtualTam
// SPDX-License-Identifier: MIT

package exporting

import "github.com/virtualtam/sparklemuffin/pkg/feed"

type CategorySubscriptions struct {
	feed.Category

	SubscribedFeeds []feed.Feed
}
