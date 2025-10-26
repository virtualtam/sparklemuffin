// Copyright (c) VirtualTam
// SPDX-License-Identifier: MIT

package feed

import (
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

// Preferences represents a user's preferences.
type Preferences struct {
	UserUUID    string
	ShowEntries EntryVisibility

	UpdatedAt time.Time
}
