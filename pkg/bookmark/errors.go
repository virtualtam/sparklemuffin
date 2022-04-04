package bookmark

import "errors"

var (
	ErrNotFound             error = errors.New("not found")
	ErrTitleRequired        error = errors.New("title required")
	ErrUIDInvalid           error = errors.New("invalid UID")
	ErrUIDRequired          error = errors.New("UID required")
	ErrURLAlreadyRegistered error = errors.New("URL already registered")
	ErrURLInvalid           error = errors.New("invalid URL")
	ErrURLRequired          error = errors.New("URL required")
	ErrUserUUIDRequired     error = errors.New("user UUID required")
)
