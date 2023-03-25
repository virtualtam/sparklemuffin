package postgresql

import "time"

type Session struct {
	UserUUID               string    `db:"user_uuid"`
	RememberTokenHash      string    `db:"remember_token_hash"`
	RememberTokenExpiresAt time.Time `db:"remember_token_expires_at"`
}
