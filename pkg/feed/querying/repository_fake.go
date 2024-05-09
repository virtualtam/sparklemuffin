// Copyright (c) VirtualTam
// SPDX-License-Identifier: MIT

package querying

import (
	"sort"

	"github.com/virtualtam/sparklemuffin/pkg/feed"
)

var _ Repository = &fakeRepository{}

type fakeRepository struct {
	Categories    []feed.Category
	Entries       []feed.Entry
	Feeds         []feed.Feed
	Subscriptions []feed.Subscription
}

func (r *fakeRepository) FeedSubscriptionCategoryGetAll(userUUID string) ([]SubscriptionCategory, error) {
	var subscriptionCategories []SubscriptionCategory

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

				subscriptionUnread++
			}

			categoryUnread += subscriptionUnread

			for _, f := range r.Feeds {
				if f.UUID != subscription.FeedUUID {
					continue
				}

				subscribedFeed := SubscribedFeed{
					Feed:   f,
					Unread: subscriptionUnread,
				}

				subscribedFeeds = append(subscribedFeeds, subscribedFeed)

				break
			}
		}

		subscriptionCategory := SubscriptionCategory{
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

func (r *fakeRepository) FeedSubscriptionEntryGetN(userUUID string, n uint, offset uint) ([]SubscriptionEntry, error) {
	var subscriptionEntries []SubscriptionEntry

	for _, subscription := range r.Subscriptions {
		if subscription.UserUUID != userUUID {
			continue
		}

		for _, entry := range r.Entries {
			if entry.FeedUUID != subscription.FeedUUID {
				continue
			}

			subscriptionEntry := SubscriptionEntry{
				Entry: entry,
				Read:  false,
			}

			subscriptionEntries = append(subscriptionEntries, subscriptionEntry)
			break
		}
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

func (r *fakeRepository) FeedSubscriptionEntryGetNByCategory(userUUID string, categoryUUID string, n uint, offset uint) ([]SubscriptionEntry, error) {
	var subscriptionEntries []SubscriptionEntry

	for _, subscription := range r.Subscriptions {
		if subscription.UserUUID != userUUID {
			continue
		}

		if subscription.CategoryUUID != categoryUUID {
			continue
		}

		for _, entry := range r.Entries {
			if entry.FeedUUID != subscription.FeedUUID {
				continue
			}

			subscriptionEntry := SubscriptionEntry{
				Entry: entry,
				Read:  false,
			}

			subscriptionEntries = append(subscriptionEntries, subscriptionEntry)
			break
		}
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

func (r *fakeRepository) FeedSubscriptionEntryGetNBySubscription(userUUID string, subscriptionUUID string, n uint, offset uint) ([]SubscriptionEntry, error) {
	var subscriptionEntries []SubscriptionEntry

	subscription, err := r.FeedSubscriptionGetByUUID(userUUID, subscriptionUUID)
	if err != nil {
		return []SubscriptionEntry{}, err
	}

	for _, entry := range r.Entries {
		if entry.FeedUUID != subscription.FeedUUID {
			continue
		}

		subscriptionEntry := SubscriptionEntry{
			Entry: entry,
			Read:  false,
		}

		subscriptionEntries = append(subscriptionEntries, subscriptionEntry)
		break
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
