// Copyright (c) VirtualTam
// SPDX-License-Identifier: MIT

package importing

import (
	"context"
	"errors"

	"github.com/rs/zerolog/log"
	"github.com/virtualtam/opml-go"

	"github.com/virtualtam/sparklemuffin/pkg/feed"
)

// Service handles feed subscription import operations.
type Service struct {
	*feed.Service
}

// NewService initializes and returns a new Service.
func NewService(feedService *feed.Service) *Service {
	return &Service{
		Service: feedService,
	}
}

// ImportFromOPMLDocument imports feed subscriptions and categories from an OPML document.
func (s *Service) ImportFromOPMLDocument(ctx context.Context, userUUID string, document *opml.Document) (Status, error) {
	var status Status
	var errs []error

	categoriesFeeds := opmlToCategoriesFeeds(document.Body.Outlines)

	for categoryName, feedURLs := range categoriesFeeds {
		category, created, err := s.GetOrCreateCategory(ctx, userUUID, categoryName)
		if err != nil {
			return Status{}, err
		}

		status.Categories.Inc(created)

		for _, feedURL := range feedURLs {
			newFeed, created, err := s.GetOrCreateFeedAndEntries(ctx, feedURL)
			if err != nil {
				log.
					Error().
					Err(err).
					Str("feed_url", feedURL).
					Msg("feed: failed to create feed")

				errs = append(errs, err)

				continue
			}

			status.Feeds.Inc(created)

			newSubscription := feed.Subscription{
				UserUUID:     userUUID,
				CategoryUUID: category.UUID,
				FeedUUID:     newFeed.UUID,
			}

			_, created, err = s.GetOrCreateSubscription(ctx, newSubscription)
			if err != nil {
				return Status{}, err
			}

			status.Subscriptions.Inc(created)
		}
	}

	return status, errors.Join(errs...)
}
