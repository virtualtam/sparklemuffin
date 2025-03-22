// Copyright (c) VirtualTam
// SPDX-License-Identifier: MIT

package querying

import "errors"

var (
	ErrPageNumberOutOfBounds = errors.New("querying: invalid page index (out of bounds)")
)
