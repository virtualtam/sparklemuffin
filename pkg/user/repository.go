package user

// Repository provides access to the User repository.
type Repository interface {
	// AddUser saves a new user.
	AddUser(User) error

	// GetUserByEmail returns the User registered with a given email address.
	GetUserByEmail(string) (User, error)
}
