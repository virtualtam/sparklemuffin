package user

// Repository provides access to the User repository.
type Repository interface {
	// AddUser saves a new user.
	AddUser(User) error

	// DeleteUser deletes an existing user and all related data.
	DeleteUserByUUID(string) error

	// GetAllUsers returns a list of all User accounts.
	GetAllUsers() ([]User, error)

	// GetUserByEmail returns the User registered with a given email address.
	GetUserByEmail(string) (User, error)

	// GetUserByRememberTokenHash returns the User corresponding to a given
	// RememberToken hash.
	GetUserByRememberTokenHash(string) (User, error)

	// GetUserByUUID returns the User corresponding to a given UUID.
	GetUserByUUID(string) (User, error)

	// IsUserEmailRegistered returns whether there is an existing user
	// registered with this email address.
	IsUserEmailRegistered(email string) (bool, error)

	// UpdateUser updates an existing user.
	UpdateUser(User) error

	// UpdateInfo updates an existing user's account information.
	UpdateUserInfo(InfoUpdate) error

	// UpdatePasswordHash updates an existing user's account password hash.
	UpdateUserPasswordHash(PasswordHashUpdate) error

	// UpdateUserRememberToken updates an existing user's remember token hash.
	UpdateUserRememberTokenHash(User) error
}
