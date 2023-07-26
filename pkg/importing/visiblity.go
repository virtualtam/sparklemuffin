package importing

import "errors"

// Visibility defines how bookmarks are imported.
type Visibility string

const (
	// Keep the existing value for imported bookmarks; if missing,
	// bookmarks will be imported as public.
	VisibilityDefault Visibility = "default"

	// Import all bookmarks as private.
	VisibilityPrivate Visibility = "private"

	// Import all bookmarks as public.
	VisibilityPublic Visibility = "public"
)

var (
	ErrVisibilityInvalid error = errors.New("invalid value for visibility")
)
