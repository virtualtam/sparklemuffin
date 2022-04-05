package bookmark

var _ Repository = &FakeRepository{}

type FakeRepository struct {
	Bookmarks []Bookmark
}

func (r *FakeRepository) BookmarkAdd(bookmark Bookmark) error {
	r.Bookmarks = append(r.Bookmarks, bookmark)
	return nil
}

func (r *FakeRepository) BookmarkDelete(userUUID, uid string) error {
	for index, b := range r.Bookmarks {
		if b.UserUUID == userUUID && b.UID == uid {
			r.Bookmarks = append(r.Bookmarks[:index], r.Bookmarks[index+1:]...)
			return nil
		}
	}

	return ErrNotFound
}

func (r *FakeRepository) BookmarkGetAll(userUUID string) ([]Bookmark, error) {
	bookmarks := []Bookmark{}

	for _, b := range r.Bookmarks {
		if b.UserUUID == userUUID {
			bookmarks = append(bookmarks, b)
		}
	}

	return bookmarks, nil
}

func (r *FakeRepository) BookmarkGetByURL(userUUID, url string) (Bookmark, error) {
	for _, b := range r.Bookmarks {
		if b.UserUUID == userUUID && b.URL == url {
			return b, nil
		}
	}

	return Bookmark{}, ErrNotFound
}

func (r *FakeRepository) BookmarkGetByUID(userUUID, uid string) (Bookmark, error) {
	for _, b := range r.Bookmarks {
		if b.UserUUID == userUUID && b.UID == uid {
			return b, nil
		}
	}

	return Bookmark{}, ErrNotFound
}

func (r *FakeRepository) BookmarkIsURLRegistered(userUUID, url string) (bool, error) {
	for _, b := range r.Bookmarks {
		if b.UserUUID == userUUID && b.URL == url {
			return true, nil
		}
	}

	return false, nil
}

func (r *FakeRepository) BookmarkUpdate(bookmark Bookmark) error {
	for index, b := range r.Bookmarks {
		if b.UserUUID == bookmark.UserUUID && b.UID == bookmark.UID {
			r.Bookmarks[index] = bookmark
			return nil
		}
	}

	return ErrNotFound
}
