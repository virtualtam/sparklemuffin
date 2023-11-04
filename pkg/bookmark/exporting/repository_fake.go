package exporting

import "github.com/virtualtam/sparklemuffin/pkg/bookmark"

var _ Repository = &FakeRepository{}

type FakeRepository struct {
	Bookmarks []bookmark.Bookmark
}

func (r *FakeRepository) BookmarkGetAll(userUUID string) ([]bookmark.Bookmark, error) {
	bookmarks := []bookmark.Bookmark{}

	for _, b := range r.Bookmarks {
		if b.UserUUID == userUUID {
			bookmarks = append(bookmarks, b)
		}
	}

	return bookmarks, nil
}

func (r *FakeRepository) BookmarkGetAllPrivate(userUUID string) ([]bookmark.Bookmark, error) {
	bookmarks := []bookmark.Bookmark{}

	for _, b := range r.Bookmarks {
		if b.UserUUID == userUUID && b.Private {
			bookmarks = append(bookmarks, b)
		}
	}

	return bookmarks, nil
}

func (r *FakeRepository) BookmarkGetAllPublic(userUUID string) ([]bookmark.Bookmark, error) {
	bookmarks := []bookmark.Bookmark{}

	for _, b := range r.Bookmarks {
		if b.UserUUID == userUUID && !b.Private {
			bookmarks = append(bookmarks, b)
		}
	}

	return bookmarks, nil
}
