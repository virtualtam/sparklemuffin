// Copyright (c) VirtualTam
// SPDX-License-Identifier: MIT

package exporting

import (
	"fmt"
	"time"

	"github.com/virtualtam/netscape-go/v2"
	"github.com/virtualtam/sparklemuffin/pkg/bookmark"
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

func (s *Service) getBookmarks(userUUID string, visibility Visibility) ([]bookmark.Bookmark, error) {
	switch visibility {
	case VisibilityAll:
		return s.r.BookmarkGetAll(userUUID)
	case VisibilityPrivate:
		return s.r.BookmarkGetAllPrivate(userUUID)
	case VisibilityPublic:
		return s.r.BookmarkGetAllPublic(userUUID)
	default:
		return []bookmark.Bookmark{}, ErrVisibilityInvalid
	}
}

// ExportAsJSONDocument exports a given user's bookmarks matching the
// provided Visibility as a JSON bookmark document.
func (s *Service) ExportAsJSONDocument(userUUID string, visibility Visibility) (*jsonDocument, error) {
	bookmarks, err := s.getBookmarks(userUUID, visibility)
	if err != nil {
		return &jsonDocument{}, err
	}

	now := time.Now().UTC()

	document := &jsonDocument{
		Title:      fmt.Sprintf("SparkleMuffin export of %s bookmarks", visibility),
		ExportedAt: now,
	}

	for _, b := range bookmarks {
		jsonBookmark := jsonBookmark{
			URL:         b.URL,
			Title:       b.Title,
			Description: b.Description,
			Private:     b.Private,
			Tags:        b.Tags,
			CreatedAt:   b.CreatedAt,
			UpdatedAt:   b.UpdatedAt,
		}

		document.Bookmarks = append(document.Bookmarks, jsonBookmark)
	}

	return document, nil
}

// ExportAsNetscapeDocument exports a given user's bookmarks matching the
// provided Visibility as a Netscape bookmark document.
func (s *Service) ExportAsNetscapeDocument(userUUID string, visibility Visibility) (*netscape.Document, error) {
	bookmarks, err := s.getBookmarks(userUUID, visibility)
	if err != nil {
		return &netscape.Document{}, err
	}

	documentTitle := fmt.Sprintf("SparkleMuffin export of %s bookmarks", visibility)
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
