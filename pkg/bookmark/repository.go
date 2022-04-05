package bookmark

// Repository provides access to user bookmarks.
type Repository interface {
	// BookmarkAdd adds a new bookmark for the logged in user.
	BookmarkAdd(bookmark Bookmark) error

	// BookmarkDelete deletes a given bookmark for the logged in user.
	BookmarkDelete(userUUID, uid string) error

	// BookmarkGetAll returns all bookmarks for a given user UUID.
	BookmarkGetAll(userUUID string) ([]Bookmark, error)

	// BookmarkGetByUID returns the bookmark for a given user UUID and UID.
	BookmarkGetByUID(userUUID, uid string) (Bookmark, error)

	// BookmarkGetByURL returns the bookmark for a given user UUID and URL.
	BookmarkGetByURL(userUUID, url string) (Bookmark, error)

	// BookmarkIsURLRegistered returns whether a user has already saved a
	// bookmark with a given URL.
	BookmarkIsURLRegistered(userUUD, url string) (bool, error)

	// BookmarkUpdate updates an existing bookmark for the logged in user.
	BookmarkUpdate(bookmark Bookmark) error
}
