package importing

import (
	"github.com/virtualtam/netscape-go/v2"
	"github.com/virtualtam/yawbe/pkg/bookmark"
)

// Service handles bookmark import operations.
type Service struct {
	r  Repository
	vr bookmark.ValidationRepository
}

// NewService initializes and returns a new Service.
func NewService(r Repository) *Service {
	return &Service{
		r:  r,
		vr: &validationRepository{},
	}
}

func (s *Service) bulkImport(bookmarks []bookmark.Bookmark) (Status, error) {
	var status Status

	filteredBookmarks := []bookmark.Bookmark{}

	for _, b := range bookmarks {
		err := b.ValidateForAddition(s.vr)

		if err == nil {
			filteredBookmarks = append(filteredBookmarks, b)
			continue
		}

		status.Invalid++
	}

	if len(filteredBookmarks) == 0 {
		return status, nil
	}

	rowsAffected, err := s.r.BookmarkAddMany(filteredBookmarks)
	if err != nil {
		return Status{}, err
	}

	status.New = int(rowsAffected)
	status.Skipped = len(filteredBookmarks) - status.New

	return status, nil
}

func (s *Service) ImportFromNetscapeDocument(userUUID string, document *netscape.Document, visibility Visibility) (Status, error) {
	bookmarks := []bookmark.Bookmark{}

	flattenedDocument := document.Flatten()

	for _, netscapeBookmark := range flattenedDocument.Root.Bookmarks {
		bookmark := bookmark.NewBookmark(userUUID)

		bookmark.URL = netscapeBookmark.URL
		bookmark.Title = netscapeBookmark.Title
		bookmark.Description = netscapeBookmark.Description
		bookmark.Tags = netscapeBookmark.Tags

		switch visibility {
		case VisibilityDefault:
			bookmark.Private = netscapeBookmark.Private
		case VisibilityPrivate:
			bookmark.Private = true
		case VisibilityPublic:
			bookmark.Private = false
		default:
			return Status{}, ErrVisibilityInvalid
		}

		if !netscapeBookmark.CreatedAt.IsZero() {
			bookmark.CreatedAt = netscapeBookmark.CreatedAt
		}

		if !netscapeBookmark.UpdatedAt.IsZero() {
			bookmark.UpdatedAt = netscapeBookmark.UpdatedAt
		} else {
			bookmark.UpdatedAt = bookmark.CreatedAt
		}

		bookmark.Normalize()

		bookmarks = append(bookmarks, *bookmark)
	}

	return s.bulkImport(bookmarks)
}
