// Copyright (c) VirtualTam
// SPDX-License-Identifier: MIT

package importing

import "errors"

// OnConflictStrategy represents the strategy to apply when importing data
// that is already present in the repository.
type OnConflictStrategy string

const (
	// Overwrite existing data with imported data.
	OnConflictOverwrite OnConflictStrategy = "overwrite"

	// Keep existing data and ignore new data.
	OnConflictKeepExisting OnConflictStrategy = "keep"
)

var (
	ErrOnConflictStrategyInvalid = errors.New("invalid value for on-conflict strategy")
)
