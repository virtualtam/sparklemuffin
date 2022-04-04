package bookmark

// Repository provides access to user bookmarks.
type Repository interface {
	// BookmarkAdd adds a new bookmark for the logged in user.
	BookmarkAdd(bookmark Bookmark) error

	// BookmarkDelete deletes a given bookmark for the logged in user.
	BookmarkDelete(userUUID, uid string) error

	// BookmarkGetAll returns all bookmarks for the logged in user.
	BookmarkGetAll(userUUID string) ([]Bookmark, error)

	// BookmarkGetByURL returns the bookmark for a given URL for the logged in
	// user.
	BookmarkGetByURL(userUUID, url string) (Bookmark, error)

	// BookmarkUpdate updates an existing bookmark for the logged in user.
	BookmarkUpdate(bookmark Bookmark) error
}
