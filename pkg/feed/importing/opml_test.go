// Copyright (c) VirtualTam
// SPDX-License-Identifier: MIT

package importing

import (
	"fmt"
	"testing"

	"github.com/virtualtam/opml-go"
)

func TestOpmlToCategoriesFeeds(t *testing.T) {
	cases := []struct {
		tname    string
		outlines []opml.Outline
		want     map[string][]string
	}{
		{
			tname: "no outline",
		},
		{
			tname: "uncategorized",
			outlines: []opml.Outline{
				{
					Text:    "Outline 1",
					Title:   "Outline 1",
					Type:    opml.OutlineTypeSubscription,
					HtmlUrl: "http://dev1.local",
					XmlUrl:  "http://dev1.local/feed",
				},
				{
					Text:    "Outline 2",
					Title:   "Outline 2",
					Type:    opml.OutlineTypeSubscription,
					HtmlUrl: "http://dev2.local",
					XmlUrl:  "http://dev2.local/feed",
				},
			},
			want: map[string][]string{
				defaultCategoryName: {
					"http://dev1.local/feed",
					"http://dev2.local/feed",
				},
			},
		},
		{
			tname: "categorized, 1 category",
			outlines: []opml.Outline{
				{
					Text:  "Category 1",
					Title: "Category 1",
					Outlines: []opml.Outline{
						{
							Text:    "Outline 1",
							Title:   "Outline 1",
							Type:    opml.OutlineTypeSubscription,
							HtmlUrl: "http://dev1.local",
							XmlUrl:  "http://dev1.local/feed",
						},
						{
							Text:    "Outline 2",
							Title:   "Outline 2",
							Type:    opml.OutlineTypeSubscription,
							HtmlUrl: "http://dev2.local",
							XmlUrl:  "http://dev2.local/feed",
						},
					},
				},
			},
			want: map[string][]string{
				"Category 1": {
					"http://dev1.local/feed",
					"http://dev2.local/feed",
				},
			},
		},
		{
			tname: "categorized, 2 categories",
			outlines: []opml.Outline{
				{
					Text:  "Category 1",
					Title: "Category 1",
					Outlines: []opml.Outline{
						{
							Text:    "Outline 1",
							Title:   "Outline 1",
							Type:    opml.OutlineTypeSubscription,
							HtmlUrl: "http://dev1.local",
							XmlUrl:  "http://dev1.local/feed",
						},
					},
				},
				{
					Text:  "Category 2",
					Title: "Category 2",
					Outlines: []opml.Outline{
						{
							Text:    "Outline 2",
							Title:   "Outline 2",
							Type:    opml.OutlineTypeSubscription,
							HtmlUrl: "http://dev2.local",
							XmlUrl:  "http://dev2.local/feed",
						},
					},
				},
			},
			want: map[string][]string{
				"Category 1": {
					"http://dev1.local/feed",
				},
				"Category 2": {
					"http://dev2.local/feed",
				},
			},
		},
		{
			tname: "categorized and uncategorized",
			outlines: []opml.Outline{
				{
					Text:  "Category 1",
					Title: "Category 1",
					Outlines: []opml.Outline{
						{
							Text:    "Outline 1",
							Title:   "Outline 1",
							Type:    opml.OutlineTypeSubscription,
							HtmlUrl: "http://dev1.local",
							XmlUrl:  "http://dev1.local/feed",
						},
						{
							Text:    "Outline 2",
							Title:   "Outline 2",
							Type:    opml.OutlineTypeSubscription,
							HtmlUrl: "http://dev2.local",
							XmlUrl:  "http://dev2.local/feed",
						},
					},
				},
				{
					Text:    "Outline 3",
					Title:   "Outline 3",
					Type:    opml.OutlineTypeSubscription,
					HtmlUrl: "http://dev3.local",
					XmlUrl:  "http://dev3.local/feed",
				},
			},
			want: map[string][]string{
				"Category 1": {
					"http://dev1.local/feed",
					"http://dev2.local/feed",
				},
				"Default": {
					"http://dev3.local/feed",
				},
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.tname, func(t *testing.T) {
			got := opmlToCategoriesFeeds(tc.outlines)

			if fmt.Sprint(got) != fmt.Sprint(tc.want) {
				t.Errorf("want %q, got %q", tc.want, got)
			}
		})
	}
}
