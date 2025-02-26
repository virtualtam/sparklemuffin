// Copyright (c) VirtualTam
// SPDX-License-Identifier: MIT

package exporting

import "time"

type jsonDocument struct {
	Title      string    `json:"title"`
	ExportedAt time.Time `json:"exported_at"`

	Bookmarks []jsonBookmark `json:"bookmarks"`
}

type jsonBookmark struct {
	URL         string `json:"url"`
	Title       string `json:"title"`
	Description string `json:"description,omitempty"`

	Private bool     `json:"private"`
	Tags    []string `json:"tags,omitempty"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
