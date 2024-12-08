// Copyright (c) VirtualTam
// SPDX-License-Identifier: MIT

package querying

import (
	"fmt"
	"testing"

	"github.com/virtualtam/sparklemuffin/internal/test/assert"
	"github.com/virtualtam/sparklemuffin/pkg/feed"
)

func AssertPageEquals(t *testing.T, got, want FeedPage) {
	t.Helper()

	if got.PageNumber != want.PageNumber {
		t.Errorf("want PageNumber %d, got %d", want.PageNumber, got.PageNumber)
	}
	if got.PreviousPageNumber != want.PreviousPageNumber {
		t.Errorf("want PreviousPageNumber %d, got %d", want.PreviousPageNumber, got.PreviousPageNumber)
	}
	if got.NextPageNumber != want.NextPageNumber {
		t.Errorf("want NextPageNumber %d, got %d", want.NextPageNumber, got.NextPageNumber)
	}
	if got.TotalPages != want.TotalPages {
		t.Errorf("want TotalPages %d, got %d", want.TotalPages, got.TotalPages)
	}
	if got.Offset != want.Offset {
		t.Errorf("want Offset %d, got %d", want.Offset, got.Offset)
	}

	if got.PageTitle != want.PageTitle {
		t.Errorf("want Header %q, got %q", want.PageTitle, got.PageTitle)
	}
	if got.Description != want.Description {
		t.Errorf("want Description %q, got %q", want.Description, got.Description)
	}
	if got.Unread != want.Unread {
		t.Errorf("want Unread %d, got %d", want.Unread, got.Unread)
	}

	AssertCategoriesEqual(t, got.Categories, want.Categories)
	AssertSubscriptionEntriesEqual(t, got.Entries, want.Entries)
}

func AssertCategoriesEqual(t *testing.T, got, want []SubscribedFeedsByCategory) {
	t.Helper()

	if len(got) != len(want) {
		t.Fatalf("want %d Categories, got %d", len(want), len(got))
	}

	for i, wantCategory := range want {
		gotCategory := got[i]

		// Embedded feed.Category fields
		if gotCategory.Name != wantCategory.Name {
			t.Errorf("want Category %d Name %q, got %q", i, wantCategory.Name, gotCategory.Name)
		}
		if gotCategory.Slug != wantCategory.Slug {
			t.Errorf("want Category %d Slug %q, got %q", i, wantCategory.Slug, gotCategory.Slug)
		}

		assert.TimeEquals(t, fmt.Sprintf("Category %d CreatedAt", i), gotCategory.CreatedAt, wantCategory.CreatedAt)
		assert.TimeEquals(t, fmt.Sprintf("Category %d UpdatedAt", i), gotCategory.UpdatedAt, wantCategory.UpdatedAt)

		// querying.Category fields
		if gotCategory.Unread != wantCategory.Unread {
			t.Errorf("want Category %d Unread %d, got %d", i, wantCategory.Unread, gotCategory.Unread)
		}

		AssertSubscribedFeedsEqual(t, i, gotCategory.SubscribedFeeds, wantCategory.SubscribedFeeds)
	}
}

