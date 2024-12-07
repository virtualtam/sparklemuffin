// Copyright (c) VirtualTam
// SPDX-License-Identifier: MIT

package querying

import (
	"testing"

	"github.com/virtualtam/sparklemuffin/pkg/feed"
)

type SubscribedFeedsByCategory struct {
	feed.Category

	Unread uint

	SubscribedFeeds []SubscribedFeed
}

type SubscribedFeed struct {
	feed.Feed

	Alias  string
	Unread uint
}

type SubscribedFeedEntry struct {
	feed.Entry

	FeedTitle string
	Read      bool
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

type Subscription struct {
	UUID         string
	CategoryUUID string
	Alias        string

	FeedTitle       string
	FeedDescription string
}

type SubscriptionsByCategory struct {
	feed.Category

	Subscriptions []Subscription
}
