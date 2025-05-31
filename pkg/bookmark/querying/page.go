// Copyright (c) VirtualTam
// SPDX-License-Identifier: MIT

package querying

import (
	"encoding/base64"

	"github.com/virtualtam/sparklemuffin/internal/paginate"
	"github.com/virtualtam/sparklemuffin/pkg/bookmark"
)

// A BookmarkPage holds a set of paginated bookmarks.
type BookmarkPage struct {
	paginate.Page

	// Owner exposes public metadata for the User owning the bookmarks.
	Owner Owner

	Bookmarks []bookmark.Bookmark
}

// NewBookmarkPage initializes and returns a new BookmarkPage.
func NewBookmarkPage(owner Owner, number uint, totalPages uint, totalBookmarkCount uint, bookmarks []bookmark.Bookmark) BookmarkPage {
	page := BookmarkPage{
		Page:      paginate.NewPage(number, totalPages, bookmarksPerPage, totalBookmarkCount),
		Owner:     owner,
		Bookmarks: bookmarks,
	}

	return page
}

// NewBookmarkSearchResultPage initializes and returns a new BookmarkPage containing search results.
func NewBookmarkSearchResultPage(owner Owner, searchTerms string, searchResultCount uint, number uint, totalPages uint, bookmarks []bookmark.Bookmark) BookmarkPage {
	page := NewBookmarkPage(owner, number, totalPages, searchResultCount, bookmarks)
	page.SearchTerms = searchTerms

	return page
}

// A Tag holds metadata for a given bookmark tag.
type Tag struct {
	Name        string
	EncodedName string
	Count       uint
}

// NewTag initializes and returns a new Tag.
func NewTag(name string, count uint) Tag {
	return Tag{
		Name:        name,
		EncodedName: base64.URLEncoding.EncodeToString([]byte(name)),
		Count:       count,
	}
}

// A TagPage holds a set of paginated bookmark tags.
type TagPage struct {
	paginate.Page

	Tags []Tag
}

// NewTagPage initializes and returns a new TagPage.
func NewTagPage(number uint, totalPages uint, tagCount uint, tags []Tag) TagPage {
	page := TagPage{
		Page: paginate.NewPage(number, totalPages, tagsPerPage, tagCount),
		Tags: tags,
	}

	return page
}

// NewTagFilterResultPage initializes and returns a new bookmark Page containing filtered results.
func NewTagFilterResultPage(searchTerms string, tagCount uint, number uint, totalPages uint, tags []Tag) TagPage {
	page := NewTagPage(number, totalPages, tagCount, tags)
	page.SearchTerms = searchTerms

	return page
}
