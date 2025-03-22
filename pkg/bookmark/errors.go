// Copyright (c) VirtualTam
// SPDX-License-Identifier: MIT

package bookmark

import (
	"errors"
	"fmt"
)

var (
	ErrNotFound                    = errors.New("not found")
	ErrTagNameContainsWhitespace   = errors.New("tag name contains whitespace")
	ErrTagNameRequired             = errors.New("tag name required")
	ErrTagNewNameEqualsCurrentName = errors.New("new tag name is the same as current tag name")
	ErrTitleRequired               = errors.New("title required")
	ErrUIDInvalid                  = errors.New("invalid UID")
	ErrUIDRequired                 = errors.New("UID required")
	ErrURLAlreadyRegistered        = errors.New("URL already registered")
	ErrURLInvalid                  = errors.New("invalid URL")
	ErrURLNoHost                   = errors.New("URL has no host")
	ErrURLNoScheme                 = errors.New("URL has no scheme")
	ErrURLRequired                 = errors.New("URL required")
)

func newValidationError(field string, e error) error {
	return fmt.Errorf("%s: %w", field, e)
}
