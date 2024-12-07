// Copyright (c) VirtualTam
// SPDX-License-Identifier: MIT

package controller

import (
	feedquerying "github.com/virtualtam/sparklemuffin/pkg/feed/querying"
)

type feedQueryingPage struct {
	feedquerying.FeedPage

	CSRFToken string
	URLPath   string
}

type feedEntryMetadataMarkReadForm struct {
	CSRFToken string `schema:"csrf_token"`
}
