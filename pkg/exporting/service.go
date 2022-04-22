package exporting

import (
	"fmt"

	"github.com/virtualtam/netscape-go/v2"
	"github.com/virtualtam/yawbe/pkg/bookmark"
)

// Service handles bookmark export operations.
type Service struct {
	r Repository
}

// NewService initializes and returns a new Service.
func NewService(r Repository) *Service {
	return &Service{
		r: r,
	}
}

// ExportAsNetscapeDocument exports a given user's bookmarks matching the
// provided Visibility as a Netscape bookmark document.
func (s *Service) ExportAsNetscapeDocument(userUUID string, visibility Visibility) (*netscape.Document, error) {
	var bookmarks []bookmark.Bookmark
	var err error

	switch visibility {
	case VisibilityAll:
		bookmarks, err = s.r.BookmarkGetAll(userUUID)
	case VisibilityPrivate:
		bookmarks, err = s.r.BookmarkGetAllPrivate(userUUID)
	case VisibilityPublic:
		bookmarks, err = s.r.BookmarkGetAllPublic(userUUID)
	default:
		err = ErrVisibilityInvalid
	}

	if err != nil {
		return &netscape.Document{}, err
	}

	documentTitle := fmt.Sprintf("YAWBE export of %s bookmarks", visibility)
	document := &netscape.Document{
		Title: documentTitle,
		Root: netscape.Folder{
			Name: documentTitle,
		},
	}

	for _, b := range bookmarks {
		netscapeBookmark := netscape.Bookmark{
			CreatedAt:   b.CreatedAt,
			UpdatedAt:   b.UpdatedAt,
			Title:       b.Title,
			URL:         b.URL,
			Description: b.Description,
			Private:     b.Private,
			Tags:        b.Tags,
		}

		document.Root.Bookmarks = append(document.Root.Bookmarks, netscapeBookmark)
	}

	return document, nil
}
