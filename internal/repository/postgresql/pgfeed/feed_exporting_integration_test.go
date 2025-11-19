// Copyright (c) VirtualTam
// SPDX-License-Identifier: MIT

package pgfeed_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/jaswdr/faker/v2"
	"github.com/virtualtam/opml-go"

	"github.com/virtualtam/sparklemuffin/internal/repository/postgresql/pgbase"
	"github.com/virtualtam/sparklemuffin/internal/repository/postgresql/pgfeed"
	"github.com/virtualtam/sparklemuffin/internal/repository/postgresql/pguser"
	"github.com/virtualtam/sparklemuffin/internal/test/assert"
	"github.com/virtualtam/sparklemuffin/pkg/feed/exporting"
	"github.com/virtualtam/sparklemuffin/pkg/user"
)

func TestFeedExportingService(t *testing.T) {
	pool := pgbase.CreateAndMigrateTestDatabase(t)

	r := pgfeed.NewRepository(pool)
	es := exporting.NewService(r)

	ur := pguser.NewRepository(pool)
	us := user.NewService(ur)

	fake := faker.New()

	u := pgbase.GenerateFakeUser(t, &fake)

	if err := us.Add(t.Context(), u); err != nil {
		t.Fatalf("failed to create user: %q", err)
	}

	testUser, err := us.ByNickName(t.Context(), u.NickName)
	if err != nil {
		t.Fatalf("failed to retrieve user: %q", err)
	}

	now := time.Now().UTC()
	fakeData := generateFakeData(t, &fake, now, testUser)
	fakeData.insert(t, r)

	t.Run("ExportAsOPMLDocument", func(t *testing.T) {
		wantDocument := opml.Document{
			Version: opml.Version2,
			Head: opml.Head{
				Title:       fmt.Sprintf("%s's feed subscriptions on SparkleMuffin", testUser.DisplayName),
				DateCreated: now,
			},
			Body: opml.Body{
				Outlines: []opml.Outline{
					{
						Text:  fakeData.categories[0].Name,
						Title: fakeData.categories[0].Name,
						Outlines: []opml.Outline{
							{
								Text:   fakeData.feeds[0].Title,
								Title:  fakeData.feeds[0].Title,
								Type:   opml.OutlineTypeSubscription,
								XmlUrl: fakeData.feeds[0].FeedURL,
							},
							{
								Text:   fakeData.feeds[1].Title,
								Title:  fakeData.feeds[1].Title,
								Type:   opml.OutlineTypeSubscription,
								XmlUrl: fakeData.feeds[1].FeedURL,
							},
						},
					},
				},
			},
		}

		got, err := es.ExportAsOPMLDocument(t.Context(), testUser)
		if err != nil {
			t.Fatalf("want no error, got %q", err)
		}

		if got.Version != wantDocument.Version {
			t.Errorf("want Version %q, got %q", opml.Version2, got.Version)
		}

		if got.Head.Title != wantDocument.Head.Title {
			t.Errorf("want Head > Title %q, got %q", wantDocument.Head.Title, got.Head.Title)
		}

		assert.TimeAlmostEquals(t, "DateCreated", got.Head.DateCreated, wantDocument.Head.DateCreated, assert.TimeComparisonDelta)
		opml.AssertOutlinesEqual(t, got.Body.Outlines, wantDocument.Body.Outlines)
	})
}
