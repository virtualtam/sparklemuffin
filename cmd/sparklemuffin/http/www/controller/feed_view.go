// Copyright (c) VirtualTam
// SPDX-License-Identifier: MIT

package controller

import (
	feedquerying "github.com/virtualtam/sparklemuffin/pkg/feed/querying"
)

type feedQueryingPage struct {
	feedquerying.FeedPage

	CSRFToken string
}
