// Copyright (c) VirtualTam
// SPDX-License-Identifier: MIT

package bookmark

import (
	"context"
	"slices"
)

var _ Repository = &FakeRepository{}

type FakeRepository struct {
	Bookmarks []Bookmark
}

func (r *FakeRepository) BookmarkAdd(_ context.Context, bookmark Bookmark) error {
	r.Bookmarks = append(r.Bookmarks, bookmark)
	return nil
}

func (r *FakeRepository) BookmarkDelete(_ context.Context, userUUID, uid string) error {
	for index, b := range r.Bookmarks {
		if b.UserUUID == userUUID && b.UID == uid {
			r.Bookmarks = slices.Delete(r.Bookmarks, index, index+1)
			return nil
		}
	}

	return ErrNotFound
}

func (r *FakeRepository) BookmarkGetAll(_ context.Context, userUUID string) ([]Bookmark, error) {
	var bookmarks []Bookmark

	for _, b := range r.Bookmarks {
		if b.UserUUID == userUUID {
			bookmarks = append(bookmarks, b)
		}
	}

	return bookmarks, nil
}

func (r *FakeRepository) BookmarkGetByTag(_ context.Context, userUUID string, tag string) ([]Bookmark, error) {
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

func (r *FakeRepository) BookmarkGetByUID(_ context.Context, userUUID, uid string) (Bookmark, error) {
	for _, b := range r.Bookmarks {
		if b.UserUUID == userUUID && b.UID == uid {
			return b, nil
		}
	}

	return Bookmark{}, ErrNotFound
}

func (r *FakeRepository) BookmarkGetByURL(_ context.Context, userUUID, u string) (Bookmark, error) {
	for _, b := range r.Bookmarks {
		if b.UserUUID == userUUID && b.URL == u {
			return b, nil
		}
	}

	return Bookmark{}, ErrNotFound
}

func (r *FakeRepository) BookmarkIsURLRegistered(_ context.Context, userUUID, url string) (bool, error) {
	for _, b := range r.Bookmarks {
		if b.UserUUID == userUUID && b.URL == url {
			return true, nil
		}
	}

	return false, nil
}

func (r *FakeRepository) BookmarkIsURLRegisteredToAnotherUID(_ context.Context, userUUID, url, uid string) (bool, error) {
	for _, b := range r.Bookmarks {
		if b.UserUUID == userUUID && b.URL == url && b.UID != uid {
			return true, nil
		}
	}

	return false, nil
}

func (r *FakeRepository) BookmarkTagUpdateMany(_ context.Context, bookmarks []Bookmark) (int64, error) {
	for _, bookmark := range bookmarks {
		for index, b := range r.Bookmarks {
			if b.UserUUID == bookmark.UserUUID && b.UID == bookmark.UID {
				r.Bookmarks[index] = bookmark
			}
		}
	}

	return int64(len(bookmarks)), nil
}

func (r *FakeRepository) BookmarkUpdate(_ context.Context, bookmark Bookmark) error {
	for index, b := range r.Bookmarks {
		if b.UserUUID == bookmark.UserUUID && b.UID == bookmark.UID {
			r.Bookmarks[index] = bookmark
			return nil
		}
	}

	return ErrNotFound
}
