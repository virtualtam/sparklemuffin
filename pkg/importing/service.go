package importing

import (
	"github.com/virtualtam/netscape-go/v2"
	"github.com/virtualtam/sparklemuffin/pkg/bookmark"
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

func (s *Service) bulkImport(bookmarks []bookmark.Bookmark, overwriteExisting bool) (Status, error) {
	status := Status{
		overwriteExisting: overwriteExisting,
	}

	filteredBookmarks := []bookmark.Bookmark{}
	uniqueURLs := map[string]bool{}

	for _, b := range bookmarks {
		if err := b.ValidateForAddition(s.vr); err != nil {
			status.Invalid++
			continue
		}

		if _, ok := uniqueURLs[b.URL]; ok {
			// the import data contains duplicate entries
			// this may be a result of the normalization operations, or due to
			// a source allowing duplicate bookmarks -whether it is by design or
			// not
			status.Invalid++
			continue
		}

		uniqueURLs[b.URL] = true
		filteredBookmarks = append(filteredBookmarks, b)

	}

	if len(filteredBookmarks) == 0 {
		return status, nil
	}

	var rowsAffected int64
	var err error

	if overwriteExisting {
		rowsAffected, err = s.r.BookmarkUpsertMany(filteredBookmarks)
	} else {
		rowsAffected, err = s.r.BookmarkAddMany(filteredBookmarks)
	}

	if err != nil {
		return Status{}, err
	}

	status.NewOrUpdated = int(rowsAffected)
	status.Skipped = len(filteredBookmarks) - status.NewOrUpdated

	return status, nil
}

// ImportFromNetscapeDocument performs a bulk import from a Netscape bookmark export.
//
// The import will ignore:
// - duplicate bookmarks for a given URL; only the first entry will be imported;
// - bookmarks with missing or invalid values for required fields, such as the Title and URL.
func (s *Service) ImportFromNetscapeDocument(userUUID string, document *netscape.Document, visibility Visibility, overwrite OnConflictStrategy) (Status, error) {
	var overwriteExisting bool

	switch overwrite {
	case OnConflictOverwrite:
		overwriteExisting = true
	case OnConflictKeepExisting:
	default:
		return Status{}, ErrOnConflictStrategyInvalid
	}

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

	return s.bulkImport(bookmarks, overwriteExisting)
}
