// Copyright (c) VirtualTam
// SPDX-License-Identifier: MIT

package bookmark

import (
	"errors"
	"fmt"
)

var (
	ErrNotFound                    error = errors.New("not found")
	ErrTagNameContainsWhitespace   error = errors.New("tag name contains whitespace")
	ErrTagNameRequired             error = errors.New("tag name required")
	ErrTagNewNameEqualsCurrentName error = errors.New("new tag name is the same as current tag name")
	ErrTitleRequired               error = errors.New("title required")
	ErrUIDInvalid                  error = errors.New("invalid UID")
	ErrUIDRequired                 error = errors.New("UID required")
	ErrURLAlreadyRegistered        error = errors.New("URL already registered")
	ErrURLInvalid                  error = errors.New("invalid URL")
	ErrURLNoHost                   error = errors.New("URL has no host")
	ErrURLNoScheme                 error = errors.New("URL has no scheme")
	ErrURLRequired                 error = errors.New("URL required")
)

func newValidationError(field string, e error) error {
	return fmt.Errorf("%s: %w", field, e)
}
