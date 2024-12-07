// Copyright (c) VirtualTam
// SPDX-License-Identifier: MIT

package querying

import (
	"github.com/virtualtam/sparklemuffin/internal/paginate"
	"github.com/virtualtam/sparklemuffin/pkg/feed"
)

const (
	entriesPerPage uint   = 20
	PageHeaderAll  string = "All"
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

type (
	getCountCallback              func() (uint, error)
	subscriptionEntryGetNCallback func(offset uint) ([]SubscribedFeedEntry, error)
)

func (s *Service) feedsByPage(
	userUUID string,
	number uint,
	getCount getCountCallback,
	subscriptionEntryGetN subscriptionEntryGetNCallback,
	pageHeader string,
	pageDescription string,
) (FeedPage, error) {
	if number < 1 {
		return FeedPage{}, ErrPageNumberOutOfBounds
	}

	entryCount, err := getCount()
	if err != nil {
		return FeedPage{}, err
	}

	totalPages := paginate.PageCount(entryCount, entriesPerPage)

	if number > totalPages {
		return FeedPage{}, ErrPageNumberOutOfBounds
	}

	categories, err := s.r.FeedSubscriptionCategoryGetAll(userUUID)
	if err != nil {
		return FeedPage{}, err
	}

	if len(categories) == 0 {
		// early return: nothing to display
		return NewFeedPage(1, 1, pageHeader, pageDescription, []SubscribedFeedsByCategory{}, []SubscribedFeedEntry{}), nil
	}

	offset := (number - 1) * entriesPerPage

	entries, err := subscriptionEntryGetN(offset)
	if err != nil {
		return FeedPage{}, err
	}

	return NewFeedPage(number, totalPages, pageHeader, pageDescription, categories, entries), nil
}

// FeedsByPage returns a Page containing a limited and offset number of feeds.
func (s *Service) FeedsByPage(userUUID string, number uint) (FeedPage, error) {
	getCountFn := func() (uint, error) {
		return s.r.FeedEntryGetCount(userUUID)
	}

	subscriptionEntryGetNFn := func(offset uint) ([]SubscribedFeedEntry, error) {
		return s.r.FeedSubscriptionEntryGetN(userUUID, entriesPerPage, offset)
	}

	return s.feedsByPage(userUUID, number, getCountFn, subscriptionEntryGetNFn, PageHeaderAll, "")
}

// FeedsByCategoryAndPage returns a Page containing a limited and offset number of feeds.
func (s *Service) FeedsByCategoryAndPage(userUUID string, category feed.Category, number uint) (FeedPage, error) {
	getCountFn := func() (uint, error) {
		return s.r.FeedEntryGetCountByCategory(userUUID, category.UUID)
	}

	subscriptionEntryGetNFn := func(offset uint) ([]SubscribedFeedEntry, error) {
		return s.r.FeedSubscriptionEntryGetNByCategory(userUUID, category.UUID, entriesPerPage, offset)
	}

	return s.feedsByPage(userUUID, number, getCountFn, subscriptionEntryGetNFn, category.Name, "")
}

// FeedsBySubscriptionAndPage returns a Page containing a limited and offset number of feeds.
func (s *Service) FeedsBySubscriptionAndPage(userUUID string, subscription feed.Subscription, number uint) (FeedPage, error) {
	getCountFn := func() (uint, error) {
		return s.r.FeedEntryGetCountBySubscription(userUUID, subscription.UUID)
	}

	subscriptionEntryGetNFn := func(offset uint) ([]SubscribedFeedEntry, error) {
		return s.r.FeedSubscriptionEntryGetNBySubscription(userUUID, subscription.UUID, entriesPerPage, offset)
	}

	feed, err := s.r.FeedGetByUUID(subscription.FeedUUID)
	if err != nil {
		return FeedPage{}, err
	}

	return s.feedsByPage(userUUID, number, getCountFn, subscriptionEntryGetNFn, feed.Title, feed.Description)
}

func (s *Service) feedsByQueryAndPage(
	userUUID string,
	query string,
	number uint,
	getCount getCountCallback,
	subscriptionEntryGetN subscriptionEntryGetNCallback,
	pageHeader string,
	pageDescription string,
) (FeedPage, error) {
	if number < 1 {
		return FeedPage{}, ErrPageNumberOutOfBounds
	}

	entryCount, err := getCount()
	if err != nil {
		return FeedPage{}, err
	}

	totalPages := paginate.PageCount(entryCount, entriesPerPage)

	if number > totalPages {
		return FeedPage{}, ErrPageNumberOutOfBounds
	}

	categories, err := s.r.FeedSubscriptionCategoryGetAll(userUUID)
	if err != nil {
		return FeedPage{}, err
	}

	if len(categories) == 0 {
		// early return: nothing to display
		return NewFeedSearchResultPage(query, 0, 1, 1, PageHeaderAll, pageDescription, []SubscribedFeedsByCategory{}, []SubscribedFeedEntry{}), nil
	}

	offset := (number - 1) * entriesPerPage

	entries, err := subscriptionEntryGetN(offset)
	if err != nil {
		return FeedPage{}, err
	}

	return NewFeedSearchResultPage(query, entryCount, number, totalPages, pageHeader, pageDescription, categories, entries), nil
}

func (s *Service) FeedsByQueryAndPage(userUUID string, query string, number uint) (FeedPage, error) {
	getCountFn := func() (uint, error) {
		return s.r.FeedEntryGetCountByQuery(userUUID, query)
	}

	subscriptionEntryGetNFn := func(offset uint) ([]SubscribedFeedEntry, error) {
		return s.r.FeedSubscriptionEntryGetNByQuery(userUUID, query, entriesPerPage, offset)
	}

	return s.feedsByQueryAndPage(userUUID, query, number, getCountFn, subscriptionEntryGetNFn, PageHeaderAll, "")
}

func (s *Service) FeedsByCategoryAndQueryAndPage(userUUID string, category feed.Category, query string, number uint) (FeedPage, error) {
	getCountFn := func() (uint, error) {
		return s.r.FeedEntryGetCountByCategoryAndQuery(userUUID, category.UUID, query)
	}

	subscriptionEntryGetNFn := func(offset uint) ([]SubscribedFeedEntry, error) {
		return s.r.FeedSubscriptionEntryGetNByCategoryAndQuery(userUUID, category.UUID, query, entriesPerPage, offset)
	}

	return s.feedsByQueryAndPage(userUUID, query, number, getCountFn, subscriptionEntryGetNFn, category.Name, "")
}
func (s *Service) FeedsBySubscriptionAndQueryAndPage(userUUID string, subscription feed.Subscription, query string, number uint) (FeedPage, error) {
	getCountFn := func() (uint, error) {
		return s.r.FeedEntryGetCountBySubscriptionAndQuery(userUUID, subscription.UUID, query)
	}

	subscriptionEntryGetNFn := func(offset uint) ([]SubscribedFeedEntry, error) {
		return s.r.FeedSubscriptionEntryGetNBySubscriptionAndQuery(userUUID, subscription.UUID, query, entriesPerPage, offset)
	}

	feed, err := s.r.FeedGetByUUID(subscription.FeedUUID)
	if err != nil {
		return FeedPage{}, err
	}

	return s.feedsByQueryAndPage(userUUID, query, number, getCountFn, subscriptionEntryGetNFn, feed.Title, feed.Description)
}

func (s *Service) SubscriptionByUUID(userUUID string, subscriptionUUID string) (Subscription, error) {
	return s.r.FeedQueryingSubscriptionByUUID(userUUID, subscriptionUUID)
}

func (s *Service) SubscriptionsByCategory(userUUID string) ([]SubscriptionsByCategory, error) {
	return s.r.FeedQueryingSubscriptionsByCategory(userUUID)
}
