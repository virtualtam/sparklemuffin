// Copyright (c) VirtualTam
// SPDX-License-Identifier: MIT

package importing

import (
	"net/http"
	"testing"
	"time"

	"github.com/jaswdr/faker/v2"
	"github.com/virtualtam/opml-go"

	"github.com/virtualtam/sparklemuffin/internal/test/feedtest"
	"github.com/virtualtam/sparklemuffin/pkg/feed"
	"github.com/virtualtam/sparklemuffin/pkg/feed/fetching"
	"github.com/virtualtam/sparklemuffin/pkg/user"
)

func TestServiceImportFromOPMLDocument(t *testing.T) {
	fake := faker.New()

	testUser := user.User{
		UUID: fake.UUID().V4(),
	}

	now := time.Now().UTC()
	testFeed := feedtest.GenerateDummyFeed(t, now)
	transport := feedtest.NewRoundTripperFromFeed(t, testFeed)

	testHTTPClient := &http.Client{
		Transport: transport,
	}

	cases := []struct {
		tname string

		repositoryCategories    []feed.Category
		repositoryFeeds         []feed.Feed
		repositorySubscriptions []feed.Subscription

		outlines []opml.Outline

		wantStatus Status
	}{
		{
			tname: "empty document",
		},
		{
			tname: "new subscriptions, uncategorized",
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
			wantStatus: Status{
				Categories: StatusCount{
					Total:   1,
					Created: 1,
				},
				Feeds: StatusCount{
					Total:   2,
					Created: 2,
				},
				Subscriptions: StatusCount{
					Total:   2,
					Created: 2,
				},
			},
		},
		{
			tname: "new subscriptions, categorized, 2 categories",
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
			wantStatus: Status{
				Categories: StatusCount{
					Total:   2,
					Created: 2,
				},
				Feeds: StatusCount{
					Total:   2,
					Created: 2,
				},
				Subscriptions: StatusCount{
					Total:   2,
					Created: 2,
				},
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.tname, func(t *testing.T) {
			r := &feed.FakeRepository{
				Categories:    tc.repositoryCategories,
				Feeds:         tc.repositoryFeeds,
				Subscriptions: tc.repositorySubscriptions,
			}
			feedClient := fetching.NewClient(testHTTPClient, "sparklemuffin/test")

			feedService := feed.NewService(r, feedClient)
			s := NewService(feedService)

			document := &opml.Document{
				Body: opml.Body{
					Outlines: tc.outlines,
				},
			}

			status, err := s.ImportFromOPMLDocument(t.Context(), testUser.UUID, document)
			if err != nil {
				t.Fatalf("want no error, got %q", err)
			}

			if status.AdminSummary() != tc.wantStatus.AdminSummary() {
				t.Errorf("want Status Summary %q, got %q", tc.wantStatus.AdminSummary(), status.AdminSummary())
			}
		})
	}
}
