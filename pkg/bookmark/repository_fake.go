// Copyright (c) VirtualTam
// SPDX-License-Identifier: MIT

package bookmark

import "slices"

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
			r.Bookmarks = slices.Delete(r.Bookmarks, index, index+1)
			return nil
		}
	}

	return ErrNotFound
}

func (r *FakeRepository) BookmarkGetAll(userUUID string) ([]Bookmark, error) {
	var bookmarks []Bookmark

	for _, b := range r.Bookmarks {
		if b.UserUUID == userUUID {
			bookmarks = append(bookmarks, b)
		}
	}

	return bookmarks, nil
}

func (r *FakeRepository) BookmarkGetByTag(userUUID string, tag string) ([]Bookmark, error) {
	var bookmarks []Bookmark

	for _, b := range r.Bookmarks {
		for _, bookmarkTag := range b.Tags {
			if bookmarkTag == tag {
				bookmarks = append(bookmarks, b)
			}
		}
	}

	return bookmarks, nil
}

func (r *FakeRepository) BookmarkGetByUID(userUUID, uid string) (Bookmark, error) {
	for _, b := range r.Bookmarks {
		if b.UserUUID == userUUID && b.UID == uid {
			return b, nil
		}
	}

	return Bookmark{}, ErrNotFound
}

func (r *FakeRepository) BookmarkGetByURL(userUUID, u string) (Bookmark, error) {
	for _, b := range r.Bookmarks {
		if b.UserUUID == userUUID && b.URL == u {
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

func (r *FakeRepository) BookmarkIsURLRegisteredToAnotherUID(userUUID, url, uid string) (bool, error) {
	for _, b := range r.Bookmarks {
		if b.UserUUID == userUUID && b.URL == url && b.UID != uid {
			return true, nil
		}
	}

	return false, nil
}

func (r *FakeRepository) BookmarkTagUpdateMany(bookmarks []Bookmark) (int64, error) {
	for _, bookmark := range bookmarks {
		for index, b := range r.Bookmarks {
			if b.UserUUID == bookmark.UserUUID && b.UID == bookmark.UID {
				r.Bookmarks[index] = bookmark
			}
		}
	}

	return int64(len(bookmarks)), nil
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
