// Copyright (c) VirtualTam
// SPDX-License-Identifier: MIT

package assert

import (
	"testing"
	"time"
)

// DatesAlmostEquals checks whether two dates are almost equal.
func DatesAlmostEqual(t *testing.T, fieldName string, got, want time.Time, delta time.Duration) {
	t.Helper()

	if got.Sub(want).Abs() > delta {
		t.Errorf("want %s %q, got %q", fieldName, want.String(), got.String())
	}
}
