package querying

import "github.com/virtualtam/yawbe/pkg/bookmark"

// Repository provides access to query user bookmarks.
type Repository interface {
	// BookmarkGetCount returns the number of bookmarks for a given user.
	BookmarkGetCount(userUUID string) (int, error)

	// BookmarkGetN returns at most n bookmarks for a given user, starting at
	// a given offset.
	BookmarkGetN(userUUID string, n int, offset int) ([]bookmark.Bookmark, error)

	// BookmarkSearchCount returns the number of bookmarks for a given user and
	// search terms.
	BookmarkSearchCount(userUUID string, searchTerms string) (int, error)

	// BookmarkSearchN returns at most n bookmarks for a given user and search
	// terms, starting at a given offset.
	BookmarkSearchN(userUUID string, searchTerms string, n int, offset int) ([]bookmark.Bookmark, error)
}
