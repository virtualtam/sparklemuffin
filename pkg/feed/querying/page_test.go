// Copyright (c) VirtualTam
// SPDX-License-Identifier: MIT

package querying

import (
	"fmt"
	"testing"

	"github.com/virtualtam/sparklemuffin/internal/assert"
)

func TestNewPage(t *testing.T) {
	cases := []struct {
		tname      string
		number     uint
		totalPages uint
		want       FeedPage
	}{
		{
			tname:      "page 1 of 1",
			number:     1,
			totalPages: 1,
			want: FeedPage{
				PageNumber:         1,
				PreviousPageNumber: 1,
				NextPageNumber:     1,
				TotalPages:         1,
				Offset:             1,
			},
		},
		{
			tname:      "page 1 of 8",
			number:     1,
			totalPages: 8,
			want: FeedPage{
				PageNumber:         1,
				PreviousPageNumber: 1,
				NextPageNumber:     2,
				TotalPages:         8,
				Offset:             1,
			},
		},
		{
			tname:      "page 7 of 8",
			number:     7,
			totalPages: 8,
			want: FeedPage{
				PageNumber:         7,
				PreviousPageNumber: 6,
				NextPageNumber:     8,
				TotalPages:         8,
				Offset:             6*entriesPerPage + 1,
			},
		},
		{
			tname:      "page 8 of 8",
			number:     8,
			totalPages: 8,
			want: FeedPage{
				PageNumber:         8,
				PreviousPageNumber: 7,
				NextPageNumber:     8,
				TotalPages:         8,
				Offset:             7*entriesPerPage + 1,
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.tname, func(t *testing.T) {
			got := NewFeedPage(tc.number, tc.totalPages, []SubscriptionCategory{}, []SubscriptionEntry{})
			assertPagesEqual(t, got, tc.want)
		})
	}
}

func assertPagesEqual(t *testing.T, got, want FeedPage) {
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

	if got.Unread != want.Unread {
		t.Errorf("want Unread %d, got %d", want.Unread, got.Unread)
	}

	assertCategoriesEqual(t, got.Categories, want.Categories)
	assertSubscriptionEntriesEqual(t, got.Entries, want.Entries)
}

func assertCategoriesEqual(t *testing.T, got, want []SubscriptionCategory) {
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

		assertSubscribedFeedsEqual(t, i, gotCategory.SubscribedFeeds, wantCategory.SubscribedFeeds)
	}
}

func assertSubscriptionEntriesEqual(t *testing.T, got []SubscriptionEntry, want []SubscriptionEntry) {
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

		assert.TimeEquals(t, fmt.Sprintf("Entry %d PublishedAt", i), gotEntry.PublishedAt, wantEntry.PublishedAt)
		assert.TimeEquals(t, fmt.Sprintf("Entry %d UpdatedAt", i), gotEntry.UpdatedAt, wantEntry.UpdatedAt)

		// querying.Entry fields
		if gotEntry.Read != wantEntry.Read {
			t.Errorf("want Entry %d Read %t, got %t", i, wantEntry.Read, gotEntry.Read)
		}
	}
}

func assertSubscribedFeedsEqual(t *testing.T, categoryIndex int, got []SubscribedFeed, want []SubscribedFeed) {
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
		if gotSubscribedFeed.URL != wantSubscribedFeed.URL {
			t.Errorf("want URL %q, got %q", wantSubscribedFeed.URL, gotSubscribedFeed.URL)
		}

		assert.TimeEquals(t, "CreatedAt", gotSubscribedFeed.CreatedAt, wantSubscribedFeed.CreatedAt)
		assert.TimeEquals(t, "UpdatedAt", gotSubscribedFeed.UpdatedAt, wantSubscribedFeed.UpdatedAt)
		assert.TimeEquals(t, "FetchedAt", gotSubscribedFeed.FetchedAt, wantSubscribedFeed.FetchedAt)

		// querying.SubscribedFeed fields
		if gotSubscribedFeed.Unread != wantSubscribedFeed.Unread {
			t.Errorf("want Category %d Unread %d, got %d", i, wantSubscribedFeed.Unread, gotSubscribedFeed.Unread)
		}
	}
}
