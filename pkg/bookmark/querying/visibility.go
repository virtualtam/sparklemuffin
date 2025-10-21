// Copyright (c) VirtualTam
// SPDX-License-Identifier: MIT

package querying

// Visibility represents a visibility filter for bookmarks.
type Visibility string

const (
	// VisibilityAll indicates public and private bookmarks will be returned.
	VisibilityAll Visibility = "all"

	// VisibilityPrivate indicates only private bookmarks will be returned.
	VisibilityPrivate Visibility = "private"

	// VisibilityPublic indicates only public bookmarks will be returned.
	VisibilityPublic Visibility = "public"
)
