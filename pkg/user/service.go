package user

import (
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
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
func (s *Service) Add(user User) error {
	err := s.runValidationFuncs(
		&user,
		s.normalizeEmail,
		s.requireEmail,
		s.ensureEmailIsNotRegistered,
		s.normalizeNickName,
		s.requireNickName,
		s.ensureNickNameIsNotRegistered,
		s.normalizeDisplayName,
		s.requireDisplayName,
		s.requirePassword,
		s.hashPassword,
		s.requirePasswordHash,
		s.generateUUID,
		s.requireUUID,
		s.setCreatedUpdatedAt,
	)
	if err != nil {
		return err
	}

	return s.r.UserAdd(user)
}

// All returns a list of all users.
func (s *Service) All() ([]User, error) {
	return s.r.UserGetAll()
}

// Authenticate checks user-submitted credentials to determine whether a user
// submitted the correct login information.
func (s *Service) Authenticate(email, password string) (User, error) {
	user, err := s.getUserByEmail(email)
	if err != nil {
		return User{}, err
	}

	err = bcrypt.CompareHashAndPassword(
		[]byte(user.PasswordHash),
		[]byte(password),
	)

	switch err {
	case nil:
		return user, nil
	case bcrypt.ErrMismatchedHashAndPassword:
		return User{}, ErrPasswordIncorrect
	default:
		return User{}, err
	}
}

// ByUUID returns the user corresponding to a given UUID.
func (s *Service) ByUUID(userUUID string) (User, error) {
	user := User{UUID: userUUID}

	err := s.runValidationFuncs(
		&user,
		s.requireUUID,
	)
	if err != nil {
		return User{}, err
	}

	return s.r.UserGetByUUID(user.UUID)
}

// DeleteByUUID deletes an existing user and all related data.
func (s *Service) DeleteByUUID(userUUID string) error {
	user := User{UUID: userUUID}

	err := s.runValidationFuncs(
		&user,
		s.requireUUID,
	)
	if err != nil {
		return err
	}

	return s.r.UserDeleteByUUID(userUUID)
}

// Update updates an existing user.
func (s *Service) Update(user User) error {
	err := s.runValidationFuncs(
		&user,
		s.requireUUID,
		s.normalizeEmail,
		s.requireEmail,
		s.ensureEmailIsNotRegisteredToAnotherUser,
		s.normalizeNickName,
		s.requireNickName,
		s.ensureNickNameIsNotRegisteredToAnotherUser,
		s.normalizeDisplayName,
		s.requireDisplayName,
		s.requirePassword,
		s.hashPassword,
		s.requirePasswordHash,
		s.refreshUpdatedAt,
	)
	if err != nil {
		return err
	}

	return s.r.UserUpdate(user)
}

// UpdateInfo updates an existing user's account information.
func (s *Service) UpdateInfo(info InfoUpdate) error {
	user := User{
		UUID:        info.UUID,
		Email:       info.Email,
		NickName:    info.NickName,
		DisplayName: info.DisplayName,
	}

	err := s.runValidationFuncs(
		&user,
		s.requireUUID,
		s.normalizeEmail,
		s.requireEmail,
		s.ensureEmailIsNotRegisteredToAnotherUser,
		s.normalizeNickName,
		s.requireNickName,
		s.ensureNickNameIsNotRegisteredToAnotherUser,
		s.normalizeDisplayName,
		s.requireDisplayName,
		s.refreshUpdatedAt,
	)
	if err != nil {
		return err
	}

	info.UpdatedAt = user.UpdatedAt

	return s.r.UserUpdateInfo(info)
}

// UpdatePassword updates an existing user's password.
func (s *Service) UpdatePassword(passwordUpdate PasswordUpdate) error {
	// validate current password
	user := User{
		UUID:     passwordUpdate.UUID,
		Password: passwordUpdate.CurrentPassword,
	}

	err := s.runValidationFuncs(
		&user,
		s.requireUUID,
		s.requirePassword,
	)
	if err != nil {
		return err
	}

	existingUser, err := s.ByUUID(user.UUID)
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
		UUID:     passwordUpdate.UUID,
		Password: passwordUpdate.NewPassword,
	}

	err = s.runValidationFuncs(
		&user,
		s.requireUUID,
		s.requirePassword,
		s.hashPassword,
		s.requirePasswordHash,
	)
	if err != nil {
		return err
	}

	return s.UpdatePasswordHash(user)
}

