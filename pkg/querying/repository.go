package querying

import "github.com/virtualtam/yawbe/pkg/bookmark"

// Repository provides access to query user bookmarks.
type Repository interface {
	// BookmarkGetCount returns the number of bookmarks for a given user.
	BookmarkGetCount(userUUID string, visibility Visibility) (uint, error)

	// BookmarkGetN returns at most n bookmarks for a given user, starting at
	// a given offset.
	BookmarkGetN(userUUID string, visibility Visibility, n uint, offset uint) ([]bookmark.Bookmark, error)

	// BookmarkSearchCount returns the number of bookmarks for a given user and
	// search terms.
	BookmarkSearchCount(userUUID string, visibility Visibility, searchTerms string) (uint, error)

	// BookmarkSearchN returns at most n bookmarks for a given user and search
	// terms, starting at a given offset.
	BookmarkSearchN(userUUID string, visibility Visibility, searchTerms string, n uint, offset uint) ([]bookmark.Bookmark, error)

	// OwnerGetByUUID returns the Owner corresponding to a given UUID.
	OwnerGetByUUID(string) (Owner, error)
}
