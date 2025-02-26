// Copyright (c) VirtualTam
// SPDX-License-Identifier: MIT

package exporting

// Format represents the file format used for the export.
type Format string

const (
	// JSON document.
	FormatJSON Format = "json"

	// Netscape Bookmark File.
	FormatNetscape Format = "netscape"
)
