// Copyright (c) VirtualTam
// SPDX-License-Identifier: MIT

package exporting

import (
	"context"

	"github.com/virtualtam/sparklemuffin/pkg/bookmark"
)

var _ Repository = &FakeRepository{}

type FakeRepository struct {
	Bookmarks []bookmark.Bookmark
}

func (r *FakeRepository) BookmarkGetAll(_ context.Context, userUUID string) ([]bookmark.Bookmark, error) {
	var bookmarks []bookmark.Bookmark

	for _, b := range r.Bookmarks {
		if b.UserUUID == userUUID {
			bookmarks = append(bookmarks, b)
		}
	}

	return bookmarks, nil
}

func (r *FakeRepository) BookmarkGetAllPrivate(_ context.Context, userUUID string) ([]bookmark.Bookmark, error) {
	var bookmarks []bookmark.Bookmark

	for _, b := range r.Bookmarks {
		if b.UserUUID == userUUID && b.Private {
			bookmarks = append(bookmarks, b)
		}
	}

	return bookmarks, nil
}

func (r *FakeRepository) BookmarkGetAllPublic(_ context.Context, userUUID string) ([]bookmark.Bookmark, error) {
	var bookmarks []bookmark.Bookmark

	for _, b := range r.Bookmarks {
		if b.UserUUID == userUUID && !b.Private {
			bookmarks = append(bookmarks, b)
		}
	}

	return bookmarks, nil
}
