// Copyright (c) VirtualTam
// SPDX-License-Identifier: MIT

package fetching

import "testing"

func TestLastModified(t *testing.T) {
	lastModifiedStr := "Sun, 18 Jun 2023 23:18:01 GMT"

	t.Run("round-trip", func(t *testing.T) {
		got := formatLastModified(parseLastModified(lastModifiedStr))

		if got != lastModifiedStr {
			t.Errorf("want %q, got %q", lastModifiedStr, got)
		}
	})
}
