// Copyright (c) VirtualTam
// SPDX-License-Identifier: MIT

package exporting

import (
	"fmt"
	"testing"
	"time"

	"github.com/jaswdr/faker"
	"github.com/virtualtam/opml-go"
	"github.com/virtualtam/sparklemuffin/internal/test/assert"
	"github.com/virtualtam/sparklemuffin/pkg/feed"
	"github.com/virtualtam/sparklemuffin/pkg/user"
)

func TestServiceExportAsOPMLDocument(t *testing.T) {
	fake := faker.New()

	testUser := user.User{
		UUID:        fake.UUID().V4(),
		DisplayName: "Test User",
	}

	now := time.Now().UTC()
	wantTitle := fmt.Sprintf("%s's feed subscriptions on SparkleMuffin", testUser.DisplayName)
	wantHead := opml.Head{
		Title:       wantTitle,
		DateCreated: now,
	}

	cases := []struct {
		tname                   string
		categoriesSubscriptions []CategorySubscriptions
		want                    opml.Document
	}{
		{
			tname: "no subscriptions",
			want: opml.Document{
				Version: opml.Version2,
				Head:    wantHead,
			},
		},
		{
			tname: "categorized subscriptions",
			categoriesSubscriptions: []CategorySubscriptions{
				{
					Category: feed.Category{
						Name: "Category 1",
					},
					SubscribedFeeds: []feed.Feed{
						{
							Title:   "Test Feed 1",
							FeedURL: "http://dev1.local/feed",
						},
						{
							Title:   "Test Feed 2",
							FeedURL: "http://dev2.local/feed",
						},
					},
				},
			},
			want: opml.Document{
				Version: opml.Version2,
				Head:    wantHead,
				Body: opml.Body{
					Outlines: []opml.Outline{
						{
							Text:  "Category 1",
							Title: "Category 1",
							Outlines: []opml.Outline{
								{
									Text:   "Test Feed 1",
									Title:  "Test Feed 1",
									Type:   opml.OutlineTypeSubscription,
									XmlUrl: "http://dev1.local/feed",
								},
								{
									Text:   "Test Feed 2",
									Title:  "Test Feed 2",
									Type:   opml.OutlineTypeSubscription,
									XmlUrl: "http://dev2.local/feed",
								},
							},
						},
					},
				},
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.tname, func(t *testing.T) {
			r := &fakeRepository{
				categoriesSubscriptions: tc.categoriesSubscriptions,
			}
			s := NewService(r)

			got, err := s.ExportAsOPMLDocument(testUser)

			if err != nil {
				t.Fatalf("want no error, got %q", err)
			}

			if got.Version != opml.Version2 {
				t.Errorf("want Version %q, got %q", opml.Version2, got.Version)
			}

			if got.Head.Title != tc.want.Head.Title {
				t.Errorf("want Head > Title %q, got %q", tc.want.Head.Title, got.Head.Title)
			}

			assert.TimeAlmostEquals(t, "DateCreated", got.Head.DateCreated, tc.want.Head.DateCreated, assert.TimeComparisonDelta)
			opml.AssertOutlinesEqual(t, got.Body.Outlines, tc.want.Body.Outlines)
		})
	}
}
