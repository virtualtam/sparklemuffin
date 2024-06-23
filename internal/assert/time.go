// Copyright (c) VirtualTam
// SPDX-License-Identifier: MIT

package assert

import (
	"testing"
	"time"
)

const (
	// TimeComparisonDelta provides a sensible default when comparing datetimes in tests.
	TimeComparisonDelta time.Duration = 1 * time.Second
)

// TimeAlmostEquals checks whether two dates are almost equal.
//
// This helper should be used in tests covering data insertion and retrieval.
func TimeAlmostEquals(t *testing.T, fieldName string, got, want time.Time, delta time.Duration) {
	t.Helper()

	if got.Sub(want).Abs() > delta {
		t.Errorf("want %s %q, got %q", fieldName, want.String(), got.String())
	}
}

// TimeEquals checks whether two dates are equal.
//
// This helper should be used in tests covering data retrieval.
func TimeEquals(t *testing.T, fieldName string, got, want time.Time) {
	t.Helper()

	if !got.Equal(want) {
		t.Errorf("want %s %q, got %q", fieldName, want.String(), got.String())
	}
}
