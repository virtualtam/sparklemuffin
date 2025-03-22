// Copyright (c) VirtualTam
// SPDX-License-Identifier: MIT

package session

import "errors"

var (
	ErrNotFound                  error = errors.New("not found")
	ErrRememberTokenRequired     error = errors.New("remember token required")
	ErrRememberTokenHashRequired error = errors.New("remember token hash required")
)
