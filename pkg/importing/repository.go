package importing

import "github.com/virtualtam/yawbe/pkg/bookmark"

var _ bookmark.ValidationRepository = &validationRepository{}

type validationRepository struct{}

func (r *validationRepository) BookmarkIsURLRegistered(userUUID, url string) (bool, error) {
	// unicity checks for bulk operations must be handled by the persistence
	// layer
	return false, nil
}

func (r *validationRepository) BookmarkIsURLRegisteredToAnotherUID(userUUID, url, uid string) (bool, error) {
	// unicity checks for bulk operations must be handled by the persistence
	// layer
	return false, nil
}

type Repository interface {
	// BookmarkAddMany adds a collection of new bookmarks.
	BookmarkAddMany(bookmarks []bookmark.Bookmark) (int64, error)

	// BookmarkUpsertMany adds a collection of new bookmarks and updates
	// existing bookmarks in case of conflict.
	BookmarkUpsertMany(bookmarks []bookmark.Bookmark) (int64, error)
}
