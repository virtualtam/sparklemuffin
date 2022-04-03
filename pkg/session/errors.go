package session

import "errors"

var (
	ErrNotFound                  error = errors.New("not found")
	ErrRememberTokenRequired     error = errors.New("remember token required")
	ErrRememberTokenHashRequired error = errors.New("remember token hash required")
	ErrUserUUIDRequired          error = errors.New("user UUID required")
)
