// Copyright VirtualTam 2022, 2026
// SPDX-License-Identifier: MIT

package bookmark

import (
	"errors"
)

var (
	ErrNotFound             = errors.New("bookmark: not found")
	ErrTitleRequired        = errors.New("bookmark: title required")
	ErrUIDInvalid           = errors.New("bookmark: invalid UID")
	ErrUIDRequired          = errors.New("bookmark: UID required")
	ErrURLAlreadyRegistered = errors.New("bookmark: URL already registered")
	ErrURLInvalid           = errors.New("bookmark: invalid URL")
	ErrURLNoHost            = errors.New("bookmark: URL has no host")
	ErrURLNoScheme          = errors.New("bookmark: URL has no scheme")
	ErrURLRequired          = errors.New("bookmark: URL required")
)
