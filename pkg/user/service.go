package user

import (
	"strings"
	"time"

	"github.com/virtualtam/yawbe/pkg/hash"
	"golang.org/x/crypto/bcrypt"
)

// Service handles operations for the user domain.
type Service struct {
	r    Repository
	hmac *hash.HMAC
}

// NewService initializes and returns a User Repository.
func NewService(r Repository, hmacKey string) *Service {
	hmac := hash.NewHMAC(hmacKey)
	return &Service{
		r:    r,
		hmac: hmac,
	}
}

// All returns a list of all users.
func (s *Service) All() ([]User, error) {
	return s.r.GetAllUsers()
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

// ByRememberToken returns the user corresponding to a giver RememberToken.
func (s *Service) ByRememberToken(rememberToken string) (User, error) {
	user := User{RememberToken: rememberToken}

	err := s.runValidationFuncs(
		&user,
		s.requireRememberToken,
		s.hashRememberToken,
		s.requireRememberTokenHash,
	)
	if err != nil {
		return User{}, err
	}

	return s.r.GetUserByRememberTokenHash(user.RememberTokenHash)
}

// Update updates an existing user.
func (s *Service) Update(user User) error {
	err := s.runValidationFuncs(
		&user,
		s.requireUUID,
		s.normalizeEmail,
		s.requireEmail,
		s.requirePasswordHash,
		s.hashRememberToken,
		s.refreshUpdatedAt,
	)
	if err != nil {
		return err
	}

	return s.r.UpdateUser(user)
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

	return s.r.GetUserByEmail(user.Email)
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

func (s *Service) hashRememberToken(user *User) error {
	if user.RememberToken == "" {
		return nil
	}

	hash, err := s.hmac.Hash(user.RememberToken)
	if err != nil {
		return err
	}

	user.RememberTokenHash = hash

	return nil
}

func (s *Service) normalizeEmail(user *User) error {
	user.Email = strings.ToLower(user.Email)
	user.Email = strings.TrimSpace(user.Email)

	return nil
}

func (s *Service) requireEmail(user *User) error {
	if user.Email == "" {
		return ErrEmailRequired
	}
	return nil
}

func (s *Service) requirePasswordHash(user *User) error {
	if user.PasswordHash == "" {
		return ErrPasswordHashRequired
	}
	return nil
}

func (s *Service) requireRememberToken(user *User) error {
	if user.RememberToken == "" {
		return ErrRememberTokenRequired
	}

	return nil
}

func (s *Service) requireRememberTokenHash(user *User) error {
	if user.RememberToken == "" {
		return ErrRememberTokenHashRequired
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
