package querying

// Owner exposes public matadata for the User owning the displayed bookmarks.
type Owner struct {
	// UUID is the internal identifier for this User.
	UUID string

	// NickName is the handle used in user-specific URLs, and may only contain
	// alphanumerical characters, the dash character, or the underscore character.
	NickName string

	// DisplayName is the handle used in the Web interface for this User.
	DisplayName string
}
