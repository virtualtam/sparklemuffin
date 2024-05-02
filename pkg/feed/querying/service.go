// Copyright (c) VirtualTam
// SPDX-License-Identifier: MIT

package querying

import "github.com/virtualtam/sparklemuffin/internal/paginate"

const (
	entriesPerPage uint = 20
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

	subscriptionEntryCount, err := s.r.FeedSubscriptionEntryGetCount(userUUID)
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
		return NewFeedPage(1, 1, []SubscriptionCategory{}, []SubscriptionEntry{}), nil
	}

	offset := (number - 1) * entriesPerPage

	entries, err := s.r.FeedSubscriptionEntryGetN(userUUID, entriesPerPage, offset)
	if err != nil {
		return FeedPage{}, err
	}

	return NewFeedPage(number, totalPages, categories, entries), nil
}

// FeedsByCategoryAndPage returns a Page containing a limited and offset number of feeds.
func (s *Service) FeedsByCategoryAndPage(userUUID string, categoryUUID string, number uint) (FeedPage, error) {
	if number < 1 {
		return FeedPage{}, ErrPageNumberOutOfBounds
	}

	subscriptionEntryCount, err := s.r.FeedSubscriptionEntryGetCountByCategory(userUUID, categoryUUID)
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
		return NewFeedPage(1, 1, []SubscriptionCategory{}, []SubscriptionEntry{}), nil
	}

	offset := (number - 1) * entriesPerPage

	entries, err := s.r.FeedSubscriptionEntryGetNByCategory(userUUID, categoryUUID, entriesPerPage, offset)
	if err != nil {
		return FeedPage{}, err
	}

	return NewFeedPage(number, totalPages, categories, entries), nil
}
