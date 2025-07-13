// Copyright (c) VirtualTam
// SPDX-License-Identifier: MIT

package user

import "errors"

var (
	ErrNotFound                     = errors.New("user: not found")
	ErrDisplayNameRequired          = errors.New("user: display name required")
	ErrEmailAlreadyRegistered       = errors.New("user: email already registered")
	ErrEmailRequired                = errors.New("user: email required")
	ErrNickNameAlreadyRegistered    = errors.New("user: nickname already registered")
	ErrNickNameInvalid              = errors.New("user: invalid nickname")
	ErrNickNameRequired             = errors.New("user: nickname required")
	ErrPasswordConfirmationMismatch = errors.New("user: new password and confirmation do not match")
	ErrPasswordHashRequired         = errors.New("user: password hash required")
	ErrPasswordIncorrect            = errors.New("user: incorrect password")
	ErrPasswordRequired             = errors.New("user: password required")
	ErrUUIDRequired                 = errors.New("user: UUID required")
)
