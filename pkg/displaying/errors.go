package displaying

import "errors"

var (
	ErrPageNumberOutOfBounds error = errors.New("page: invalid index (out of bounds)")
)
