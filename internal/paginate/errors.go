// Copyright (c) VirtualTam
// SPDX-License-Identifier: MIT

package paginate

import "errors"

var (
	ErrPageNumberOutOfBounds = errors.New("paginate: invalid page index (out of bounds)")
)
