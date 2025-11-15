// Copyright (c) VirtualTam
// SPDX-License-Identifier: MIT

package querying

import (
	"context"
	"errors"
	"sort"

	"github.com/virtualtam/sparklemuffin/pkg/feed"
)

var _ Repository = &fakeRepository{}

type fakeRepository struct {
	Categories      []feed.Category
	Entries         []feed.Entry
	EntriesMetadata []feed.EntryMetadata
	Feeds           []feed.Feed
	Subscriptions   []feed.Subscription
}

func (r *fakeRepository) FeedGetByUUID(_ context.Context, feedUUID string) (feed.Feed, error) {
	for _, f := range r.Feeds {
		if f.UUID == feedUUID {
			return f, nil
		}
	}

	return feed.Feed{}, feed.ErrFeedNotFound
}

func (r *fakeRepository) FeedSubscriptionCategoryGetAll(_ context.Context, userUUID string) ([]SubscribedFeedsByCategory, error) {
	var subscriptionCategories []SubscribedFeedsByCategory

	for _, category := range r.Categories {
		if category.UserUUID != userUUID {
			continue
		}

		var categoryUnread uint
		var subscribedFeeds []SubscribedFeed

		for _, subscription := range r.Subscriptions {
			if subscription.CategoryUUID != category.UUID {
				continue
			}

			var subscriptionUnread uint

			for _, entry := range r.Entries {
				if entry.FeedUUID != subscription.FeedUUID {
					continue
				}

				var read bool

				for _, entryMetadata := range r.EntriesMetadata {
					if entryMetadata.EntryUID == entry.UID && entryMetadata.Read {
						read = true
						break
					}
				}

				if !read {
					subscriptionUnread++
				}
			}

			categoryUnread += subscriptionUnread

			for _, f := range r.Feeds {
				if f.UUID != subscription.FeedUUID {
					continue
				}

				subscribedFeed := SubscribedFeed{
					Feed:   f,
					Alias:  subscription.Alias,
					Unread: subscriptionUnread,
				}

				subscribedFeeds = append(subscribedFeeds, subscribedFeed)

				break
			}
		}

		subscriptionCategory := SubscribedFeedsByCategory{
			Category:        category,
			Unread:          categoryUnread,
			SubscribedFeeds: subscribedFeeds,
		}

		subscriptionCategories = append(subscriptionCategories, subscriptionCategory)
	}

	sort.Slice(subscriptionCategories, func(i, j int) bool {
		return subscriptionCategories[i].Name < subscriptionCategories[j].Name
	})

	return subscriptionCategories, nil
}

func (r *fakeRepository) FeedEntryGetCount(_ context.Context, userUUID string, showEntries feed.EntryVisibility) (uint, error) {
	var count uint

	for _, subscription := range r.Subscriptions {
		if subscription.UserUUID != userUUID {
			continue
		}

		for _, entry := range r.Entries {
			if entry.FeedUUID != subscription.FeedUUID {
				continue
			}

			count++
		}
	}

	return count, nil
}

func (r *fakeRepository) FeedEntryGetCountByCategory(ctx context.Context, userUUID string, showEntries feed.EntryVisibility, categoryUUID string) (uint, error) {
	var count uint

	for _, subscription := range r.Subscriptions {
		if subscription.UserUUID != userUUID {
			continue
		}

		if subscription.CategoryUUID != categoryUUID {
			continue
		}

		subCount, err := r.FeedEntryGetCountBySubscription(ctx, userUUID, showEntries, subscription.UUID)
		if err != nil {
			return 0, err
		}

		count += subCount
	}

	return count, nil
}

func (r *fakeRepository) FeedEntryGetCountBySubscription(ctx context.Context, userUUID string, showEntries feed.EntryVisibility, subscriptionUUID string) (uint, error) {
	var count uint

	subscription, err := r.FeedSubscriptionGetByUUID(ctx, userUUID, subscriptionUUID)
	if err != nil {
		return 0, err
	}

	for _, entry := range r.Entries {
		if entry.FeedUUID != subscription.FeedUUID {
			continue
		}

		count++
	}

	return count, nil
}

func (r *fakeRepository) FeedEntryGetCountByQuery(_ context.Context, userUUID string, showEntries feed.EntryVisibility, searchTerms string) (uint, error) {
	return 0, errors.New("not implemented")
}

func (r *fakeRepository) FeedEntryGetCountByCategoryAndQuery(_ context.Context, userUUID string, showEntries feed.EntryVisibility, categoryUUID string, searchTerms string) (uint, error) {
	return 0, errors.New("not implemented")
}

func (r *fakeRepository) FeedEntryGetCountBySubscriptionAndQuery(_ context.Context, userUUID string, showEntries feed.EntryVisibility, subscriptionUUID string, searchTerms string) (uint, error) {
	return 0, errors.New("not implemented")
}

func (r *fakeRepository) FeedSubscriptionGetByUUID(_ context.Context, userUUID string, subscriptionUUID string) (feed.Subscription, error) {
	for _, subscription := range r.Subscriptions {
		if subscription.UserUUID == userUUID && subscription.UUID == subscriptionUUID {
			return subscription, nil
		}
	}

	return feed.Subscription{}, feed.ErrSubscriptionNotFound
}

func (r *fakeRepository) feedSubscriptionGetByFeed(_ context.Context, userUUID string, feedUUID string) (feed.Subscription, error) {
	for _, subscription := range r.Subscriptions {
		if subscription.UserUUID == userUUID && subscription.FeedUUID == feedUUID {
			return subscription, nil
		}
	}

	return feed.Subscription{}, feed.ErrSubscriptionNotFound
}

