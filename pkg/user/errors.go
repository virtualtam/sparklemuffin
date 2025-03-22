// Copyright (c) VirtualTam
// SPDX-License-Identifier: MIT

package user

import "errors"

var (
	ErrNotFound                     = errors.New("not found")
	ErrDisplayNameRequired          = errors.New("display name required")
	ErrEmailAlreadyRegistered       = errors.New("email already registered")
	ErrEmailRequired                = errors.New("email required")
	ErrNickNameAlreadyRegistered    = errors.New("nickname already registered")
	ErrNickNameInvalid              = errors.New("invalid nickname")
	ErrNickNameRequired             = errors.New("nickname required")
	ErrPasswordConfirmationMismatch = errors.New("new password and confirmation do not match")
	ErrPasswordHashRequired         = errors.New("password hash required")
	ErrPasswordIncorrect            = errors.New("incorrect password")
	ErrPasswordRequired             = errors.New("password required")
	ErrUUIDRequired                 = errors.New("UUID required")
)
