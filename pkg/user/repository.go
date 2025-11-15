// Copyright (c) VirtualTam
// SPDX-License-Identifier: MIT

package user

import (
	"context"
)

// Repository provides access to the User repository.
type Repository interface {
	// UserAdd saves a new user.
	UserAdd(ctx context.Context, u User) error

	// UserDeleteByUUID deletes an existing user and all related data.
	UserDeleteByUUID(ctx context.Context, uuid string) error

	// UserGetAll returns a list of all User accounts.
	UserGetAll(ctx context.Context) ([]User, error)

	// UserGetByEmail returns the User registered with a given email address.
	UserGetByEmail(ctx context.Context, email string) (User, error)

	// UserGetByNickName returns the User registered with a given nickname.
	UserGetByNickName(ctx context.Context, nick string) (User, error)

	// UserGetByUUID returns the User corresponding to a given UUID.
	UserGetByUUID(ctx context.Context, uuid string) (User, error)

	// UserIsEmailRegistered returns whether there is an existing user
	// registered with this email address.
	UserIsEmailRegistered(ctx context.Context, email string) (bool, error)

	// UserIsNickNameRegistered returns whether there is an existing user
	// registered with this nickname.
	UserIsNickNameRegistered(ctx context.Context, nick string) (bool, error)

	// UserUpdate updates an existing user.
	UserUpdate(ctx context.Context, u User) error

	// UserUpdateInfo updates an existing user's account information.
	UserUpdateInfo(ctx context.Context, info InfoUpdate) error

	// UserUpdatePasswordHash updates an existing user's account password hash.
	UserUpdatePasswordHash(ctx context.Context, passwordHashUpdate PasswordHashUpdate) error
}
