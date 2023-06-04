package bookmark

import "errors"

var (
	ErrNotFound                         error = errors.New("not found")
	ErrTagCurrentNameContainsWhitespace error = errors.New("current tag contains whitespace")
	ErrTagCurrentNameRequired           error = errors.New("current tag required")
	ErrTagNameContainsWhitespace        error = errors.New("tag contains whitespace")
	ErrTagNameRequired                  error = errors.New("tag required")
	ErrTagNewNameContainsWhitespace     error = errors.New("new tag contains whitespace")
	ErrTagNewNameEqualsCurrentName      error = errors.New("new tag is the same as current tag")
	ErrTagNewNameRequired               error = errors.New("new tag required")
	ErrTitleRequired                    error = errors.New("title required")
	ErrUIDInvalid                       error = errors.New("invalid UID")
	ErrUIDRequired                      error = errors.New("UID required")
	ErrURLAlreadyRegistered             error = errors.New("URL already registered")
	ErrURLInvalid                       error = errors.New("invalid URL")
	ErrURLRequired                      error = errors.New("URL required")
	ErrUserUUIDRequired                 error = errors.New("user UUID required")
)
