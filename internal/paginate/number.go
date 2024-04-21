// Copyright (c) VirtualTam
// SPDX-License-Identifier: MIT

package paginate

import (
	"net/url"
	"strconv"
)

const (
	pageNumberQueryParam string = "page"
)

// GetPageNumber retrieves or sets the page number for paginated views.
func GetPageNumber(urlQuery url.Values) (uint, string, error) {
	pageNumberValue := urlQuery.Get(pageNumberQueryParam)

	if pageNumberValue == "" {
		return 1, pageNumberValue, nil
	}

	pageNumber64, err := strconv.ParseUint(pageNumberValue, 10, 64)
	if err != nil {
		return 0, pageNumberValue, err
	}

	return uint(pageNumber64), pageNumberValue, nil
}
