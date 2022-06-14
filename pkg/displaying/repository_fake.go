package displaying

import (
	"sort"

	"github.com/virtualtam/yawbe/pkg/bookmark"
)

var _ Repository = &fakeRepository{}

type fakeRepository struct {
	bookmarks []bookmark.Bookmark
}

func (r *fakeRepository) BookmarkGetN(userUUID string, n int, offset int) ([]bookmark.Bookmark, error) {
	userBookmarks := []bookmark.Bookmark{}

	for _, b := range r.bookmarks {
		if b.UserUUID == userUUID {
			userBookmarks = append(userBookmarks, b)
		}
	}

	sort.Slice(userBookmarks, func(i, j int) bool {
		return userBookmarks[i].CreatedAt.After(userBookmarks[j].CreatedAt)
	})

	var nBookmarks int

	if n > len(userBookmarks[offset:]) {
		nBookmarks = len(userBookmarks[offset:])
	} else {
		nBookmarks = n
	}

	return userBookmarks[offset : offset+nBookmarks], nil
}

func (r *fakeRepository) BookmarkGetCount(userUUID string) (int, error) {
	var userBookmarkCount int

	for _, b := range r.bookmarks {
		if b.UserUUID == userUUID {
			userBookmarkCount++
		}
	}

	return userBookmarkCount, nil
}
