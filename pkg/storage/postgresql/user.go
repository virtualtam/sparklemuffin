package postgresql

type User struct {
	UUID              string `db:"uuid"`
	Email             string `db:"email"`
	PasswordHash      string `db:"password_hash"`
	RememberTokenHash string `db:"remember_token_hash"`
	IsAdmin           bool   `db:"is_admin"`
}
