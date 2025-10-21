// Copyright (c) VirtualTam
// SPDX-License-Identifier: MIT

package importing

import "errors"

// OnConflictStrategy represents the strategy to apply when importing data
// that is already present in the repository.
type OnConflictStrategy string

const (
	// OnConflictOverwrite indicates existing bookmarks will be overwritten by imported bookmarks.
	OnConflictOverwrite OnConflictStrategy = "overwrite"

	// OnConflictKeepExisting indicates existing bookmarks will be left unchanged.
	OnConflictKeepExisting OnConflictStrategy = "keep"
)

var (
	ErrOnConflictStrategyInvalid = errors.New("invalid value for on-conflict strategy")
)
