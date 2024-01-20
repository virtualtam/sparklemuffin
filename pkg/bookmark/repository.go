// Copyright (c) VirtualTam
// SPDX-License-Identifier: MIT

package bookmark

// ValidationRepository provides methods for Bookmark validation.
type ValidationRepository interface {
	// BookmarkIsURLRegistered returns whether a user has already saved a
	// bookmark with a given URL.
	BookmarkIsURLRegistered(userUUID, url string) (bool, error)

	// BookmarkIsURLRegisteredToAnotherUID returns whether a user has already
	// saved a bookmark with a given URL and a different UID.
	BookmarkIsURLRegisteredToAnotherUID(userUUID, url, uid string) (bool, error)
}

// Repository provides access to user bookmarks.
type Repository interface {
	ValidationRepository

	// BookmarkAdd adds a new bookmark for the logged in user.
	BookmarkAdd(bookmark Bookmark) error

	// BookmarkDelete deletes a given bookmark for the logged in user.
	BookmarkDelete(userUUID, uid string) error

	// BookmarkGetAll returns all bookmarks for a given user UUID.
	BookmarkGetAll(userUUID string) ([]Bookmark, error)

	// BookmarkGetByTag returns all bookmarks for a given user UUID and tag.
	BookmarkGetByTag(userUUID string, tag string) ([]Bookmark, error)

	// BookmarkGetByUID returns the bookmark for a given user UUID and UID.
	BookmarkGetByUID(userUUID, uid string) (Bookmark, error)

	// BookmarkGetByURL returns the bookmark for a given user UUID and URL.
	BookmarkGetByURL(userUUID, u string) (Bookmark, error)

	// BookmarkTagUpdateMany updates a tag for collection of existing bookmarks.
	BookmarkTagUpdateMany(bookmarks []Bookmark) (int64, error)

	// BookmarkUpdate updates an existing bookmark for the logged in user.
	BookmarkUpdate(bookmark Bookmark) error
}
