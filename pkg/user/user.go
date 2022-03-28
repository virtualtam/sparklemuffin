package user

// User represents an authenticated user.
type User struct {
	Email             string
	Password          string
	PasswordHash      string
	RememberToken     string
	RememberTokenHash string
}
