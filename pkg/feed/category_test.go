// Copyright (c) VirtualTam
// SPDX-License-Identifier: MIT

package feed

import (
	"testing"
	"time"

	"github.com/virtualtam/sparklemuffin/internal/assert"
)

const (
	categoryDateComparisonDelta time.Duration = 1 * time.Second
)

func assertCategoriesEqual(t *testing.T, got, want Category) {
	t.Helper()

	if got.Name != want.Name {
		t.Errorf("want Name %q, got %q", want.Name, got.Name)
	}
	if got.Slug != want.Slug {
		t.Errorf("want Slug %q, got %q", want.Slug, got.Slug)
	}

	assert.DatesAlmostEqual(t, "CreatedAt", got.CreatedAt, want.CreatedAt, categoryDateComparisonDelta)
	assert.DatesAlmostEqual(t, "UpdatedAt", got.UpdatedAt, want.UpdatedAt, categoryDateComparisonDelta)
}
