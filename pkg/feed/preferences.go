// Copyright (c) VirtualTam
// SPDX-License-Identifier: MIT

package feed

import (
	"slices"
	"time"
)

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

// Preferences represents a user's preferences.
type Preferences struct {
	UserUUID    string
	ShowEntries EntryVisibility

	UpdatedAt time.Time
}

// ValidateForUpdate ensures mandatory fields are properly set when updating existing Preferences.
func (p *Preferences) ValidateForUpdate() error {
	if !slices.Contains(allEntryVisibilities, p.ShowEntries) {
		return ErrPreferencesEntryVisibilityUnknown
	}

	return nil
}
