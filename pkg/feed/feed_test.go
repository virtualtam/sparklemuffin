// Copyright (c) VirtualTam
// SPDX-License-Identifier: MIT

package feed

import "testing"

func assertFeedEquals(t *testing.T, got, want Feed) {
	t.Helper()

	if got.Title != want.Title {
		t.Errorf("want title %q, got %q", want.Title, got.Title)
	}
	if got.URL != want.URL {
		t.Errorf("want URL %q, got %q", want.URL, got.URL)
	}
}
