package user

// Repository provides access to the User repository.
type Repository interface {
	// AddUser saves a new user.
	AddUser(User) error

	// GetAllUsers returns a list of all User accounts.
	GetAllUsers() ([]User, error)

	// GetUserByEmail returns the User registered with a given email address.
	GetUserByEmail(string) (User, error)

	// GetUserByRememberTokenHash returns the User corresponding to a given
	// RememberToken hash.
	GetUserByRememberTokenHash(string) (User, error)

	// IsUserEmailRegistered returns whether there is an existing user
	// registered with this email address.
	IsUserEmailRegistered(email string) (bool, error)

	// UpdateUser updates an existing user.
	UpdateUser(User) error

	// UpdateUserRememberToken updates an existing user's remember token hash.
	UpdateUserRememberTokenHash(User) error
}
