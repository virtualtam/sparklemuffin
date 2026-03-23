// Copyright VirtualTam 2022, 2026
// SPDX-License-Identifier: MIT

package paginate

import "errors"

var (
	ErrPageNumberOutOfBounds = errors.New("paginate: invalid page index (out of bounds)")
)
