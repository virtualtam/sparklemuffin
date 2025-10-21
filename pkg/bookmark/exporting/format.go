// Copyright (c) VirtualTam
// SPDX-License-Identifier: MIT

package exporting

// Format represents the file format used for the export.
type Format string

const (
	// FormatJSON indicates bookmarks will be exported as a JSON document.
	FormatJSON Format = "json"

	// FormatNetscape indicates bookmarks will be exported as a Netscape Bookmark File.
	FormatNetscape Format = "netscape"
)
