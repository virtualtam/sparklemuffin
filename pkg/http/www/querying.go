package www

import (
	"strconv"
)

func getPageNumber(pageNumberParam string) (uint, error) {
	if pageNumberParam == "" {
		return 1, nil
	}

	pageNumber64, err := strconv.ParseUint(pageNumberParam, 10, 64)
	if err != nil {
		return 0, err
	}

	return uint(pageNumber64), nil
}
