// Copyright VirtualTam 2022, 2026
// SPDX-License-Identifier: MIT

package controller

import (
	"github.com/virtualtam/sparklemuffin/pkg/feed"
	feedquerying "github.com/virtualtam/sparklemuffin/pkg/feed/querying"
)

type feedQueryingPage struct {
	feedquerying.FeedPage

	URLPath string

	Preferences feed.Preferences
}
