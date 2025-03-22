// Copyright (c) VirtualTam
// SPDX-License-Identifier: MIT

package session

import "errors"

var (
	ErrNotFound                  = errors.New("not found")
	ErrRememberTokenRequired     = errors.New("remember token required")
	ErrRememberTokenHashRequired = errors.New("remember token hash required")
)
