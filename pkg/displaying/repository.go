package displaying

import "github.com/virtualtam/yawbe/pkg/bookmark"

type Repository interface {
	// BookmarkGetCount returns the number of bookmarks for a  given user.
	BookmarkGetCount(userUUID string) (int, error)

	// BookmarkGetN returns at most n bookmarks for a given user, starting at
	BookmarkGetN(userUUID string, n int, offset int) ([]bookmark.Bookmark, error)
}
