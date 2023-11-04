package querying

import "errors"

var (
	ErrPageNumberOutOfBounds error = errors.New("querying: invalid page index (out of bounds)")
	ErrOwnerNotFound         error = errors.New("querying: owner not found")
)
