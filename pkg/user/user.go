package user

import "time"

// User represents an authenticated user.
type User struct {
	UUID              string
	Email             string
	Password          string
	PasswordHash      string
	RememberToken     string
	RememberTokenHash string
	IsAdmin           bool

	CreatedAt time.Time
	UpdatedAt time.Time
}
