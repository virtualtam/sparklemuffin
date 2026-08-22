// Copyright VirtualTam 2022, 2026
// SPDX-License-Identifier: MIT

package querying

import (
	"context"
	"errors"

	"github.com/virtualtam/sparklemuffin/internal/paginate"
	"github.com/virtualtam/sparklemuffin/pkg/bookmark"
)

const bookmarksPerPage uint = 20

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

// BookmarkCountsByTag returns the number of bookmarks for a given user,
// grouped by tag name.
func (s *Service) BookmarkCountsByTag(ctx context.Context, userUUID string) (map[string]uint, error) {
	return s.r.BookmarkGetCountsByTag(ctx, userUUID)
}

// BookmarksByPage returns a Page containing a limited and offset number of bookmarks.
func (s *Service) BookmarksByPage(ctx context.Context, ownerUUID string, visibility Visibility, number uint) (BookmarkPage, error) {
	owner, err := s.r.OwnerGetByUUID(ctx, ownerUUID)
	if err != nil {
		return BookmarkPage{}, err
	}

	if number < 1 {
		return BookmarkPage{}, paginate.ErrPageNumberOutOfBounds
	}

	bookmarkCount, err := s.r.BookmarkGetCount(ctx, ownerUUID, visibility)
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

	bookmarks, err := s.r.BookmarkGetN(ctx, ownerUUID, visibility, bookmarksPerPage, dbOffset)
	if err != nil {
		return BookmarkPage{}, err
	}

	return NewBookmarkPage(owner, number, totalPages, bookmarkCount, bookmarks), nil
}

// BookmarksBySearchQueryAndPage returns a SearchPage containing a limited and offset
// number of bookmarks for a given set of search terms.
func (s *Service) BookmarksBySearchQueryAndPage(ctx context.Context, ownerUUID string, visibility Visibility, searchTerms string, number uint) (BookmarkPage, error) {
	owner, err := s.r.OwnerGetByUUID(ctx, ownerUUID)
	if err != nil {
		return BookmarkPage{}, err
	}

	if number < 1 {
		return BookmarkPage{}, paginate.ErrPageNumberOutOfBounds
	}

	bookmarkCount, err := s.r.BookmarkSearchCount(ctx, ownerUUID, visibility, searchTerms)
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

	bookmarks, err := s.r.BookmarkSearchN(ctx, ownerUUID, visibility, searchTerms, bookmarksPerPage, dbOffset)
	if err != nil {
		return BookmarkPage{}, err
	}

	return NewBookmarkSearchResultPage(owner, searchTerms, bookmarkCount, number, totalPages, bookmarks), nil
}

// PublicBookmarkByUID returns a Page containing a single public bookmark.
func (s *Service) PublicBookmarkByUID(ctx context.Context, ownerUUID string, uid string) (BookmarkPage, error) {
	owner, err := s.r.OwnerGetByUUID(ctx, ownerUUID)
	if err != nil {
		return BookmarkPage{}, err
	}

	b, err := s.r.BookmarkGetPublicByUID(ctx, owner.UUID, uid)
	if errors.Is(err, bookmark.ErrNotFound) {
		return NewBookmarkPage(owner, 1, 1, 0, []bookmark.Bookmark{}), nil
	} else if err != nil {
		return BookmarkPage{}, err
	}

	return NewBookmarkPage(owner, 1, 1, 1, []bookmark.Bookmark{b}), nil
}

// PublicBookmarksByPage returns a Page containing a limited and offset number of public bookmarks.
func (s *Service) PublicBookmarksByPage(ctx context.Context, ownerUUID string, number uint) (BookmarkPage, error) {
	return s.BookmarksByPage(ctx, ownerUUID, VisibilityPublic, number)
}

// PublicBookmarksBySearchQueryAndPage returns a SearchPage containing a limited and offset
// number of bookmarks for a given set of search terms.
func (s *Service) PublicBookmarksBySearchQueryAndPage(ctx context.Context, ownerUUID string, searchTerms string, number uint) (BookmarkPage, error) {
	return s.BookmarksBySearchQueryAndPage(ctx, ownerUUID, VisibilityPublic, searchTerms, number)
}
