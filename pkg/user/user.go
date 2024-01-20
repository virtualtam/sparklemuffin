// Copyright (c) VirtualTam
// SPDX-License-Identifier: MIT

package user

import "time"

// User represents a registered user.
type User struct {
	// UUID is the internal identifier for this User.
	UUID string

	// Email is the identifier a User logs in with.
	Email string

	// NickName is the handle used in user-specific URLs, and may only contain
	// alphanumerical characters, the dash character, or the underscore character.
	NickName string

	// DisplayName is the handle used in the Web interface for this User.
	DisplayName string

	// Password is the clear-text password for this User, that will be set when
	// creating or updating the User, and cleared once it has been hashed.
	Password string

	// PasswordHash contains the securely hashed password for this User.
	PasswordHash string

	// IsAdmin represents whether this User has administration privileges.
	IsAdmin bool

	CreatedAt time.Time
	UpdatedAt time.Time
}

// InfoUpdate represents an account information update for an authenticated
// user.
type InfoUpdate struct {
	UUID        string
	Email       string
	NickName    string
	DisplayName string
	UpdatedAt   time.Time
}

// PasswordHashUpdate represents a password change for an authenticated user.
type PasswordUpdate struct {
	UUID                    string
	CurrentPassword         string
	NewPassword             string
	NewPasswordConfirmation string
}

// PasswordHashUpdate represents a password hash change for an authenticated user.
type PasswordHashUpdate struct {
	UUID         string
	PasswordHash string
	UpdatedAt    time.Time
}
