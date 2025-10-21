// Copyright (c) VirtualTam
// SPDX-License-Identifier: MIT

package feed

import (
	"testing"
	"time"

	"github.com/virtualtam/sparklemuffin/internal/test/assert"
)

const (
	categoryDateComparisonDelta = 1 * time.Second
)

func assertCategoryEquals(t *testing.T, got, want Category) {
	t.Helper()

	if want.UUID != "-" {
		// Skip UUID checks for newly created entries
		if got.UUID != want.UUID {
			t.Errorf("want UUID %q, got %q", want.UUID, got.UUID)
		}
	}

	if got.UserUUID != want.UserUUID {
		t.Errorf("want UserUUID %q, got %q", want.UserUUID, got.UserUUID)
	}
	if got.Name != want.Name {
		t.Errorf("want Name %q, got %q", want.Name, got.Name)
	}
	if got.Slug != want.Slug {
		t.Errorf("want Slug %q, got %q", want.Slug, got.Slug)
	}

	assert.TimeAlmostEquals(t, "CreatedAt", got.CreatedAt, want.CreatedAt, categoryDateComparisonDelta)
	assert.TimeAlmostEquals(t, "UpdatedAt", got.UpdatedAt, want.UpdatedAt, categoryDateComparisonDelta)
}
