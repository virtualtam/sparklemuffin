// Copyright (c) VirtualTam
// SPDX-License-Identifier: MIT

package feed

import (
	"testing"
	"time"

	"github.com/virtualtam/sparklemuffin/internal/assert"
)

const (
	entryDateComparisonDelta time.Duration = 1 * time.Second
)

func assertEntriesEqual(t *testing.T, gotEntries []Entry, wantEntries []Entry) {
	t.Helper()

	if len(gotEntries) != len(wantEntries) {
		t.Fatalf("want %d entries, got %d", len(wantEntries), len(gotEntries))
	}

	for i, wantEntry := range wantEntries {
		gotEntry := gotEntries[i]

		if gotEntry.FeedUUID != wantEntry.FeedUUID {
			t.Errorf("want FeedUUID %q, got %q", wantEntry.FeedUUID, gotEntry.FeedUUID)
		}
		if gotEntry.Title != wantEntry.Title {
			t.Errorf("want Title %q, got %q", wantEntry.Title, gotEntry.Title)
		}
		if gotEntry.URL != wantEntry.URL {
			t.Errorf("want URL %q, got %q", wantEntry.URL, gotEntry.URL)
		}

		assert.TimeAlmostEquals(t, "PublishedAt", gotEntry.PublishedAt, wantEntry.PublishedAt, entryDateComparisonDelta)
		assert.TimeAlmostEquals(t, "UpdatedAt", gotEntry.UpdatedAt, wantEntry.UpdatedAt, entryDateComparisonDelta)
	}
}
