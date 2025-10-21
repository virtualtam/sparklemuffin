// Copyright (c) VirtualTam
// SPDX-License-Identifier: MIT

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

	var filteredBookmarks []bookmark.Bookmark
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

	var bookmarks []bookmark.Bookmark

	flattenedDocument := document.Flatten()

	for _, netscapeBookmark := range flattenedDocument.Root.Bookmarks {
		newBookmark := bookmark.NewBookmark(userUUID)

		newBookmark.URL = netscapeBookmark.URL
		newBookmark.Title = netscapeBookmark.Title
		newBookmark.Description = netscapeBookmark.Description
		newBookmark.Tags = netscapeBookmark.Tags

		switch visibility {
		case VisibilityDefault:
			newBookmark.Private = netscapeBookmark.Private
		case VisibilityPrivate:
			newBookmark.Private = true
		case VisibilityPublic:
			newBookmark.Private = false
		default:
			return Status{}, ErrVisibilityInvalid
		}

		if !netscapeBookmark.CreatedAt.IsZero() {
			newBookmark.CreatedAt = netscapeBookmark.CreatedAt
		}

		if !netscapeBookmark.UpdatedAt.IsZero() {
			newBookmark.UpdatedAt = netscapeBookmark.UpdatedAt
		} else {
			newBookmark.UpdatedAt = newBookmark.CreatedAt
		}

		newBookmark.Normalize()

		bookmarks = append(bookmarks, *newBookmark)
	}

	return s.bulkImport(bookmarks, overwriteExisting)
}
