package user

import "errors"

var (
	ErrNotFound                  error = errors.New("not found")
	ErrEmailAlreadyRegistered    error = errors.New("email already registered")
	ErrEmailRequired             error = errors.New("email required")
	ErrPasswordIncorrect         error = errors.New("incorrect password")
	ErrPasswordRequired          error = errors.New("password required")
	ErrPasswordHashRequired      error = errors.New("password hash required")
	ErrRememberTokenRequired     error = errors.New("remember token required")
	ErrRememberTokenHashRequired error = errors.New("remember token hash required")
	ErrUUIDRequired              error = errors.New("UUID required")
)
