// Copyright (c) VirtualTam
// SPDX-License-Identifier: MIT

package querying

import (
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

func (r *fakeRepository) FeedGetByUUID(feedUUID string) (feed.Feed, error) {
	for _, f := range r.Feeds {
		if f.UUID == feedUUID {
			return f, nil
		}
	}

	return feed.Feed{}, feed.ErrFeedNotFound
}

func (r *fakeRepository) FeedSubscriptionCategoryGetAll(userUUID string) ([]SubscribedFeedsByCategory, error) {
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

func (r *fakeRepository) FeedEntryGetCount(userUUID string) (uint, error) {
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

func (r *fakeRepository) FeedEntryGetCountByCategory(userUUID string, categoryUUID string) (uint, error) {
	var count uint

	for _, subscription := range r.Subscriptions {
		if subscription.UserUUID != userUUID {
			continue
		}

		if subscription.CategoryUUID != categoryUUID {
			continue
		}

		subCount, err := r.FeedEntryGetCountBySubscription(userUUID, subscription.UUID)
		if err != nil {
			return 0, err
		}

		count += subCount
	}

	return count, nil
}

func (r *fakeRepository) FeedEntryGetCountByQuery(userUUID string, searchTerms string) (uint, error) {
	return 0, errors.New("not implemented")
}

func (r *fakeRepository) FeedEntryGetCountByCategoryAndQuery(userUUID string, categoryUUID string, searchTerms string) (uint, error) {
	return 0, errors.New("not implemented")
}

func (r *fakeRepository) FeedEntryGetCountBySubscriptionAndQuery(userUUID string, subscriptionUUID string, searchTerms string) (uint, error) {
	return 0, errors.New("not implemented")
}

func (r *fakeRepository) FeedSubscriptionGetByUUID(userUUID string, subscriptionUUID string) (feed.Subscription, error) {
	for _, subscription := range r.Subscriptions {
		if subscription.UserUUID == userUUID && subscription.UUID == subscriptionUUID {
			return subscription, nil
		}
	}

	return feed.Subscription{}, feed.ErrSubscriptionNotFound
}

func (r *fakeRepository) FeedEntryGetCountBySubscription(userUUID string, subscriptionUUID string) (uint, error) {
	var count uint

	subscription, err := r.FeedSubscriptionGetByUUID(userUUID, subscriptionUUID)
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

func (r *fakeRepository) subscribedFeedEntryGetByFeed(feedUUID string) ([]SubscribedFeedEntry, error) {
	var subscriptionEntries []SubscribedFeedEntry

	f, err := r.FeedGetByUUID(feedUUID)
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
			Entry:     entry,
			FeedTitle: f.Title,
			Read:      read,
		}

		subscriptionEntries = append(subscriptionEntries, subscriptionEntry)
	}

	return subscriptionEntries, nil
}

func (r *fakeRepository) FeedSubscriptionEntryGetN(userUUID string, n uint, offset uint) ([]SubscribedFeedEntry, error) {
	var userEntries []SubscribedFeedEntry

	for _, subscription := range r.Subscriptions {
		if subscription.UserUUID != userUUID {
			continue
		}

		subscriptionEntries, err := r.subscribedFeedEntryGetByFeed(subscription.FeedUUID)
		if err != nil {
			return []SubscribedFeedEntry{}, err
		}

		userEntries = append(userEntries, subscriptionEntries...)
	}

	// FIXME subscriptions should be sorted **before** querying
	sort.Slice(userEntries, func(i, j int) bool {
		return userEntries[i].PublishedAt.After(userEntries[j].PublishedAt)
	})

	var nEntries uint

	if n > uint(len(userEntries[offset:])) {
		nEntries = uint(len(userEntries[offset:]))
	} else {
		nEntries = n
	}

	return userEntries[offset : offset+nEntries], nil
}

func (r *fakeRepository) FeedSubscriptionEntryGetNByCategory(userUUID string, categoryUUID string, n uint, offset uint) ([]SubscribedFeedEntry, error) {
	var categoryEntries []SubscribedFeedEntry

	for _, subscription := range r.Subscriptions {
		if subscription.UserUUID != userUUID {
			continue
		}

		if subscription.CategoryUUID != categoryUUID {
			continue
		}

		subscriptionEntries, err := r.subscribedFeedEntryGetByFeed(subscription.FeedUUID)
		if err != nil {
			return []SubscribedFeedEntry{}, err
		}

		categoryEntries = append(categoryEntries, subscriptionEntries...)
	}

	// FIXME subscriptions should be sorted **before** querying
	sort.Slice(categoryEntries, func(i, j int) bool {
		return categoryEntries[i].PublishedAt.After(categoryEntries[j].PublishedAt)
	})

	var nEntries uint

	if n > uint(len(categoryEntries[offset:])) {
		nEntries = uint(len(categoryEntries[offset:]))
	} else {
		nEntries = n
	}

	return categoryEntries[offset : offset+nEntries], nil
}

func (r *fakeRepository) FeedSubscriptionEntryGetNBySubscription(userUUID string, subscriptionUUID string, n uint, offset uint) ([]SubscribedFeedEntry, error) {
	subscription, err := r.FeedSubscriptionGetByUUID(userUUID, subscriptionUUID)
	if err != nil {
		return []SubscribedFeedEntry{}, err
	}

	subscriptionEntries, err := r.subscribedFeedEntryGetByFeed(subscription.FeedUUID)
	if err != nil {
		return []SubscribedFeedEntry{}, err
	}

	// FIXME subscriptions should be sorted **before** querying
	sort.Slice(subscriptionEntries, func(i, j int) bool {
		return subscriptionEntries[i].PublishedAt.After(subscriptionEntries[j].PublishedAt)
	})

	var nEntries uint

	if n > uint(len(subscriptionEntries[offset:])) {
		nEntries = uint(len(subscriptionEntries[offset:]))
	} else {
		nEntries = n
	}

	return subscriptionEntries[offset : offset+nEntries], nil
}

func (r *fakeRepository) FeedSubscriptionEntryGetNByQuery(userUUID string, searchTerms string, entriesPerPage uint, offset uint) ([]SubscribedFeedEntry, error) {
	return []SubscribedFeedEntry{}, errors.New("not implemented")
}

func (r *fakeRepository) FeedSubscriptionEntryGetNByCategoryAndQuery(userUUID string, categoryUUID string, searchTerms string, entriesPerPage uint, offset uint) ([]SubscribedFeedEntry, error) {
	return []SubscribedFeedEntry{}, errors.New("not implemented")
}

func (r *fakeRepository) FeedSubscriptionEntryGetNBySubscriptionAndQuery(userUUID string, subscriptionUUID string, searchTerms string, entriesPerPage uint, offset uint) ([]SubscribedFeedEntry, error) {
	return []SubscribedFeedEntry{}, errors.New("not implemented")
}

func (r *fakeRepository) FeedQueryingSubscriptionByUUID(userUUID string, subscriptionUUID string) (Subscription, error) {
	return Subscription{}, errors.New("not implemented")
}

func (r *fakeRepository) FeedQueryingSubscriptionsByCategory(userUUID string) ([]SubscriptionsByCategory, error) {
	return []SubscriptionsByCategory{}, errors.New("not implemented")
}
