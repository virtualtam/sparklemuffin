// Copyright (c) VirtualTam
// SPDX-License-Identifier: MIT

package querying

import (
	"context"

	"github.com/virtualtam/sparklemuffin/internal/paginate"
	"github.com/virtualtam/sparklemuffin/pkg/feed"
)

const (
	entriesPerPage uint   = 20
	PageHeaderAll  string = "All"
)

// Service handles operations related to displaying and paginating feeds.
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
	getCountFn              func() (uint, error)
	subscriptionEntryGetNFn func(offset uint) ([]SubscribedFeedEntry, error)
)

func (s *Service) feedsByPage(
	ctx context.Context,
	userUUID string,
	number uint,
	getCount getCountFn,
	subscriptionEntryGetN subscriptionEntryGetNFn,
	pageTitle string,
	pageDescription string,
) (FeedPage, error) {
	if number < 1 {
		return FeedPage{}, paginate.ErrPageNumberOutOfBounds
	}

	entryCount, err := getCount()
	if err != nil {
		return FeedPage{}, err
	}

	totalPages := paginate.PageCount(entryCount, entriesPerPage)

	if number > totalPages {
		return FeedPage{}, paginate.ErrPageNumberOutOfBounds
	}

	categories, err := s.r.FeedSubscriptionCategoryGetAll(ctx, userUUID)
	if err != nil {
		return FeedPage{}, err
	}

	if len(categories) == 0 {
		// early return: nothing to display
		return NewFeedPage(1, 1, pageTitle, pageDescription, []SubscribedFeedsByCategory{}, 0, []SubscribedFeedEntry{}), nil
	}

	offset := (number - 1) * entriesPerPage

	entries, err := subscriptionEntryGetN(offset)
	if err != nil {
		return FeedPage{}, err
	}

	return NewFeedPage(number, totalPages, pageTitle, pageDescription, categories, entryCount, entries), nil
}

// FeedsByPage returns a Page containing a limited and offset number of feeds.
func (s *Service) FeedsByPage(ctx context.Context, userUUID string, preferences feed.Preferences, number uint) (FeedPage, error) {
	getCountFn := func() (uint, error) {
		return s.r.FeedEntryGetCount(ctx, userUUID, preferences.ShowEntries)
	}

	subscriptionEntryGetNFn := func(offset uint) ([]SubscribedFeedEntry, error) {
		return s.r.FeedSubscriptionEntryGetN(ctx, userUUID, preferences, entriesPerPage, offset)
	}

	return s.feedsByPage(ctx, userUUID, number, getCountFn, subscriptionEntryGetNFn, PageHeaderAll, "")
}

// FeedsByCategoryAndPage returns a Page containing a limited and offset number of feeds.
func (s *Service) FeedsByCategoryAndPage(ctx context.Context, userUUID string, preferences feed.Preferences, category feed.Category, number uint) (FeedPage, error) {
	getCountFn := func() (uint, error) {
		return s.r.FeedEntryGetCountByCategory(ctx, userUUID, preferences.ShowEntries, category.UUID)
	}

	subscriptionEntryGetNFn := func(offset uint) ([]SubscribedFeedEntry, error) {
		return s.r.FeedSubscriptionEntryGetNByCategory(ctx, userUUID, preferences, category.UUID, entriesPerPage, offset)
	}

	return s.feedsByPage(ctx, userUUID, number, getCountFn, subscriptionEntryGetNFn, category.Name, "")
}

// FeedsBySubscriptionAndPage returns a Page containing a limited and offset number of feeds.
func (s *Service) FeedsBySubscriptionAndPage(ctx context.Context, userUUID string, preferences feed.Preferences, subscription feed.Subscription, number uint) (FeedPage, error) {
	getCountFn := func() (uint, error) {
		return s.r.FeedEntryGetCountBySubscription(ctx, userUUID, preferences.ShowEntries, subscription.UUID)
	}

	subscriptionEntryGetNFn := func(offset uint) ([]SubscribedFeedEntry, error) {
		return s.r.FeedSubscriptionEntryGetNBySubscription(ctx, userUUID, preferences, subscription.UUID, entriesPerPage, offset)
	}

	f, err := s.r.FeedGetByUUID(ctx, subscription.FeedUUID)
	if err != nil {
		return FeedPage{}, err
	}

	pageTitle := f.Title
	if subscription.Alias != "" {
		pageTitle = subscription.Alias
	}

	return s.feedsByPage(ctx, userUUID, number, getCountFn, subscriptionEntryGetNFn, pageTitle, f.Description)
}

