package bookmark

// TagNameUpdate represents a tag name update for all bookmarks for an authenticated user.
type TagNameUpdate struct {
	UserUUID    string
	CurrentName string
	NewName     string
}
