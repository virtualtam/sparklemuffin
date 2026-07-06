// Copyright VirtualTam 2022, 2026
// SPDX-License-Identifier: MIT

package querying

import (
	"context"
	"sort"
	"strings"

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

func (r *fakeRepository) BookmarkGetN(_ context.Context, userUUID string, visibility Visibility, n uint, offset uint) ([]bookmark.Bookmark, error) {
	var userBookmarks []bookmark.Bookmark

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

func (r *fakeRepository) BookmarkGetCount(_ context.Context, userUUID string, visibility Visibility) (uint, error) {
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

func (r *fakeRepository) BookmarkGetPublicByUID(_ context.Context, userUUID, uid string) (bookmark.Bookmark, error) {
	for _, b := range r.bookmarks {
		if b.UserUUID == userUUID && b.UID == uid && !b.Private {
			return b, nil
		}
	}

	return bookmark.Bookmark{}, bookmark.ErrNotFound
}

func bookmarkMatchesSearch(b bookmark.Bookmark, searchTerms string) bool {
	term := strings.ToLower(searchTerms)

	if strings.Contains(strings.ToLower(b.Title), term) {
		return true
	}
	if strings.Contains(strings.ToLower(b.URL), term) {
		return true
	}
	if strings.Contains(strings.ToLower(b.Description), term) {
		return true
	}
	for _, tag := range b.Tags {
		if strings.Contains(strings.ToLower(tag), term) {
			return true
		}
	}
	return false
}

func (r *fakeRepository) bookmarkSearchMatches(userUUID string, visibility Visibility, searchTerms string) []bookmark.Bookmark {
	var matches []bookmark.Bookmark

	for _, b := range r.bookmarks {
		if b.UserUUID != userUUID {
			continue
		}
		if !visibilityMatches(visibility, b.Private) {
			continue
		}
		if bookmarkMatchesSearch(b, searchTerms) {
			matches = append(matches, b)
		}
	}

	sort.Slice(matches, func(i, j int) bool {
		return matches[i].CreatedAt.After(matches[j].CreatedAt)
	})

	return matches
}

func (r *fakeRepository) BookmarkSearchCount(_ context.Context, userUUID string, visibility Visibility, searchTerms string) (uint, error) {
	return uint(len(r.bookmarkSearchMatches(userUUID, visibility, searchTerms))), nil
}

func (r *fakeRepository) BookmarkSearchN(_ context.Context, userUUID string, visibility Visibility, searchTerms string, n uint, offset uint) ([]bookmark.Bookmark, error) {
	matches := r.bookmarkSearchMatches(userUUID, visibility, searchTerms)

	if offset >= uint(len(matches)) {
		return []bookmark.Bookmark{}, nil
	}

	end := min(offset+n, uint(len(matches)))
	return matches[offset:end], nil
}

func (r *fakeRepository) OwnerGetByUUID(_ context.Context, userUUID string) (Owner, error) {
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

func (r *fakeRepository) aggregateTags(userUUID string, visibility Visibility) []Tag {
	counts := map[string]uint{}

	for _, b := range r.bookmarks {
		if b.UserUUID != userUUID {
			continue
		}
		if !visibilityMatches(visibility, b.Private) {
			continue
		}
		for _, tag := range b.Tags {
			counts[tag]++
		}
	}

	tags := make([]Tag, 0, len(counts))
	for name, count := range counts {
		tags = append(tags, NewTag(name, count))
	}

	sort.Slice(tags, func(i, j int) bool {
		if tags[i].Count != tags[j].Count {
			return tags[i].Count > tags[j].Count
		}
		return tags[i].Name < tags[j].Name
	})

	return tags
}

func (r *fakeRepository) BookmarkTagGetAll(_ context.Context, userUUID string, visibility Visibility) ([]Tag, error) {
	return r.aggregateTags(userUUID, visibility), nil
}

func (r *fakeRepository) BookmarkTagGetCount(_ context.Context, userUUID string, visibility Visibility) (uint, error) {
	return uint(len(r.aggregateTags(userUUID, visibility))), nil
}

func (r *fakeRepository) BookmarkTagGetN(_ context.Context, userUUID string, visibility Visibility, n uint, offset uint) ([]Tag, error) {
	tags := r.aggregateTags(userUUID, visibility)

	if offset >= uint(len(tags)) {
		return []Tag{}, nil
	}

	end := min(offset+n, uint(len(tags)))
	return tags[offset:end], nil
}

func (r *fakeRepository) filterTags(userUUID string, visibility Visibility, filterTerm string) []Tag {
	all := r.aggregateTags(userUUID, visibility)

	filtered := all[:0]
	for _, tag := range all {
		if strings.Contains(strings.ToLower(tag.Name), strings.ToLower(filterTerm)) {
			filtered = append(filtered, tag)
		}
	}

	return filtered
}

func (r *fakeRepository) BookmarkTagFilterCount(_ context.Context, userUUID string, visibility Visibility, filterTerm string) (uint, error) {
	return uint(len(r.filterTags(userUUID, visibility, filterTerm))), nil
}

func (r *fakeRepository) BookmarkTagFilterN(_ context.Context, userUUID string, visibility Visibility, filterTerm string, n uint, offset uint) ([]Tag, error) {
	tags := r.filterTags(userUUID, visibility, filterTerm)

	if offset >= uint(len(tags)) {
		return []Tag{}, nil
	}

	end := min(offset+n, uint(len(tags)))
	return tags[offset:end], nil
}
