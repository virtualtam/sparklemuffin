// Copyright VirtualTam 2022, 2026
// SPDX-License-Identifier: MIT

package feed

// EntryMetadata tracks user-specific metadata for a given feed entry.
type EntryMetadata struct {
	UserUUID string
	EntryUID string

	Read bool
}
