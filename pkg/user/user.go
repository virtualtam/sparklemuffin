package user

// User represents an authenticated user.
type User struct {
	UUID              string
	Email             string
	Password          string
	PasswordHash      string
	RememberToken     string
	RememberTokenHash string
}
