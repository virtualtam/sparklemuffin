// Copyright (c) VirtualTam
// SPDX-License-Identifier: MIT

package paginate

import "math"

// PageCount returns the total number of pages for a given number of items and items per page.
func PageCount(itemCount, itemsPerPage uint) uint {
	if itemCount == 0 {
		return 1
	}

	return uint(math.Ceil(float64(itemCount) / float64(itemsPerPage)))
}
