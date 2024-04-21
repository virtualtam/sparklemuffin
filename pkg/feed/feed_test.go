// Copyright (c) VirtualTam
// SPDX-License-Identifier: MIT

package feed

import (
	"testing"
	"time"

	"github.com/virtualtam/sparklemuffin/internal/assert"
)

const (
	feedDateComparisonDelta time.Duration = 1 * time.Second
)

func assertFeedEquals(t *testing.T, got, want Feed) {
	t.Helper()

	if got.Slug != want.Slug {
		t.Errorf("want Slug %q, got %q", want.Slug, got.Slug)
	}
	if got.Title != want.Title {
		t.Errorf("want Title %q, got %q", want.Title, got.Title)
	}
	if got.FeedURL != want.FeedURL {
		t.Errorf("want FeedURL %q, got %q", want.FeedURL, got.FeedURL)
	}

	assert.TimeAlmostEquals(t, "CreatedAt", got.CreatedAt, want.CreatedAt, feedDateComparisonDelta)
	assert.TimeAlmostEquals(t, "UpdatedAt", got.UpdatedAt, want.UpdatedAt, feedDateComparisonDelta)
	assert.TimeAlmostEquals(t, "FetchedAt", got.FetchedAt, want.FetchedAt, feedDateComparisonDelta)
}
