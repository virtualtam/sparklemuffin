// Copyright (c) VirtualTam
// SPDX-License-Identifier: MIT

package feed

// EntryMetadata tracks user-specific metadata for a given feed entry.
type EntryMetadata struct {
	UserUUID string
	EntryUID string

	Read bool
}
