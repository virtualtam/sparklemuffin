// Copyright (c) VirtualTam
// SPDX-License-Identifier: MIT

package importing

import "fmt"

// Status provides details about the imported data.
type Status struct {
	overwriteExisting bool

	// Invalid entries that did not correspond to valid bookmark entries (URL, Title).
	Invalid int

	// New bookmarks, or existing bookmarks that were updated.
	NewOrUpdated int

	// Skipped bookmarks (existing bookmarks were kept).
	Skipped int
}

// Summary returns a string formatted with import information.
func (st *Status) Summary() string {
	var orUpdated string
	if st.overwriteExisting {
		orUpdated = " or updated"
	}

	return fmt.Sprintf(
		"%d new%s, %d skipped, %d invalid",
		st.NewOrUpdated,
		orUpdated,
		st.Skipped,
		st.Invalid,
	)
}