func AssertSubscriptionEntriesEqual(t *testing.T, got []SubscribedFeedEntry, want []SubscribedFeedEntry) {
	t.Helper()

	if len(got) != len(want) {
		t.Fatalf("want %d Entries, got %d", len(want), len(got))
	}

	for i, wantEntry := range want {
		gotEntry := got[i]

		// Embedded feed.Entry fields
		if gotEntry.FeedUUID != wantEntry.FeedUUID {
			t.Errorf("want Entry %d FeedUUID %q, got %q", i, wantEntry.FeedUUID, gotEntry.FeedUUID)
		}
		if gotEntry.Title != wantEntry.Title {
			t.Errorf("want Entry %d Title %q, got %q", i, wantEntry.Title, gotEntry.Title)
		}
		if gotEntry.URL != wantEntry.URL {
			t.Errorf("want Entry %d URL %q, got %q", i, wantEntry.URL, gotEntry.URL)
		}

		assert.TimeAlmostEquals(t, fmt.Sprintf("Entry %d PublishedAt", i), gotEntry.PublishedAt, wantEntry.PublishedAt, assert.TimeComparisonDelta)
		assert.TimeAlmostEquals(t, fmt.Sprintf("Entry %d UpdatedAt", i), gotEntry.UpdatedAt, wantEntry.UpdatedAt, assert.TimeComparisonDelta)

		// querying.SubscribedFeedEntry fields
		if gotEntry.SubscriptionAlias != wantEntry.SubscriptionAlias {
			t.Errorf("want Entry %d SubscriptionAlias %q, got %q", i, wantEntry.SubscriptionAlias, gotEntry.SubscriptionAlias)
		}
		if gotEntry.FeedTitle != wantEntry.FeedTitle {
			t.Errorf("want Entry %d FeedTitle %q, got %q", i, wantEntry.FeedTitle, gotEntry.FeedTitle)
		}
		if gotEntry.Read != wantEntry.Read {
			t.Errorf("want Entry %d Read %t, got %t", i, wantEntry.Read, gotEntry.Read)
		}
	}
}

func AssertSubscribedFeedsEqual(t *testing.T, categoryIndex int, got []SubscribedFeed, want []SubscribedFeed) {
	t.Helper()

	if len(got) != len(want) {
		t.Fatalf("want Category %d %d Feeds, got %d", categoryIndex, len(want), len(got))
	}

	for i, wantSubscribedFeed := range want {
		gotSubscribedFeed := got[i]

		// Embedded feed.Feed fields
		if gotSubscribedFeed.Slug != wantSubscribedFeed.Slug {
			t.Errorf("want Slug %q, got %q", wantSubscribedFeed.Slug, gotSubscribedFeed.Slug)
		}
		if gotSubscribedFeed.Title != wantSubscribedFeed.Title {
			t.Errorf("want Title %q, got %q", wantSubscribedFeed.Title, gotSubscribedFeed.Title)
		}
		if gotSubscribedFeed.FeedURL != wantSubscribedFeed.FeedURL {
			t.Errorf("want FeedURL %q, got %q", wantSubscribedFeed.FeedURL, gotSubscribedFeed.FeedURL)
		}

		assert.TimeAlmostEquals(t, "CreatedAt", gotSubscribedFeed.CreatedAt, wantSubscribedFeed.CreatedAt, assert.TimeComparisonDelta)
		assert.TimeAlmostEquals(t, "UpdatedAt", gotSubscribedFeed.UpdatedAt, wantSubscribedFeed.UpdatedAt, assert.TimeComparisonDelta)
		assert.TimeAlmostEquals(t, "FetchedAt", gotSubscribedFeed.FetchedAt, wantSubscribedFeed.FetchedAt, assert.TimeComparisonDelta)

		// querying.SubscribedFeed fields
		if gotSubscribedFeed.Unread != wantSubscribedFeed.Unread {
			t.Errorf("want Category %d Unread %d, got %d", i, wantSubscribedFeed.Unread, gotSubscribedFeed.Unread)
		}
	}
}

func AssertSubscribedFeedEntriesEqual(t *testing.T, got []SubscribedFeedEntry, want []SubscribedFeedEntry) {
	t.Helper()

	if len(got) != len(want) {
		t.Fatalf("want %d Entries, got %d", len(want), len(got))
	}

	for i, wantSubscribedFeedEntry := range want {
		feed.AssertEntryEquals(t, i, got[i].Entry, wantSubscribedFeedEntry.Entry)

		if wantSubscribedFeedEntry.FeedTitle != got[i].FeedTitle {
			t.Errorf("want Entry %d FeedTitle %q, got %q", i, wantSubscribedFeedEntry.FeedTitle, got[i].FeedTitle)
		}
		if wantSubscribedFeedEntry.Read != got[i].Read {
			t.Errorf("want Entry %d Read %t, got %t", i, wantSubscribedFeedEntry.Read, got[i].Read)
		}
	}
}