func (s *Service) UpdatePasswordHash(user User) error {
	err := s.runValidationFuncs(
		&user,
		s.requireUUID,
		s.requirePasswordHash,
		s.refreshUpdatedAt,
	)
	if err != nil {
		return err
	}

	passwordHashUpdate := PasswordHashUpdate{
		UUID:         user.UUID,
		PasswordHash: user.PasswordHash,
		UpdatedAt:    user.UpdatedAt,
	}

	return s.r.UserUpdatePasswordHash(passwordHashUpdate)
}

func (s *Service) getUserByEmail(email string) (User, error) {
	user := User{Email: email}

	err := s.runValidationFuncs(
		&user,
		s.normalizeEmail,
		s.requireEmail,
	)
	if err != nil {
		return User{}, err
	}

	return s.r.UserGetByEmail(user.Email)
}

// validationFunc defines a function that can be applied to normalize or
// validate User data.
type validationFunc func(*User) error

// runValidationFuncs applies User normalization and validation functions and
// stops at the first encountered error.
func (s *Service) runValidationFuncs(user *User, fns ...validationFunc) error {
	for _, fn := range fns {
		if err := fn(user); err != nil {
			return err
		}
	}
	return nil
}

func (s *Service) ensureNickNameIsNotRegistered(user *User) error {
	registered, err := s.r.UserIsNickNameRegistered(user.NickName)
	if err != nil {
		return err
	}
	if registered {
		return ErrNickNameAlreadyRegistered
	}
	return nil
}

func (s *Service) ensureNickNameIsNotRegisteredToAnotherUser(user *User) error {
	existingUser, err := s.r.UserGetByNickName(user.NickName)
	if errors.Is(err, ErrNotFound) {
		return nil
	}

	if existingUser.UUID == user.UUID {
		return nil
	}

	return ErrNickNameAlreadyRegistered
}

func (s *Service) ensureEmailIsNotRegistered(user *User) error {
	registered, err := s.r.UserIsEmailRegistered(user.Email)
	if err != nil {
		return err
	}
	if registered {
		return ErrEmailAlreadyRegistered
	}
	return nil
}

func (s *Service) ensureEmailIsNotRegisteredToAnotherUser(user *User) error {
	existingUser, err := s.r.UserGetByEmail(user.Email)
	if errors.Is(err, ErrNotFound) {
		return nil
	}

	if existingUser.UUID == user.UUID {
		return nil
	}

	return ErrEmailAlreadyRegistered
}

func (s *Service) generateUUID(user *User) error {
	generatedUUID, err := uuid.NewRandom()
	if err != nil {
		return err
	}

	user.UUID = generatedUUID.String()

	return nil
}

func (s *Service) hashPassword(user *User) error {
	h, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	user.PasswordHash = string(h)

	// wipe the clear-text password as soon as it is hashed
	user.Password = ""

	return nil
}

func (s *Service) normalizeDisplayName(user *User) error {
	user.DisplayName = strings.TrimSpace(user.DisplayName)
	return nil
}

func (s *Service) normalizeEmail(user *User) error {
	user.Email = strings.ToLower(user.Email)
	user.Email = strings.TrimSpace(user.Email)
	return nil
}

func (s *Service) normalizeNickName(user *User) error {
	user.NickName = strings.ToLower(user.NickName)
	user.NickName = strings.TrimSpace(user.NickName)
	return nil
}

func (s *Service) requireDisplayName(user *User) error {
	if user.DisplayName == "" {
		return ErrDisplayNameRequired
	}
	return nil
}

func (s *Service) requireEmail(user *User) error {
	if user.Email == "" {
		return ErrEmailRequired
	}
	return nil
}

func (s *Service) requireNickName(user *User) error {
	if user.NickName == "" {
		return ErrNickNameRequired
	}
	return nil
}

func (s *Service) requirePassword(user *User) error {
	if user.Password == "" {
		return ErrPasswordRequired
	}
	return nil
}

func (s *Service) requirePasswordHash(user *User) error {
	if user.PasswordHash == "" {
		return ErrPasswordHashRequired
	}
	return nil
}

func (s *Service) requireUUID(user *User) error {
	if user.UUID == "" {
		return ErrUUIDRequired
	}

	return nil
}

func (s *Service) refreshUpdatedAt(user *User) error {
	user.UpdatedAt = time.Now().UTC()
	return nil
}

func (s *Service) setCreatedUpdatedAt(user *User) error {
	now := time.Now().UTC()
	user.CreatedAt = now
	user.UpdatedAt = now

	return nil
}
