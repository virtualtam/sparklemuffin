// Copyright VirtualTam 2022, 2026
// SPDX-License-Identifier: MIT

package taxonomy

import (
	"errors"
	"fmt"
)

var (
	ErrNotFound                  = errors.New("taxonomy: not found")
	ErrTagAlreadyRegistered      = errors.New("taxonomy: tag already registered")
	ErrTagNameContainsWhitespace = errors.New("taxonomy: tag name contains whitespace")
	ErrTagNameRequired           = errors.New("taxonomy: tag name required")
	ErrTagUUIDInvalid            = errors.New("taxonomy: invalid tag UUID")
)

func newValidationError(field string, e error) error {
	return fmt.Errorf("%s: %w", field, e)
}
