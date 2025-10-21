// Copyright (c) VirtualTam
// SPDX-License-Identifier: MIT

package pgfeed_test

import (
	"net/http"
	"testing"
	"time"

	"github.com/jaswdr/faker"
	"github.com/virtualtam/opml-go"

	"github.com/virtualtam/sparklemuffin/internal/repository/postgresql/pgbase"
	"github.com/virtualtam/sparklemuffin/internal/repository/postgresql/pgfeed"
	"github.com/virtualtam/sparklemuffin/internal/repository/postgresql/pguser"
	"github.com/virtualtam/sparklemuffin/internal/test/feedtest"
	"github.com/virtualtam/sparklemuffin/pkg/feed"
	"github.com/virtualtam/sparklemuffin/pkg/feed/fetching"
	"github.com/virtualtam/sparklemuffin/pkg/feed/importing"
	"github.com/virtualtam/sparklemuffin/pkg/user"
)

func TestImportingService(t *testing.T) {
	pool := pgbase.CreateAndMigrateTestDatabase(t)

	now := time.Now().UTC()
	atomFeed := feedtest.GenerateDummyFeed(t, now)

	transport := feedtest.NewRoundTripper(t, atomFeed)

	testHTTPClient := &http.Client{
		Transport: transport,
	}
	feedClient := fetching.NewClient(testHTTPClient, "sparklemuffin/test")

	r := pgfeed.NewRepository(pool)
	s := feed.NewService(r, feedClient)
	is := importing.NewService(s)

	ur := pguser.NewRepository(pool)
	us := user.NewService(ur)

	fake := faker.New()

	u := pgbase.GenerateFakeUser(t, &fake)

	if err := us.Add(u); err != nil {
		t.Fatalf("failed to create user: %q", err)
	}

	testUser, err := us.ByNickName(u.NickName)
	if err != nil {
		t.Fatalf("failed to retrieve user: %q", err)
	}

	t.Run("ImportFromOPMLDocument", func(t *testing.T) {
		document := &opml.Document{
			Version: opml.Version2,
			Body: opml.Body{
				Outlines: []opml.Outline{
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
			},
		}

		wantStatus := importing.Status{
			Categories: importing.StatusCount{
				Total:   2,
				Created: 2,
			},
			Feeds: importing.StatusCount{
				Total:   2,
				Created: 2,
			},
			Subscriptions: importing.StatusCount{
				Total:   2,
				Created: 2,
			},
		}

		status, err := is.ImportFromOPMLDocument(testUser.UUID, document)
		if err != nil {
			t.Fatalf("failed to import OPML document: %q", err)
		}

		if status.AdminSummary() != wantStatus.AdminSummary() {
			t.Errorf("want Status Summary %q, got %q", wantStatus.AdminSummary(), status.AdminSummary())
		}
	})
}
