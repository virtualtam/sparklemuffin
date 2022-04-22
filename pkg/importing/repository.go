package importing

import "github.com/virtualtam/yawbe/pkg/bookmark"

type Repository interface {
	bookmark.ValidationRepository

	// BookmarkAddMany adds a collection of new bookmarks.
	BookmarkAddMany(bookmarks []bookmark.Bookmark) error
}
