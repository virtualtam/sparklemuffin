package user

import "errors"

var (
	ErrNotFound                     error = errors.New("not found")
	ErrEmailAlreadyRegistered       error = errors.New("email already registered")
	ErrEmailRequired                error = errors.New("email required")
	ErrPasswordConfirmationMismatch error = errors.New("new password and confirmation do not match")
	ErrPasswordHashRequired         error = errors.New("password hash required")
	ErrPasswordIncorrect            error = errors.New("incorrect password")
	ErrPasswordRequired             error = errors.New("password required")
	ErrUUIDRequired                 error = errors.New("UUID required")
)
