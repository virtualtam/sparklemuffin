package querying

// Visibility represents a visibility filter for bookmarks.
type Visibility string

const (
	// Public and private bookmarks.
	VisibilityAll Visibility = "all"

	// Private bookmarks only.
	VisibilityPrivate Visibility = "private"

	// Public bookmarks only.
	VisibilityPublic Visibility = "public"
)
