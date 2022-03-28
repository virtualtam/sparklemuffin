package user

import "errors"

var (
	ErrNotFound          error = errors.New("not found")
	ErrEmailRequired     error = errors.New("email required")
	ErrPasswordIncorrect error = errors.New("incorrect password")
)
