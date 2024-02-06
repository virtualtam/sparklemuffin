// Copyright (c) VirtualTam
// SPDX-License-Identifier: MIT

package postgresql

import (
	"time"

	"github.com/virtualtam/sparklemuffin/pkg/feed"
	fquerying "github.com/virtualtam/sparklemuffin/pkg/feed/querying"
)

var _ feed.Repository = &Repository{}
var _ fquerying.Repository = &Repository{}

var (
	linuxCategory = feed.Category{
		Name: "Linux",
		Slug: "linux",
	}
	debianFeed = feed.Feed{
		Title: "Bits from Debian",
		Slug:  "bits-from-debian",
	}
)

func (r *Repository) FeedGetCategories(userUUID string) ([]fquerying.Category, error) {
	feedCategories := []fquerying.Category{
		{
			Category: linuxCategory,
			Unread:   3,
			Feeds: []fquerying.Feed{
				{
					Feed:   debianFeed,
					Unread: 2,
				},
				{
					Feed: feed.Feed{
						Title: "Ubuntu News",
						Slug:  "ubuntu-news",
					},
					Unread: 1,
				},
			},
		},
	}

	return feedCategories, nil
}

func (r *Repository) FeedGetEntriesByPage(userUUID string) (fquerying.FeedEntries, error) {
	feedEntries := fquerying.FeedEntries{
		Category: linuxCategory,
		Feed:     debianFeed,
		Entries: []feed.Entry{
			{
				Title:       "Bits from the DPL (with description)",
				Description: "Something to announce here!",
				PublishedAt: time.Date(2024, 4, 7, 14, 30, 45, 100, time.Local),
			},
			{
				Title:       "Bits from the DPL",
				PublishedAt: time.Date(2024, 4, 3, 11, 12, 27, 100, time.Local),
			},
		},
	}

	return feedEntries, nil
}
