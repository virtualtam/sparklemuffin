package memory

type User struct {
	Email             string
	PasswordHash      string
	RememberTokenHash string
	IsAdmin           bool
}
