package importing

import "github.com/virtualtam/yawbe/pkg/bookmark"

var _ Repository = &FakeRepository{}

type FakeRepository struct {
	Bookmarks []bookmark.Bookmark
}

func (r *FakeRepository) BookmarkAddMany(bookmarks []bookmark.Bookmark) (int64, error) {
	r.Bookmarks = append(r.Bookmarks, bookmarks...)
	return int64(len(bookmarks)), nil
}

func (r *FakeRepository) BookmarkIsURLRegistered(userUUID, url string) (bool, error) {
	for _, b := range r.Bookmarks {
		if b.UserUUID == userUUID && b.URL == url {
			return true, nil
		}
	}

	return false, nil
}

func (r *FakeRepository) BookmarkIsURLRegisteredToAnotherUID(userUUID, url, uid string) (bool, error) {
	for _, b := range r.Bookmarks {
		if b.UserUUID == userUUID && b.URL == url && b.UID != uid {
			return true, nil
		}
	}

	return false, nil
}
