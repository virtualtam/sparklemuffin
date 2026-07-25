// Copyright VirtualTam 2022, 2026
// SPDX-License-Identifier: MIT

package user

import (
	"context"
	"errors"
	"time"

	"golang.org/x/crypto/bcrypt"
)

// Service handles operations for the user domain.
type Service struct {
	r Repository
}

// NewService initializes and returns a User Service.
func NewService(r Repository) *Service {
	return &Service{
		r: r,
	}
}

// Add adds a new User.
func (s *Service) Add(ctx context.Context, user User) error {
	user.Normalize()
	if err := user.ValidateForAddition(ctx, s.r); err != nil {
		return err
	}

	return s.r.UserAdd(ctx, user)
}

// All returns a list of all users.
func (s *Service) All(ctx context.Context) ([]User, error) {
	return s.r.UserGetAll(ctx)
}

// Authenticate checks user-submitted credentials to determine whether a user
// submitted the correct login information.
func (s *Service) Authenticate(ctx context.Context, email, password string) (User, error) {
	user, err := s.getUserByEmail(ctx, email)
	if err != nil {
		return User{}, err
	}

	err = bcrypt.CompareHashAndPassword(
		[]byte(user.PasswordHash),
		[]byte(password),
	)

	if errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
		return User{}, ErrPasswordIncorrect
	} else if err != nil {
		return User{}, err
	}

	return user, nil
}

// ByNickName returns the user corresponding to a given NickName.
func (s *Service) ByNickName(ctx context.Context, nick string) (User, error) {
	user := User{NickName: nick}
	user.normalizeNickName()

	if err := user.requireNickName(); err != nil {
		return User{}, err
	}
	if err := user.ensureNickNameIsValid(); err != nil {
		return User{}, err
	}

	return s.r.UserGetByNickName(ctx, user.NickName)
}

// ByUUID returns the user corresponding to a given UUID.
func (s *Service) ByUUID(ctx context.Context, userUUID string) (User, error) {
	user := User{UUID: userUUID}

	if err := user.requireUUID(); err != nil {
		return User{}, err
	}

	return s.r.UserGetByUUID(ctx, user.UUID)
}

// DeleteByUUID deletes an existing user and all related data.
func (s *Service) DeleteByUUID(ctx context.Context, userUUID string) error {
	user := User{UUID: userUUID}

	if err := user.requireUUID(); err != nil {
		return err
	}

	return s.r.UserDeleteByUUID(ctx, userUUID)
}

// Update updates an existing user.
func (s *Service) Update(ctx context.Context, user User) error {
	user.Normalize()
	user.UpdatedAt = time.Now().UTC()

	if err := user.ValidateForUpdate(ctx, s.r); err != nil {
		return err
	}

	return s.r.UserUpdate(ctx, user)
}

// UpdateInfo updates an existing user's account information.
func (s *Service) UpdateInfo(ctx context.Context, info InfoUpdate) error {
	user := User{
		UUID:        info.UserUUID,
		Email:       info.Email,
		NickName:    info.NickName,
		DisplayName: info.DisplayName,
	}
	user.Normalize()

	if err := user.ValidateForInfoUpdate(ctx, s.r); err != nil {
		return err
	}

	info.UpdatedAt = time.Now().UTC()

	return s.r.UserUpdateInfo(ctx, info)
}

// UpdatePassword updates an existing user's password.
func (s *Service) UpdatePassword(ctx context.Context, passwordUpdate PasswordUpdate) error {
	// validate current password
	user := User{
		UUID:     passwordUpdate.UserUUID,
		Password: passwordUpdate.CurrentPassword,
	}

	if err := user.requireUUID(); err != nil {
		return err
	}
	if err := user.requirePassword(); err != nil {
		return err
	}

	existingUser, err := s.ByUUID(ctx, user.UUID)
	if err != nil {
		return err
	}

	err = bcrypt.CompareHashAndPassword(
		[]byte(existingUser.PasswordHash),
		[]byte(user.Password),
	)
	if errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
		return ErrPasswordIncorrect
	}
	if err != nil {
		return err
	}

	// validate new password
	if passwordUpdate.NewPassword != passwordUpdate.NewPasswordConfirmation {
		return ErrPasswordConfirmationMismatch
	}

	// hash new password
	user = User{
		UUID:     passwordUpdate.UserUUID,
		Password: passwordUpdate.NewPassword,
	}

	if err := user.ValidateForPasswordHashUpdate(); err != nil {
		return err
	}

	passwordHashUpdate := PasswordHashUpdate{
		UserUUID:     user.UUID,
		PasswordHash: user.PasswordHash,
		UpdatedAt:    time.Now().UTC(),
	}

	return s.r.UserUpdatePasswordHash(ctx, passwordHashUpdate)
}

func (s *Service) getUserByEmail(ctx context.Context, email string) (User, error) {
	user := User{Email: email}
	user.normalizeEmail()

	if err := user.requireEmail(); err != nil {
		return User{}, err
	}

	return s.r.UserGetByEmail(ctx, user.Email)
}
