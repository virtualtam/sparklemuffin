package user

import (
	"strings"

	"golang.org/x/crypto/bcrypt"
)

// Service handles operations for the user domain.
type Service struct {
	r Repository
}

// NewService initializes and returns a User Repository.
func NewService(r Repository) *Service {
	return &Service{
		r: r,
	}
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

func (s *Service) getUserByEmail(email string) (User, error) {
	user := User{Email: email}

	err := runValidationFuncs(
		&user,
		normalizeEmail,
		requireEmail,
	)
	if err != nil {
		return User{}, err
	}

	return s.r.GetUserByEmail(user.Email)
}

// validationFunc defines a function that can be applied to normalize or
// validate User data.
type validationFunc func(*User) error

// runValidationFuncs applies User normalization and validation functions and
// stops at the first encountered error.
func runValidationFuncs(user *User, fns ...validationFunc) error {
	for _, fn := range fns {
		if err := fn(user); err != nil {
			return err
		}
	}
	return nil
}

func normalizeEmail(user *User) error {
	user.Email = strings.ToLower(user.Email)
	user.Email = strings.TrimSpace(user.Email)

	return nil
}

func requireEmail(user *User) error {
	if user.Email == "" {
		return ErrEmailRequired
	}
	return nil
}
