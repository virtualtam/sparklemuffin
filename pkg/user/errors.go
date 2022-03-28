package user

import "errors"

var (
	ErrNotFound                  error = errors.New("not found")
	ErrEmailRequired             error = errors.New("email required")
	ErrPasswordIncorrect         error = errors.New("incorrect password")
	ErrPasswordHashRequired      error = errors.New("password hash required")
	ErrRememberTokenRequired     error = errors.New("remember token required")
	ErrRememberTokenHashRequired error = errors.New("remember token hash required")
)
