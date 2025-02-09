// Copyright (c) VirtualTam
// SPDX-License-Identifier: MIT

package querying

import (
	"encoding/base64"

	"github.com/virtualtam/sparklemuffin/pkg/bookmark"
)

// A BookmarkPage holds a set of paginated bookmarks.
type BookmarkPage struct {
	// Owner exposes public metadata for the User owning the bookmarks.
	Owner Owner

	PageNumber         uint
	PreviousPageNumber uint
	NextPageNumber     uint
	TotalPages         uint
	Offset             uint

	SearchTerms       string
	SearchResultCount uint

	Bookmarks []bookmark.Bookmark
}

// NewBookmarkPage initializes and returns a new BookmarkPage.
func NewBookmarkPage(owner Owner, number uint, totalPages uint, bookmarks []bookmark.Bookmark) BookmarkPage {
	page := BookmarkPage{
		Owner:      owner,
		PageNumber: number,
		TotalPages: totalPages,
		Bookmarks:  bookmarks,
	}

	if page.PageNumber == 1 {
		page.PreviousPageNumber = 1
	} else {
		page.PreviousPageNumber = page.PageNumber - 1
	}

	if page.PageNumber == page.TotalPages {
		page.NextPageNumber = page.PageNumber
	} else {
		page.NextPageNumber = page.PageNumber + 1
	}

	page.Offset = (page.PageNumber-1)*bookmarksPerPage + 1

	return page
}

// NewBookmarkSearchResultPage initializes and returns a new BookmarkPage containing search results.
func NewBookmarkSearchResultPage(owner Owner, searchTerms string, searchResultCount uint, number uint, totalPages uint, bookmarks []bookmark.Bookmark) BookmarkPage {
	page := NewBookmarkPage(owner, number, totalPages, bookmarks)

	page.SearchTerms = searchTerms
	page.SearchResultCount = searchResultCount

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
	// Owner exposes public metadata for the User owning the bookmarks.
	Owner Owner

	PageNumber         uint
	PreviousPageNumber uint
	NextPageNumber     uint
	TotalPages         uint
	Offset             uint

	FilterTerm string

	TagCount uint
	Tags     []Tag
}

// NewTagPage initializes and returns a new TagPage.
func NewTagPage(owner Owner, number uint, totalPages uint, tagCount uint, tags []Tag) TagPage {
	page := TagPage{
		Owner:      owner,
		PageNumber: number,
		TotalPages: totalPages,
		TagCount:   tagCount,
		Tags:       tags,
	}

	if page.PageNumber == 1 {
		page.PreviousPageNumber = 1
	} else {
		page.PreviousPageNumber = page.PageNumber - 1
	}

	if page.PageNumber == page.TotalPages {
		page.NextPageNumber = page.PageNumber
	} else {
		page.NextPageNumber = page.PageNumber + 1
	}

	page.Offset = (page.PageNumber-1)*bookmarksPerPage + 1

	return page
}

// NewTagFilterResultPage initializes and returns a new bookmark Page containing filtered results.
func NewTagFilterResultPage(owner Owner, filterTerm string, tagCount uint, number uint, totalPages uint, tags []Tag) TagPage {
	page := NewTagPage(owner, number, totalPages, tagCount, tags)

	page.FilterTerm = filterTerm

	return page
}
