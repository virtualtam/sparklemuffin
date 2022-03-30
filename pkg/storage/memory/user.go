package memory

import "time"

type User struct {
	UUID              string
	Email             string
	PasswordHash      string
	RememberTokenHash string
	IsAdmin           bool

	CreatedAt time.Time
	UpdatedAt time.Time
}
