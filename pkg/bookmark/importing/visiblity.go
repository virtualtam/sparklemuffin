// Copyright (c) VirtualTam
// SPDX-License-Identifier: MIT

package importing

import "errors"

// Visibility defines how bookmarks are imported.
type Visibility string

const (
	// VisibilityDefault indicates bookmarks will be imported with their existing visibility (if set),
	// otherwise as public.
	VisibilityDefault Visibility = "default"

	// VisibilityPrivate indicates all bookmarks will be imported as private.
	VisibilityPrivate Visibility = "private"

	// VisibilityPublic indicates all bookmarks will be imported as public.
	VisibilityPublic Visibility = "public"
)

var (
	ErrVisibilityInvalid = errors.New("invalid value for visibility")
)
