package controller

import (
	"fmt"
	"net/url"
	"time"

	"github.com/gorilla/feeds"

	"github.com/virtualtam/sparklemuffin/cmd/sparklemuffin/http/www/view"
	"github.com/virtualtam/sparklemuffin/pkg/bookmark"
	"github.com/virtualtam/sparklemuffin/pkg/bookmark/querying"
)

// bookmarksToFeed initializes and returns a new feed for a list of bookmarks.
func bookmarksToFeed(publicURL *url.URL, owner querying.Owner, bookmarks []bookmark.Bookmark) (*feeds.Feed, error) {
	feedItems := []*feeds.Item{}

	for _, b := range bookmarks {
		feedItem := &feeds.Item{
			Id:    fmt.Sprintf("%s/u/%s/bookmarks/%s", publicURL.String(), owner.NickName, b.UID),
			Title: b.Title,
			Link: &feeds.Link{
				Href: b.URL,
			},
			Created: b.CreatedAt,
			Updated: b.UpdatedAt,
		}

		if b.Description != "" {
			htmlDescription, err := view.MarkdownToHTML(b.Description)
			if err != nil {
				return &feeds.Feed{}, fmt.Errorf("failed to render Markdown description: %w", err)
			}

			feedItem.Content = htmlDescription
		}

		feedItems = append(feedItems, feedItem)
	}

	now := time.Now().UTC()

	feed := &feeds.Feed{
		Title: fmt.Sprintf("%s's bookmarks", owner.DisplayName),
		Link: &feeds.Link{
			Href: fmt.Sprintf("%s/u/%s/bookmarks", publicURL.String(), owner.NickName),
		},
		Author: &feeds.Author{
			Name: owner.DisplayName,
		},
		Created: now,
		Items:   feedItems,
	}

	return feed, nil
}
