// Copyright (c) VirtualTam
// SPDX-License-Identifier: MIT

package feed

import (
	"slices"
	"time"
)

// Preferences represents a user's preferences.
type Preferences struct {
	UserUUID           string
	ShowEntries        EntryVisibility
	ShowEntrySummaries bool

	UpdatedAt time.Time
}

// EntryVisibility allows users to choose which feed entries to display, according to their read status.
type EntryVisibility string

const (
	// EntryVisibilityAll indicates the user wants to display all feed entries.
	EntryVisibilityAll EntryVisibility = "ALL"

	// EntryVisibilityRead indicates the user wants to display only read entries.
	EntryVisibilityRead EntryVisibility = "READ"

	// EntryVisibilityUnread indicates the user wants to display only unread entries.
	EntryVisibilityUnread EntryVisibility = "UNREAD"
)

var (
	allEntryVisibilities = []EntryVisibility{EntryVisibilityAll, EntryVisibilityRead, EntryVisibilityUnread}
)

// ValidateForUpdate ensures mandatory fields are set when updating existing Preferences.
func (p *Preferences) ValidateForUpdate() error {
	if !slices.Contains(allEntryVisibilities, p.ShowEntries) {
		return ErrPreferencesEntryVisibilityUnknown
	}

	return nil
}
