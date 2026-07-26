// Copyright VirtualTam 2022, 2026
// SPDX-License-Identifier: MIT

package user

import (
	"context"
	"regexp"
	"strings"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

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

// NewUser initializes and returns a new User.
func NewUser(email, nickName, displayName, password string) (User, error) {
	userUUID, err := uuid.NewRandom()
	if err != nil {
		return User{}, err
	}

	now := time.Now().UTC()

	u := User{
		CreatedAt: now,
		UpdatedAt: now,

		UUID: userUUID.String(),

		Email:       email,
		NickName:    nickName,
		DisplayName: displayName,
		Password:    password,
	}

	u.Normalize()

	return u, nil
}

// NewAdminUser initializes and returns a new User with administration privileges.
func NewAdminUser(email, nickName, displayName, password string) (User, error) {
	u, err := NewUser(email, nickName, displayName, password)
	if err != nil {
		return User{}, err
	}

	u.IsAdmin = true

	return u, nil
}

// Normalize sanitizes and normalizes all fields.
func (u *User) Normalize() {
	u.normalizeDisplayName()
	u.normalizeEmail()
	u.normalizeNickName()
}

// ValidateForAddition ensures mandatory fields are properly set when adding a new User.
func (u *User) ValidateForAddition(ctx context.Context, r ValidationRepository) error {
	fns := []func() error{
		u.requireEmail,
		u.ensureEmailIsNotRegistered(ctx, r),
		u.requireNickName,
		u.ensureNickNameIsValid,
		u.ensureNickNameIsNotRegistered(ctx, r),
		u.requireDisplayName,
		u.requirePassword,
		u.requirePasswordLength,
		u.hashPassword,
		u.requirePasswordHash,
		u.requireUUID,
	}

	for _, fn := range fns {
		if err := fn(); err != nil {
			return err
		}
	}

	return nil
}

// ValidateForUpdate ensures mandatory fields are properly set when updating an existing User.
func (u *User) ValidateForUpdate(ctx context.Context, r ValidationRepository) error {
	fns := []func() error{
		u.requireUUID,
		u.requireEmail,
		u.ensureEmailIsNotRegisteredToAnotherUser(ctx, r),
		u.requireNickName,
		u.ensureNickNameIsValid,
		u.ensureNickNameIsNotRegisteredToAnotherUser(ctx, r),
		u.requireDisplayName,
		u.requirePassword,
		u.requirePasswordLength,
		u.hashPassword,
		u.requirePasswordHash,
	}

	for _, fn := range fns {
		if err := fn(); err != nil {
			return err
		}
	}

	return nil
}

// ValidateForInfoUpdate ensures mandatory fields are properly set when updating an existing User's information.
func (u *User) ValidateForInfoUpdate(ctx context.Context, r ValidationRepository) error {
	fns := []func() error{
		u.requireUUID,
		u.requireEmail,
		u.ensureEmailIsNotRegisteredToAnotherUser(ctx, r),
		u.requireNickName,
		u.ensureNickNameIsValid,
		u.ensureNickNameIsNotRegisteredToAnotherUser(ctx, r),
		u.requireDisplayName,
	}

	for _, fn := range fns {
		if err := fn(); err != nil {
			return err
		}
	}

	return nil
}

// ValidateForPasswordHashUpdate ensures mandatory fields are properly set when updating an existing User's
// password hash.
func (u *User) ValidateForPasswordHashUpdate() error {
	fns := []func() error{
		u.requireUUID,
		u.requirePassword,
		u.requirePasswordLength,
		u.hashPassword,
		u.requirePasswordHash,
	}

	for _, fn := range fns {
		if err := fn(); err != nil {
			return err
		}
	}

	return nil
}

func (u *User) normalizeDisplayName() {
	u.DisplayName = strings.TrimSpace(u.DisplayName)
}

func (u *User) normalizeEmail() {
	u.Email = strings.ToLower(u.Email)
	u.Email = strings.TrimSpace(u.Email)
}

func (u *User) normalizeNickName() {
	u.NickName = strings.ToLower(u.NickName)
	u.NickName = strings.TrimSpace(u.NickName)
}

func (u *User) requireDisplayName() error {
	if u.DisplayName == "" {
		return ErrDisplayNameRequired
	}
	return nil
}

func (u *User) requireEmail() error {
	if u.Email == "" {
		return ErrEmailRequired
	}
	return nil
}

func (u *User) requireNickName() error {
	if u.NickName == "" {
		return ErrNickNameRequired
	}
	return nil
}

func (u *User) requirePassword() error {
	if u.Password == "" {
		return ErrPasswordRequired
	}
	return nil
}

const (
	MinPasswordLength = 8
)

func (u *User) requirePasswordLength() error {
	if len(u.Password) < MinPasswordLength {
		return ErrPasswordTooShort
	}
	return nil
}

func (u *User) requirePasswordHash() error {
	if u.PasswordHash == "" {
		return ErrPasswordHashRequired
	}
	return nil
}

func (u *User) requireUUID() error {
	if u.UUID == "" {
		return ErrUUIDRequired
	}

	return nil
}

func (u *User) ensureEmailIsNotRegistered(ctx context.Context, r ValidationRepository) func() error {
	return func() error {
		registered, err := r.UserIsEmailRegistered(ctx, u.Email)
		if err != nil {
			return err
		}
		if registered {
			return ErrEmailAlreadyRegistered
		}
		return nil
	}
}

func (u *User) ensureEmailIsNotRegisteredToAnotherUser(ctx context.Context, r ValidationRepository) func() error {
	return func() error {
		registered, err := r.UserIsEmailRegisteredToAnotherUser(ctx, u.UUID, u.Email)
		if err != nil {
			return err
		}
		if registered {
			return ErrEmailAlreadyRegistered
		}
		return nil
	}
}

func (u *User) ensureNickNameIsNotRegistered(ctx context.Context, r ValidationRepository) func() error {
	return func() error {
		registered, err := r.UserIsNickNameRegistered(ctx, u.NickName)
		if err != nil {
			return err
		}
		if registered {
			return ErrNickNameAlreadyRegistered
		}
		return nil
	}
}

func (u *User) ensureNickNameIsNotRegisteredToAnotherUser(ctx context.Context, r ValidationRepository) func() error {
	return func() error {
		registered, err := r.UserIsNickNameRegisteredToAnotherUser(ctx, u.UUID, u.NickName)
		if err != nil {
			return err
		}
		if registered {
			return ErrNickNameAlreadyRegistered
		}
		return nil
	}
}

var (
	nickNameRegex = regexp.MustCompile(`^[a-zA-Z][a-zA-Z0-9_-]+$`)
)

func (u *User) ensureNickNameIsValid() error {
	if !nickNameRegex.MatchString(u.NickName) {
		return ErrNickNameInvalid
	}

	return nil
}

func (u *User) hashPassword() error {
	h, err := bcrypt.GenerateFromPassword([]byte(u.Password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	u.PasswordHash = string(h)

	// Clear the clear-text password as soon as it is hashed.
	u.Password = ""

	return nil
}

// InfoUpdate represents an account information update for an authenticated
// user.
type InfoUpdate struct {
	UserUUID    string
	Email       string
	NickName    string
	DisplayName string
	UpdatedAt   time.Time
}

// PasswordUpdate represents a password change for an authenticated user.
type PasswordUpdate struct {
	UserUUID                string
	CurrentPassword         string
	NewPassword             string
	NewPasswordConfirmation string
}

// PasswordHashUpdate represents a password hash change for an authenticated user.
type PasswordHashUpdate struct {
	UserUUID     string
	PasswordHash string
	UpdatedAt    time.Time
}
