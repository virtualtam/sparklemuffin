package session

// Repository provides access to the user Web Session repoository.
type Repository interface {
	// SessionAdd saves a new user Session.
	SessionAdd(Session) error

	// SessionGetByRememberTokenHash returns the Session corresponding to a
	// given remember token hash.
	SessionGetByRememberTokenHash(hash string) (Session, error)
}
