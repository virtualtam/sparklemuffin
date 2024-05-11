// Copyright (c) VirtualTam
// SPDX-License-Identifier: MIT

package querying

import (
	"github.com/virtualtam/sparklemuffin/internal/paginate"
	"github.com/virtualtam/sparklemuffin/pkg/feed"
)

const (
	entriesPerPage uint   = 20
	pageHeaderAll  string = "All"
)

// Service handles oprtaions related to displaying and paginating feeds.
type Service struct {
	r Repository
}

// NewService initializes and returns a new Service.
func NewService(r Repository) *Service {
	return &Service{
		r: r,
	}
}

// FeedsByPage returns a Page containing a limited and offset number of feeds.
func (s *Service) FeedsByPage(userUUID string, number uint) (FeedPage, error) {
	if number < 1 {
		return FeedPage{}, ErrPageNumberOutOfBounds
	}

	subscriptionEntryCount, err := s.r.FeedEntryGetCount(userUUID)
	if err != nil {
		return FeedPage{}, err
	}

	totalPages := paginate.PageCount(subscriptionEntryCount, entriesPerPage)

	if number > totalPages {
		return FeedPage{}, ErrPageNumberOutOfBounds
	}

	categories, err := s.r.FeedSubscriptionCategoryGetAll(userUUID)
	if err != nil {
		return FeedPage{}, err
	}

	if len(categories) == 0 {
		// early return: nothing to display
		return NewFeedPage(1, 1, pageHeaderAll, []SubscriptionCategory{}, []SubscriptionEntry{}), nil
	}

	offset := (number - 1) * entriesPerPage

	entries, err := s.r.FeedSubscriptionEntryGetN(userUUID, entriesPerPage, offset)
	if err != nil {
		return FeedPage{}, err
	}

	return NewFeedPage(number, totalPages, pageHeaderAll, categories, entries), nil
}

// FeedsByCategoryAndPage returns a Page containing a limited and offset number of feeds.
func (s *Service) FeedsByCategoryAndPage(userUUID string, category feed.Category, number uint) (FeedPage, error) {
	if number < 1 {
		return FeedPage{}, ErrPageNumberOutOfBounds
	}

	subscriptionEntryCount, err := s.r.FeedEntryGetCountByCategory(userUUID, category.UUID)
	if err != nil {
		return FeedPage{}, err
	}

	totalPages := paginate.PageCount(subscriptionEntryCount, entriesPerPage)

	if number > totalPages {
		return FeedPage{}, ErrPageNumberOutOfBounds
	}

	categories, err := s.r.FeedSubscriptionCategoryGetAll(userUUID)
	if err != nil {
		return FeedPage{}, err
	}

	if len(categories) == 0 {
		// early return: nothing to display
		return NewFeedPage(1, 1, category.Name, []SubscriptionCategory{}, []SubscriptionEntry{}), nil
	}

	offset := (number - 1) * entriesPerPage

	entries, err := s.r.FeedSubscriptionEntryGetNByCategory(userUUID, category.UUID, entriesPerPage, offset)
	if err != nil {
		return FeedPage{}, err
	}

	return NewFeedPage(number, totalPages, category.Name, categories, entries), nil
}

// FeedsBySubscriptionAndPage returns a Page containing a limited and offset number of feeds.
func (s *Service) FeedsBySubscriptionAndPage(userUUID string, subscription feed.Subscription, number uint) (FeedPage, error) {
	if number < 1 {
		return FeedPage{}, ErrPageNumberOutOfBounds
	}

	subscriptionEntryCount, err := s.r.FeedEntryGetCountBySubscription(userUUID, subscription.UUID)
	if err != nil {
		return FeedPage{}, err
	}

	totalPages := paginate.PageCount(subscriptionEntryCount, entriesPerPage)

	if number > totalPages {
		return FeedPage{}, ErrPageNumberOutOfBounds
	}

	categories, err := s.r.FeedSubscriptionCategoryGetAll(userUUID)
	if err != nil {
		return FeedPage{}, err
	}

	feed, err := s.r.FeedGetByUUID(subscription.FeedUUID)
	if err != nil {
		return FeedPage{}, err
	}

	if len(categories) == 0 {
		// early return: nothing to display
		return NewFeedPage(1, 1, feed.Title, []SubscriptionCategory{}, []SubscriptionEntry{}), nil
	}

	offset := (number - 1) * entriesPerPage

	entries, err := s.r.FeedSubscriptionEntryGetNBySubscription(userUUID, subscription.UUID, entriesPerPage, offset)
	if err != nil {
		return FeedPage{}, err
	}

	return NewFeedPage(number, totalPages, feed.Title, categories, entries), nil
}