func (r *fakeRepository) subscribedFeedEntryGetByFeed(ctx context.Context, userUUID string, feedUUID string) ([]SubscribedFeedEntry, error) {
	var subscriptionEntries []SubscribedFeedEntry

	f, err := r.FeedGetByUUID(ctx, feedUUID)
	if err != nil {
		return []SubscribedFeedEntry{}, err
	}

	subscription, err := r.feedSubscriptionGetByFeed(ctx, userUUID, feedUUID)
	if err != nil {
		return []SubscribedFeedEntry{}, err
	}

	for _, entry := range r.Entries {
		if entry.FeedUUID != feedUUID {
			continue
		}

		var read bool

		for _, entryMetadata := range r.EntriesMetadata {
			if entryMetadata.EntryUID == entry.UID && entryMetadata.Read {
				read = true
				break
			}
		}

		subscriptionEntry := SubscribedFeedEntry{
			Entry:             entry,
			SubscriptionAlias: subscription.Alias,
			FeedTitle:         f.Title,
			Read:              read,
		}

		subscriptionEntries = append(subscriptionEntries, subscriptionEntry)
	}

	return subscriptionEntries, nil
}

func (r *fakeRepository) FeedSubscriptionEntryGetN(ctx context.Context, userUUID string, preferences feed.Preferences, n uint, offset uint) ([]SubscribedFeedEntry, error) {
	var userEntries []SubscribedFeedEntry

	for _, subscription := range r.Subscriptions {
		if subscription.UserUUID != userUUID {
			continue
		}

		subscriptionEntries, err := r.subscribedFeedEntryGetByFeed(ctx, userUUID, subscription.FeedUUID)
		if err != nil {
			return []SubscribedFeedEntry{}, err
		}

		userEntries = append(userEntries, subscriptionEntries...)
	}

	// FIXME subscriptions should be sorted **before** querying
	sort.Slice(userEntries, func(i, j int) bool {
		return userEntries[i].PublishedAt.After(userEntries[j].PublishedAt)
	})

	nEntries := min(n, uint(len(userEntries[offset:])))

	return userEntries[offset : offset+nEntries], nil
}

func (r *fakeRepository) FeedSubscriptionEntryGetNByCategory(ctx context.Context, userUUID string, preferences feed.Preferences, categoryUUID string, n uint, offset uint) ([]SubscribedFeedEntry, error) {
	var categoryEntries []SubscribedFeedEntry

	for _, subscription := range r.Subscriptions {
		if subscription.UserUUID != userUUID {
			continue
		}

		if subscription.CategoryUUID != categoryUUID {
			continue
		}

		subscriptionEntries, err := r.subscribedFeedEntryGetByFeed(ctx, userUUID, subscription.FeedUUID)
		if err != nil {
			return []SubscribedFeedEntry{}, err
		}

		categoryEntries = append(categoryEntries, subscriptionEntries...)
	}

	// FIXME subscriptions should be sorted **before** querying
	sort.Slice(categoryEntries, func(i, j int) bool {
		return categoryEntries[i].PublishedAt.After(categoryEntries[j].PublishedAt)
	})

	nEntries := min(n, uint(len(categoryEntries[offset:])))

	return categoryEntries[offset : offset+nEntries], nil
}

func (r *fakeRepository) FeedSubscriptionEntryGetNBySubscription(ctx context.Context, userUUID string, preferences feed.Preferences, subscriptionUUID string, n uint, offset uint) ([]SubscribedFeedEntry, error) {
	subscription, err := r.FeedSubscriptionGetByUUID(ctx, userUUID, subscriptionUUID)
	if err != nil {
		return []SubscribedFeedEntry{}, err
	}

	subscriptionEntries, err := r.subscribedFeedEntryGetByFeed(ctx, userUUID, subscription.FeedUUID)
	if err != nil {
		return []SubscribedFeedEntry{}, err
	}

	// FIXME subscriptions should be sorted **before** querying
	sort.Slice(subscriptionEntries, func(i, j int) bool {
		return subscriptionEntries[i].PublishedAt.After(subscriptionEntries[j].PublishedAt)
	})

	nEntries := min(n, uint(len(subscriptionEntries[offset:])))

	return subscriptionEntries[offset : offset+nEntries], nil
}

func (r *fakeRepository) FeedSubscriptionEntryGetNByQuery(_ context.Context, userUUID string, preferences feed.Preferences, searchTerms string, entriesPerPage uint, offset uint) ([]SubscribedFeedEntry, error) {
	return []SubscribedFeedEntry{}, errors.New("not implemented")
}

func (r *fakeRepository) FeedSubscriptionEntryGetNByCategoryAndQuery(_ context.Context, userUUID string, preferences feed.Preferences, categoryUUID string, searchTerms string, entriesPerPage uint, offset uint) ([]SubscribedFeedEntry, error) {
	return []SubscribedFeedEntry{}, errors.New("not implemented")
}

func (r *fakeRepository) FeedSubscriptionEntryGetNBySubscriptionAndQuery(_ context.Context, userUUID string, preferences feed.Preferences, subscriptionUUID string, searchTerms string, entriesPerPage uint, offset uint) ([]SubscribedFeedEntry, error) {
	return []SubscribedFeedEntry{}, errors.New("not implemented")
}

func (r *fakeRepository) FeedQueryingSubscriptionByUUID(_ context.Context, userUUID string, subscriptionUUID string) (Subscription, error) {
	return Subscription{}, errors.New("not implemented")
}

func (r *fakeRepository) FeedQueryingSubscriptionsByCategory(_ context.Context, userUUID string) ([]SubscriptionsByCategory, error) {
	return []SubscriptionsByCategory{}, errors.New("not implemented")
}
