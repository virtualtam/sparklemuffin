// Copyright (c) VirtualTam
// SPDX-License-Identifier: MIT

package feed

import "testing"

func assertFeedEquals(t *testing.T, got, want Feed) {
	t.Helper()

	if got.Title != want.Title {
		t.Errorf("want title %q, got %q", want.Title, got.Title)
	}
	if got.FeedURL != want.FeedURL {
		t.Errorf("want FeedURL %q, got %q", want.FeedURL, got.FeedURL)
	}
}
