// Copyright (c) VirtualTam
// SPDX-License-Identifier: MIT

package feedtest

import (
	"testing"
	"time"

	"github.com/gorilla/feeds"
)

func GenerateDummyFeed(t *testing.T, now time.Time) feeds.Feed {
	t.Helper()

	yesterday := now.Add(-24 * time.Hour)

	return feeds.Feed{
		Title:   "Local Test",
		Updated: now,
		Items: []*feeds.Item{
			{
				Id:    "http://test.local/first-post",
				Title: "First post!",
				Link: &feeds.Link{
					Href: "http://test.local/first-post",
				},
				Created: now,
				Updated: now,
			},
			{
				Id:    "http://test.local/hello-world",
				Title: "Hello World",
				Link: &feeds.Link{
					Href: "http://test.local/hello-world",
				},
				Created: yesterday,
				Updated: yesterday,
			},
		},
	}
}