func (s *Service) feedsByQueryAndPage(
	ctx context.Context,
	userUUID string,
	query string,
	number uint,
	getCount getCountFn,
	subscriptionEntryGetN subscriptionEntryGetNFn,
	pageTitle string,
	pageDescription string,
) (FeedPage, error) {
	if number < 1 {
		return FeedPage{}, paginate.ErrPageNumberOutOfBounds
	}

	entryCount, err := getCount()
	if err != nil {
		return FeedPage{}, err
	}

	totalPages := paginate.PageCount(entryCount, entriesPerPage)

	if number > totalPages {
		return FeedPage{}, paginate.ErrPageNumberOutOfBounds
	}

	categories, err := s.r.FeedSubscriptionCategoryGetAll(ctx, userUUID)
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

	return NewFeedSearchResultPage(query, entryCount, number, totalPages, pageTitle, pageDescription, categories, entries), nil
}

func (s *Service) FeedsByQueryAndPage(ctx context.Context, userUUID string, preferences feed.Preferences, query string, number uint) (FeedPage, error) {
	getCountFn := func() (uint, error) {
		return s.r.FeedEntryGetCountByQuery(ctx, userUUID, preferences.ShowEntries, query)
	}

	subscriptionEntryGetNFn := func(offset uint) ([]SubscribedFeedEntry, error) {
		return s.r.FeedSubscriptionEntryGetNByQuery(ctx, userUUID, preferences, query, entriesPerPage, offset)
	}

	return s.feedsByQueryAndPage(ctx, userUUID, query, number, getCountFn, subscriptionEntryGetNFn, PageHeaderAll, "")
}

func (s *Service) FeedsByCategoryAndQueryAndPage(ctx context.Context, userUUID string, preferences feed.Preferences, category feed.Category, query string, number uint) (FeedPage, error) {
	getCountFn := func() (uint, error) {
		return s.r.FeedEntryGetCountByCategoryAndQuery(ctx, userUUID, preferences.ShowEntries, category.UUID, query)
	}

	subscriptionEntryGetNFn := func(offset uint) ([]SubscribedFeedEntry, error) {
		return s.r.FeedSubscriptionEntryGetNByCategoryAndQuery(ctx, userUUID, preferences, category.UUID, query, entriesPerPage, offset)
	}

	return s.feedsByQueryAndPage(ctx, userUUID, query, number, getCountFn, subscriptionEntryGetNFn, category.Name, "")
}
func (s *Service) FeedsBySubscriptionAndQueryAndPage(ctx context.Context, userUUID string, preferences feed.Preferences, subscription feed.Subscription, query string, number uint) (FeedPage, error) {
	getCountFn := func() (uint, error) {
		return s.r.FeedEntryGetCountBySubscriptionAndQuery(ctx, userUUID, preferences.ShowEntries, subscription.UUID, query)
	}

	subscriptionEntryGetNFn := func(offset uint) ([]SubscribedFeedEntry, error) {
		return s.r.FeedSubscriptionEntryGetNBySubscriptionAndQuery(ctx, userUUID, preferences, subscription.UUID, query, entriesPerPage, offset)
	}

	f, err := s.r.FeedGetByUUID(ctx, subscription.FeedUUID)
	if err != nil {
		return FeedPage{}, err
	}

	pageTitle := f.Title
	if subscription.Alias != "" {
		pageTitle = subscription.Alias
	}

	return s.feedsByQueryAndPage(ctx, userUUID, query, number, getCountFn, subscriptionEntryGetNFn, pageTitle, f.Description)
}

func (s *Service) SubscriptionByUUID(ctx context.Context, userUUID string, subscriptionUUID string) (Subscription, error) {
	return s.r.FeedQueryingSubscriptionByUUID(ctx, userUUID, subscriptionUUID)
}

func (s *Service) SubscriptionsByCategory(ctx context.Context, userUUID string) ([]SubscriptionsByCategory, error) {
	return s.r.FeedQueryingSubscriptionsByCategory(ctx, userUUID)
}
