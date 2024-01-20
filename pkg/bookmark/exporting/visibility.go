// Copyright (c) VirtualTam
// SPDX-License-Identifier: MIT

package exporting

import "errors"

// Visibility represents a visibility filter for bookmarks.
type Visibility string

const (
	// Public and private bookmarks.
	VisibilityAll Visibility = "all"

	// Private bookmarks only.
	VisibilityPrivate Visibility = "private"

	// Public bookmarks only.
	VisibilityPublic Visibility = "public"
)

var (
	ErrVisibilityInvalid error = errors.New("invalid value for visibility")
)
