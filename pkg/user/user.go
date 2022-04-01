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

// InfoUpdate represents an account information update for an authenticated
// user.
type InfoUpdate struct {
	UUID      string
	Email     string
	UpdatedAt time.Time
}

// PasswordHashUpdate represents a password change for an authenticated user.
type PasswordUpdate struct {
	UUID                    string
	CurrentPassword         string
	NewPassword             string
	NewPasswordConfirmation string
}

// PasswordHashUpdate represents a password hash change for an authenticated user.
type PasswordHashUpdate struct {
	UUID         string
	PasswordHash string
	UpdatedAt    time.Time
}
