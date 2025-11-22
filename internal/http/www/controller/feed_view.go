// Copyright (c) VirtualTam
// SPDX-License-Identifier: MIT

package controller

import (
	"github.com/virtualtam/sparklemuffin/pkg/feed"
	feedquerying "github.com/virtualtam/sparklemuffin/pkg/feed/querying"
)

type feedQueryingPage struct {
	feedquerying.FeedPage

	CSRFToken string
	URLPath   string

	Preferences feed.Preferences
}
