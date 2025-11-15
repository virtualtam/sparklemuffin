// Copyright (c) VirtualTam
// SPDX-License-Identifier: MIT

package user

import (
	"context"
	"slices"
)

var _ Repository = &FakeRepository{}

type FakeRepository struct {
	// TODO refactor with map[uuid]User
	Users []User
}

func (r *FakeRepository) UserAdd(_ context.Context, user User) error {
	r.Users = append(r.Users, user)
	return nil
}

func (r *FakeRepository) UserDeleteByUUID(_ context.Context, userUUID string) error {
	for index, user := range r.Users {
		if user.UUID == userUUID {
			r.Users = slices.Delete(r.Users, index, index+1)
		}
	}

	return ErrNotFound
}

func (r *FakeRepository) UserGetAll(_ context.Context) ([]User, error) {
	return r.Users, nil
}

func (r *FakeRepository) UserGetByEmail(_ context.Context, email string) (User, error) {
	for _, user := range r.Users {
		if user.Email == email {
			return user, nil
		}
	}

	return User{}, ErrNotFound
}

func (r *FakeRepository) UserGetByNickName(_ context.Context, nick string) (User, error) {
	for _, user := range r.Users {
		if user.NickName == nick {
			return user, nil
		}
	}

	return User{}, ErrNotFound
}

func (r *FakeRepository) UserGetByUUID(_ context.Context, userUUID string) (User, error) {
	for _, user := range r.Users {
		if user.UUID == userUUID {
			return user, nil
		}
	}

	return User{}, ErrNotFound
}

func (r *FakeRepository) UserIsEmailRegistered(_ context.Context, email string) (bool, error) {
	registered := false

	for _, user := range r.Users {
		if user.Email == email {
			registered = true
			break
		}
	}

	return registered, nil
}

func (r *FakeRepository) UserIsNickNameRegistered(_ context.Context, nick string) (bool, error) {
	registered := false

	for _, user := range r.Users {
		if user.NickName == nick {
			registered = true
			break
		}
	}

	return registered, nil
}

func (r *FakeRepository) UserUpdate(_ context.Context, user User) error {
	for index, existingUser := range r.Users {
		if existingUser.UUID == user.UUID {
			r.Users[index] = user
			return nil
		}
	}

	return ErrNotFound
}

func (r *FakeRepository) UserUpdateInfo(_ context.Context, info InfoUpdate) error {
	for index, existingUser := range r.Users {
		if existingUser.UUID == info.UUID {
			r.Users[index].Email = info.Email
			r.Users[index].UpdatedAt = info.UpdatedAt
			return nil
		}
	}

	return ErrNotFound
}

func (r *FakeRepository) UserUpdatePasswordHash(_ context.Context, passwordHash PasswordHashUpdate) error {
	for index, existingUser := range r.Users {
		if existingUser.UUID == passwordHash.UUID {
			r.Users[index].PasswordHash = passwordHash.PasswordHash
			r.Users[index].UpdatedAt = passwordHash.UpdatedAt
			return nil
		}
	}

	return ErrNotFound
}
