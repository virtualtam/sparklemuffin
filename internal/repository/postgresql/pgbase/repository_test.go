// Copyright (c) VirtualTam
// SPDX-License-Identifier: MIT

package pgbase

import (
	"errors"
	"testing"

	"github.com/golang-migrate/migrate/v4"
)

func TestMigrate(t *testing.T) {
	_, db := createTestDatabase(t)
	migrater := getDatabaseMigrater(t, db)

	t.Run("up", func(t *testing.T) {
		if err := migrater.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
			t.Fatalf("failed to apply database migrations (up): %q", err)
		}
	})

	t.Run("down", func(t *testing.T) {
		if err := migrater.Down(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
			t.Fatalf("failed to apply database migrations (down): %q", err)
		}
	})
}
