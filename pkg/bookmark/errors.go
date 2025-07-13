// Copyright (c) VirtualTam
// SPDX-License-Identifier: MIT

package bookmark

import (
	"errors"
	"fmt"
)

var (
	ErrNotFound                    = errors.New("bookmark: not found")
	ErrTagNameContainsWhitespace   = errors.New("bookmark: tag name contains whitespace")
	ErrTagNameRequired             = errors.New("bookmark: tag name required")
	ErrTagNewNameEqualsCurrentName = errors.New("bookmark: new tag name is the same as current tag name")
	ErrTitleRequired               = errors.New("bookmark: title required")
	ErrUIDInvalid                  = errors.New("bookmark: invalid UID")
	ErrUIDRequired                 = errors.New("bookmark: UID required")
	ErrURLAlreadyRegistered        = errors.New("bookmark: URL already registered")
	ErrURLInvalid                  = errors.New("bookmark: invalid URL")
	ErrURLNoHost                   = errors.New("bookmark: URL has no host")
	ErrURLNoScheme                 = errors.New("bookmark: URL has no scheme")
	ErrURLRequired                 = errors.New("bookmark: URL required")
)

func newValidationError(field string, e error) error {
	return fmt.Errorf("%s: %w", field, e)
}
