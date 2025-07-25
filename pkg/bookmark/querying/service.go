// Copyright (c) VirtualTam
// SPDX-License-Identifier: MIT

package querying

import (
	"errors"

	"github.com/virtualtam/sparklemuffin/internal/paginate"
	"github.com/virtualtam/sparklemuffin/pkg/bookmark"
)

const (
	bookmarksPerPage uint = 20
	tagsPerPage      uint = 90
)

// Service handles operations related to displaying and paginating bookmarks.
type Service struct {
	r Repository
}

// NewService initializes and returns a new Service.
func NewService(r Repository) *Service {
	return &Service{
		r: r,
	}
}

// BookmarksByPage returns a Page containing a limited and offset number of bookmarks.
func (s *Service) BookmarksByPage(ownerUUID string, visibility Visibility, number uint) (BookmarkPage, error) {
	owner, err := s.r.OwnerGetByUUID(ownerUUID)
	if err != nil {
		return BookmarkPage{}, err
	}

	if number < 1 {
		return BookmarkPage{}, paginate.ErrPageNumberOutOfBounds
	}

	bookmarkCount, err := s.r.BookmarkGetCount(ownerUUID, visibility)
	if err != nil {
		return BookmarkPage{}, err
	}

	totalPages := paginate.PageCount(bookmarkCount, bookmarksPerPage)

	if number > totalPages {
		return BookmarkPage{}, paginate.ErrPageNumberOutOfBounds
	}

	if bookmarkCount == 0 {
		// early return: nothing to display
		return NewBookmarkPage(owner, 1, 1, 0, []bookmark.Bookmark{}), nil
	}

	dbOffset := (number - 1) * bookmarksPerPage

	bookmarks, err := s.r.BookmarkGetN(ownerUUID, visibility, bookmarksPerPage, dbOffset)
	if err != nil {
		return BookmarkPage{}, err
	}

	return NewBookmarkPage(owner, number, totalPages, bookmarkCount, bookmarks), nil
}

// BookmarksBySearchQueryAndPage returns a SearchPage containing a limited and offset
// number of bookmarks for a given set of search terms.
func (s *Service) BookmarksBySearchQueryAndPage(ownerUUID string, visibility Visibility, searchTerms string, number uint) (BookmarkPage, error) {
	owner, err := s.r.OwnerGetByUUID(ownerUUID)
	if err != nil {
		return BookmarkPage{}, err
	}

	if number < 1 {
		return BookmarkPage{}, paginate.ErrPageNumberOutOfBounds
	}

	bookmarkCount, err := s.r.BookmarkSearchCount(ownerUUID, visibility, searchTerms)
	if err != nil {
		return BookmarkPage{}, err
	}

	totalPages := paginate.PageCount(bookmarkCount, bookmarksPerPage)

	if number > totalPages {
		return BookmarkPage{}, paginate.ErrPageNumberOutOfBounds
	}

	if bookmarkCount == 0 {
		// early return: nothing to display
		return NewBookmarkSearchResultPage(owner, searchTerms, 0, 1, 1, []bookmark.Bookmark{}), nil
	}

	dbOffset := (number - 1) * bookmarksPerPage

	bookmarks, err := s.r.BookmarkSearchN(ownerUUID, visibility, searchTerms, bookmarksPerPage, dbOffset)
	if err != nil {
		return BookmarkPage{}, err
	}

	return NewBookmarkSearchResultPage(owner, searchTerms, bookmarkCount, number, totalPages, bookmarks), nil
}

// BookmarkByUID returns a Page containing a single bookmark.
func (s *Service) PublicBookmarkByUID(ownerUUID string, uid string) (BookmarkPage, error) {
	owner, err := s.r.OwnerGetByUUID(ownerUUID)
	if err != nil {
		return BookmarkPage{}, err
	}

	b, err := s.r.BookmarkGetPublicByUID(owner.UUID, uid)
	if errors.Is(err, bookmark.ErrNotFound) {
		return NewBookmarkPage(owner, 1, 1, 0, []bookmark.Bookmark{}), nil
	} else if err != nil {
		return BookmarkPage{}, err
	}

	return NewBookmarkPage(owner, 1, 1, 1, []bookmark.Bookmark{b}), nil
}

