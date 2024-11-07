// Copyright (c) VirtualTam
// SPDX-License-Identifier: MIT

package exporting

import (
	"fmt"
	"time"

	"github.com/virtualtam/opml-go"

	"github.com/virtualtam/sparklemuffin/pkg/user"
)

// Service handles feed subscription export operations.
type Service struct {
	r Repository
}

// NewService initializes and returns a new Service.
func NewService(r Repository) *Service {
	return &Service{
		r: r,
	}
}

// ExportAsOPMLDocument exports a given user's feed subscriptions as an OPML document.
func (s *Service) ExportAsOPMLDocument(user user.User) (*opml.Document, error) {
	categorySubscriptions, err := s.r.FeedCategorySubscriptionsGetAll(user.UUID)
	if err != nil {
		return &opml.Document{}, err
	}

	var categoryOutlines []opml.Outline

	for _, category := range categorySubscriptions {
		var feedOutlines []opml.Outline

		for _, feed := range category.SubscribedFeeds {
			feedOutline := opml.Outline{
				Text:   feed.Title,
				Title:  feed.Title,
				Type:   opml.OutlineTypeSubscription,
				XmlUrl: feed.FeedURL,
			}

			feedOutlines = append(feedOutlines, feedOutline)
		}

		categoryOutline := opml.Outline{
			Text:     category.Name,
			Title:    category.Name,
			Outlines: feedOutlines,
		}

		categoryOutlines = append(categoryOutlines, categoryOutline)
	}

	documentTitle := fmt.Sprintf("%s's feed subscriptions on SparkleMuffin", user.DisplayName)
	now := time.Now()

	document := &opml.Document{
		Version: opml.Version2,
		Head: opml.Head{
			Title:       documentTitle,
			DateCreated: now,
		},
		Body: opml.Body{
			Outlines: categoryOutlines,
		},
	}

	return document, nil
}
