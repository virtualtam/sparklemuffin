// Copyright (c) VirtualTam
// SPDX-License-Identifier: MIT

package importing

import (
	"context"

	"github.com/virtualtam/sparklemuffin/pkg/bookmark"
)

var _ Repository = &FakeRepository{}

type FakeRepository struct {
	Bookmarks []bookmark.Bookmark
}

func (r *FakeRepository) BookmarkAddMany(_ context.Context, bookmarks []bookmark.Bookmark) (int64, error) {
	return r.bookmarkUpsertMany(bookmarks, false)
}

func (r *FakeRepository) BookmarkUpsertMany(_ context.Context, bookmarks []bookmark.Bookmark) (int64, error) {
	return r.bookmarkUpsertMany(bookmarks, true)
}

func (r *FakeRepository) bookmarkUpsertMany(bookmarks []bookmark.Bookmark, overwriteExisting bool) (int64, error) {
	uniqueURLs := map[string]int{}
	for index, b := range r.Bookmarks {
		uniqueURLs[b.URL] = index
	}

	var newOrUpdated int64

	for _, b := range bookmarks {
		if index, ok := uniqueURLs[b.URL]; ok {
			// bookmark already exists
			if overwriteExisting {
				r.Bookmarks[index] = b
				newOrUpdated++
			}

			continue
		}

		r.Bookmarks = append(r.Bookmarks, b)
		uniqueURLs[b.URL] = len(r.Bookmarks) - 1
		newOrUpdated++
	}

	return newOrUpdated, nil
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
