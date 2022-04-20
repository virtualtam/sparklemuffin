package bookmark

import "time"

// Bookmark represents a Web bookmark.
type Bookmark struct {
	UID      string
	UserUUID string

	URL         string
	Title       string
	Description string

	Private bool
	Tags    []string

	CreatedAt time.Time
	UpdatedAt time.Time
}
