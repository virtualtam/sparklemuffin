package querying

import "github.com/virtualtam/yawbe/pkg/bookmark"

const (
	bookmarksPerPage uint = 20
)

// Service handles oprtaions related to displaying and paginating bookmarks.
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
func (s *Service) BookmarksByPage(ownerUUID string, visibility Visibility, number uint) (Page, error) {
	owner, err := s.r.OwnerGetByUUID(ownerUUID)
	if err != nil {
		return Page{}, err
	}

	if number < 1 {
		return Page{}, ErrPageNumberOutOfBounds
	}

	bookmarkCount, err := s.r.BookmarkGetCount(ownerUUID, visibility)
	if err != nil {
		return Page{}, err
	}

	totalPages := pageCount(bookmarkCount, bookmarksPerPage)

	if number > totalPages {
		return Page{}, ErrPageNumberOutOfBounds
	}

	if bookmarkCount == 0 {
		// early return: nothing to display
		return NewPage(owner, 1, 1, []bookmark.Bookmark{}), nil
	}

	dbOffset := (number - 1) * bookmarksPerPage

	bookmarks, err := s.r.BookmarkGetN(ownerUUID, visibility, bookmarksPerPage, dbOffset)
	if err != nil {
		return Page{}, err
	}

	return NewPage(owner, number, totalPages, bookmarks), nil
}

// BookmarksBySearchQueryAndPage returns a SearchPage containing a limited and offset
// number of bookmarks for a given set of search terms.
func (s *Service) BookmarksBySearchQueryAndPage(ownerUUID string, visibility Visibility, searchTerms string, number uint) (Page, error) {
	owner, err := s.r.OwnerGetByUUID(ownerUUID)
	if err != nil {
		return Page{}, err
	}

	if number < 1 {
		return Page{}, ErrPageNumberOutOfBounds
	}

	bookmarkCount, err := s.r.BookmarkSearchCount(ownerUUID, visibility, searchTerms)
	if err != nil {
		return Page{}, err
	}

	totalPages := pageCount(bookmarkCount, bookmarksPerPage)

	if number > totalPages {
		return Page{}, ErrPageNumberOutOfBounds
	}

	if bookmarkCount == 0 {
		// early return: nothing to display
		return NewSearchResultPage(owner, searchTerms, 0, 1, 1, []bookmark.Bookmark{}), nil
	}

	dbOffset := (number - 1) * bookmarksPerPage

	bookmarks, err := s.r.BookmarkSearchN(ownerUUID, visibility, searchTerms, bookmarksPerPage, dbOffset)
	if err != nil {
		return Page{}, err
	}

	return NewSearchResultPage(owner, searchTerms, bookmarkCount, number, totalPages, bookmarks), nil
}

// BookmarkByUID returns a Page containing a single bookmark.
func (s *Service) PublicBookmarkByUID(ownerUUID string, uid string) (Page, error) {
	owner, err := s.r.OwnerGetByUUID(ownerUUID)
	if err != nil {
		return Page{}, err
	}

	b, err := s.r.BookmarkGetPublicByUID(owner.UUID, uid)
	if err == bookmark.ErrNotFound {
		return NewPage(owner, 1, 1, []bookmark.Bookmark{}), nil
	} else if err != nil {
		return Page{}, err
	}

	return NewPage(owner, 1, 1, []bookmark.Bookmark{b}), nil
}

// PublicBookmarksByPage returns a Page containing a limited and offset number of bookmarks.
func (s *Service) PublicBookmarksByPage(ownerUUID string, number uint) (Page, error) {
	return s.BookmarksByPage(ownerUUID, VisibilityPublic, number)
}

// PublicBookmarksBySearchQueryAndPage returns a SearchPage containing a limited and offset
// number of bookmarks for a given set of search terms.
func (s *Service) PublicBookmarksBySearchQueryAndPage(ownerUUID string, searchTerms string, number uint) (Page, error) {
	return s.BookmarksBySearchQueryAndPage(ownerUUID, VisibilityPublic, searchTerms, number)
}
