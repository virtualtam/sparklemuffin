// Copyright (c) VirtualTam
// SPDX-License-Identifier: MIT

package feed

import (
	"testing"

	"github.com/virtualtam/sparklemuffin/internal/assert"
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

	if got.ETag != want.ETag {
		t.Errorf("want ETag %q, got %q", want.ETag, got.ETag)
	}

	assert.TimeAlmostEquals(t, "LastModified", got.LastModified, want.LastModified, assert.TimeComparisonDelta)
	assert.TimeAlmostEquals(t, "CreatedAt", got.CreatedAt, want.CreatedAt, assert.TimeComparisonDelta)
	assert.TimeAlmostEquals(t, "UpdatedAt", got.UpdatedAt, want.UpdatedAt, assert.TimeComparisonDelta)
	assert.TimeAlmostEquals(t, "FetchedAt", got.FetchedAt, want.FetchedAt, assert.TimeComparisonDelta)
}
