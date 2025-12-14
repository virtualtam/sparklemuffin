// Copyright (c) VirtualTam
// SPDX-License-Identifier: MIT

package session

import "errors"

var (
	ErrHmacKeyRequired           = errors.New("session: hmac key is required")
	ErrNotFound                  = errors.New("session: not found")
	ErrRememberTokenRequired     = errors.New("session: remember token required")
	ErrRememberTokenHashRequired = errors.New("session: remember token hash required")
)
