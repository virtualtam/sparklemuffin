package www

import (
	"fmt"
	"time"

	"github.com/gorilla/feeds"
	"github.com/virtualtam/yawbe/pkg/bookmark"
	"github.com/virtualtam/yawbe/pkg/querying"
)

// bookmarksToFeed initializes and returns a new feed for a list of bookmarks.
func bookmarksToFeed(owner querying.Owner, bookmarks []bookmark.Bookmark) (*feeds.Feed, error) {
	feedItems := []*feeds.Item{}

	for _, b := range bookmarks {
		feedItem := &feeds.Item{
			Id:    b.UID,
			Title: b.Title,
			Link: &feeds.Link{
				Href: b.URL,
			},
			Created: b.CreatedAt,
			Updated: b.UpdatedAt,
		}

		if b.Description != "" {
			htmlDescription, err := markdownToHTML(b.Description)
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
			Href: fmt.Sprintf("/u/%s/bookmarks", owner.NickName),
		},
		Author: &feeds.Author{
			Name: owner.DisplayName,
		},
		Created: now,
		Items:   feedItems,
	}

	return feed, nil
}
