// Copyright (c) VirtualTam
// SPDX-License-Identifier: MIT

package exporting

import "errors"

// Visibility represents a visibility filter for bookmarks.
type Visibility string

const (
	// VisibilityAll indicates public and private bookmarks will be exported.
	VisibilityAll Visibility = "all"

	// VisibilityPrivate indicates only private bookmarks will be exported.
	VisibilityPrivate Visibility = "private"

	// VisibilityPublic indicates only public bookmarks will be exported.
	VisibilityPublic Visibility = "public"
)

var (
	ErrVisibilityInvalid = errors.New("invalid value for visibility")
)
