package querying

import "github.com/virtualtam/yawbe/pkg/bookmark"

const (
	bookmarksPerPage int = 20
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

// ByPage returns a Page containing a limited and offset number of bookmarks.
func (s *Service) ByPage(userUUID string, visibility Visibility, number int) (Page, error) {
	if number < 1 {
		return Page{}, ErrPageNumberOutOfBounds
	}

	bookmarkCount, err := s.r.BookmarkGetCount(userUUID, visibility)
	if err != nil {
		return Page{}, err
	}

	totalPages := pageCount(bookmarkCount, bookmarksPerPage)

	if number > totalPages {
		return Page{}, ErrPageNumberOutOfBounds
	}

	if bookmarkCount == 0 {
		// early return: nothing to display
		return NewPage(1, 1, []bookmark.Bookmark{}), nil
	}

	dbOffset := (number - 1) * bookmarksPerPage

	bookmarks, err := s.r.BookmarkGetN(userUUID, visibility, bookmarksPerPage, dbOffset)
	if err != nil {
		return Page{}, err
	}

	return NewPage(number, totalPages, bookmarks), nil
}

// BySearchQueryAndPage returns a SearchPage containing a limited and offset
// number of bookmarks for a given set of search terms.
func (s *Service) BySearchQueryAndPage(userUUID string, visibility Visibility, searchTerms string, number int) (Page, error) {
	if number < 1 {
		return Page{}, ErrPageNumberOutOfBounds
	}

	bookmarkCount, err := s.r.BookmarkSearchCount(userUUID, visibility, searchTerms)
	if err != nil {
		return Page{}, err
	}

	totalPages := pageCount(bookmarkCount, bookmarksPerPage)

	if number > totalPages {
		return Page{}, ErrPageNumberOutOfBounds
	}

	if bookmarkCount == 0 {
		// early return: nothing to display
		return NewSearchResultPage(searchTerms, 0, 1, 1, []bookmark.Bookmark{}), nil
	}

	dbOffset := (number - 1) * bookmarksPerPage

	bookmarks, err := s.r.BookmarkSearchN(userUUID, visibility, searchTerms, bookmarksPerPage, dbOffset)
	if err != nil {
		return Page{}, err
	}

	return NewSearchResultPage(searchTerms, bookmarkCount, number, totalPages, bookmarks), nil
}