// PublicBookmarksByPage returns a Page containing a limited and offset number of bookmarks.
func (s *Service) PublicBookmarksByPage(ownerUUID string, number uint) (BookmarkPage, error) {
	return s.BookmarksByPage(ownerUUID, VisibilityPublic, number)
}

// PublicBookmarksBySearchQueryAndPage returns a SearchPage containing a limited and offset
// number of bookmarks for a given set of search terms.
func (s *Service) PublicBookmarksBySearchQueryAndPage(ownerUUID string, searchTerms string, number uint) (BookmarkPage, error) {
	return s.BookmarksBySearchQueryAndPage(ownerUUID, VisibilityPublic, searchTerms, number)
}

// Tags return all tags for a given user.
func (s *Service) Tags(userUUID string, visibility Visibility) ([]Tag, error) {
	return s.r.BookmarkTagGetAll(userUUID, visibility)
}

// TagNamesByCount returns all tag names for a given user,
// sorted by count in descending order.
func (s *Service) TagNamesByCount(userUUID string, visibility Visibility) ([]string, error) {
	tags, err := s.r.BookmarkTagGetAll(userUUID, visibility)
	if err != nil {
		return []string{}, err
	}

	tagNames := make([]string, len(tags))

	for i, tag := range tags {
		tagNames[i] = tag.Name
	}

	return tagNames, nil
}

// TagsByPage returns a Page containing a limited and offset number of tags.
func (s *Service) TagsByPage(ownerUUID string, visibility Visibility, number uint) (TagPage, error) {
	if number < 1 {
		return TagPage{}, paginate.ErrPageNumberOutOfBounds
	}

	tagCount, err := s.r.BookmarkTagGetCount(ownerUUID, visibility)
	if err != nil {
		return TagPage{}, err
	}

	totalPages := paginate.PageCount(tagCount, tagsPerPage)

	if number > totalPages {
		return TagPage{}, paginate.ErrPageNumberOutOfBounds
	}

	if tagCount == 0 {
		// early return: nothing to display
		return NewTagPage(1, 1, 0, []Tag{}), nil
	}

	dbOffset := (number - 1) * tagsPerPage

	tags, err := s.r.BookmarkTagGetN(ownerUUID, visibility, tagsPerPage, dbOffset)
	if err != nil {
		return TagPage{}, err
	}

	return NewTagPage(number, totalPages, tagCount, tags), nil
}

// TagsByFilterQueryAndPage returns a TagSearchPage containing a limited and offset
// number of tags for a given filter term.
func (s *Service) TagsByFilterQueryAndPage(ownerUUID string, visibility Visibility, filterTerm string, number uint) (TagPage, error) {
	if number < 1 {
		return TagPage{}, paginate.ErrPageNumberOutOfBounds
	}

	tagCount, err := s.r.BookmarkTagFilterCount(ownerUUID, visibility, filterTerm)
	if err != nil {
		return TagPage{}, err
	}

	totalPages := paginate.PageCount(tagCount, tagsPerPage)

	if number > totalPages {
		return TagPage{}, paginate.ErrPageNumberOutOfBounds
	}

	if tagCount == 0 {
		// early return: nothing to display
		return NewTagFilterResultPage(filterTerm, 0, 1, 1, []Tag{}), nil
	}

	dbOffset := (number - 1) * tagsPerPage

	tags, err := s.r.BookmarkTagFilterN(ownerUUID, visibility, filterTerm, tagsPerPage, dbOffset)
	if err != nil {
		return TagPage{}, err
	}

	return NewTagFilterResultPage(filterTerm, tagCount, number, totalPages, tags), nil
}
