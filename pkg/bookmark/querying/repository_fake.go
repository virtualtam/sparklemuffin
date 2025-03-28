// Copyright (c) VirtualTam
// SPDX-License-Identifier: MIT

package querying

import (
	"errors"
	"sort"

	"github.com/virtualtam/sparklemuffin/pkg/bookmark"
	"github.com/virtualtam/sparklemuffin/pkg/user"
)

var _ Repository = &fakeRepository{}

type fakeRepository struct {
	bookmarks []bookmark.Bookmark
	users     []user.User
}

func visibilityMatches(visibility Visibility, private bool) bool {
	switch visibility {
	case VisibilityPrivate:
		return private
	case VisibilityPublic:
		return !private
	default:
		return true
	}
}

func (r *fakeRepository) BookmarkGetN(userUUID string, visibility Visibility, n uint, offset uint) ([]bookmark.Bookmark, error) {
	userBookmarks := []bookmark.Bookmark{}

	for _, b := range r.bookmarks {
		if b.UserUUID == userUUID {
			if !visibilityMatches(visibility, b.Private) {
				continue
			}
			userBookmarks = append(userBookmarks, b)
		}
	}

	sort.Slice(userBookmarks, func(i, j int) bool {
		return userBookmarks[i].CreatedAt.After(userBookmarks[j].CreatedAt)
	})

	nBookmarks := min(n, uint(len(userBookmarks[offset:])))

	return userBookmarks[offset : offset+nBookmarks], nil
}

func (r *fakeRepository) BookmarkGetCount(userUUID string, visibility Visibility) (uint, error) {
	var userBookmarkCount uint

	for _, b := range r.bookmarks {
		if b.UserUUID == userUUID {
			if !visibilityMatches(visibility, b.Private) {
				continue
			}
			userBookmarkCount++
		}
	}

	return userBookmarkCount, nil
}

func (r *fakeRepository) BookmarkGetPublicByUID(userUUID, uid string) (bookmark.Bookmark, error) {
	for _, b := range r.bookmarks {
		if b.UserUUID == userUUID && b.UID == uid && !b.Private {
			return b, nil
		}
	}

	return bookmark.Bookmark{}, bookmark.ErrNotFound
}

func (r *fakeRepository) BookmarkSearchCount(userUUID string, visibility Visibility, searchTerms string) (uint, error) {
	return 0, errors.New("not implemented")
}

func (r *fakeRepository) BookmarkSearchN(userUUID string, visibility Visibility, searchTerms string, n uint, offset uint) ([]bookmark.Bookmark, error) {
	return []bookmark.Bookmark{}, errors.New("not implemented")
}

func (r *fakeRepository) OwnerGetByUUID(userUUID string) (Owner, error) {
	for _, u := range r.users {
		if u.UUID == userUUID {
			owner := Owner{
				UUID:        u.UUID,
				NickName:    u.NickName,
				DisplayName: u.DisplayName,
			}
			return owner, nil
		}
	}

	return Owner{}, ErrOwnerNotFound
}

func (r *fakeRepository) BookmarkTagGetAll(userUUID string, visibility Visibility) ([]Tag, error) {
	return []Tag{}, errors.New("not implemented")
}

func (r *fakeRepository) BookmarkTagGetCount(userUUID string, visibility Visibility) (uint, error) {
	return 0, errors.New("not implemented")
}

func (r *fakeRepository) BookmarkTagGetN(userUUID string, visibility Visibility, n uint, offset uint) ([]Tag, error) {
	return []Tag{}, errors.New("not implemented")
}

func (r *fakeRepository) BookmarkTagFilterCount(userUUID string, visibility Visibility, searchTerms string) (uint, error) {
	return 0, errors.New("not implemented")
}

func (r *fakeRepository) BookmarkTagFilterN(userUUID string, visibility Visibility, searchTerms string, n uint, offset uint) ([]Tag, error) {
	return []Tag{}, errors.New("not implemented")
}
