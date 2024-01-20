// Copyright (c) VirtualTam
// SPDX-License-Identifier: MIT

package user

// Repository provides access to the User repository.
type Repository interface {
	// UserAdd saves a new user.
	UserAdd(User) error

	// UserDelete deletes an existing user and all related data.
	UserDeleteByUUID(uuid string) error

	// UserGetAll returns a list of all User accounts.
	UserGetAll() ([]User, error)

	// UserGetByEmail returns the User registered with a given email address.
	UserGetByEmail(email string) (User, error)

	// UserGetByNickName returns the User registered with a given nickname.
	UserGetByNickName(nick string) (User, error)

	// UserGetByUUID returns the User corresponding to a given UUID.
	UserGetByUUID(string) (User, error)

	// UserIsEmailRegistered returns whether there is an existing user
	// registered with this email address.
	UserIsEmailRegistered(email string) (bool, error)

	// UserIsNickNameRegistered returns whether there is an existing user
	// registered with this nickname.
	UserIsNickNameRegistered(nick string) (bool, error)

	// UserUpdate updates an existing user.
	UserUpdate(User) error

	// UserUpdateInfo updates an existing user's account information.
	UserUpdateInfo(InfoUpdate) error

	// UserUpdatePasswordHash updates an existing user's account password hash.
	UserUpdatePasswordHash(PasswordHashUpdate) error
}
