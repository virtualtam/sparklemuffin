// Copyright (c) VirtualTam
// SPDX-License-Identifier: MIT

package feed

import (
	"errors"
	"fmt"
	"time"

	"github.com/mmcdole/gofeed"
	"github.com/rs/zerolog/log"
	"github.com/virtualtam/sparklemuffin/pkg/feed/fetching"
)

// Service handles operations for the feed domain.
type Service struct {
	r Repository

	client *fetching.Client
}

// NewService initializes and returns a Feed Service.
func NewService(r Repository, client *fetching.Client) *Service {
	return &Service{
		r:      r,
		client: client,
	}
}

// CreateCategory creates a new Category for a given User.
func (s *Service) CreateCategory(userUUID string, name string) (Category, error) {
	category, err := NewCategory(userUUID, name)
	if err != nil {
		return Category{}, err
	}

	if err := category.ValidateForAddition(s.r); err != nil {
		return Category{}, err
	}

	if err := s.r.FeedCategoryCreate(category); err != nil {
		return Category{}, err
	}

	return category, nil
}

// GetOrCreateCategory returns an existing Category or creates it.
func (s *Service) GetOrCreateCategory(userUUID string, name string) (Category, bool, error) {
	var isCreated bool

	category, err := s.r.FeedCategoryGetByName(userUUID, name)
	if errors.Is(err, ErrCategoryNotFound) {
		category, err = s.CreateCategory(userUUID, name)
		if err != nil {
			return Category{}, false, err
		}

		isCreated = true
	} else if err != nil {
		return Category{}, false, err
	}

	return category, isCreated, nil
}

// CategoryBySlug returns the category for a given user and slug.
func (s *Service) CategoryBySlug(userUUID string, slug string) (Category, error) {
	category := Category{Slug: slug}

	if err := category.validateSlug(); err != nil {
		return Category{}, err
	}

	return s.r.FeedCategoryGetBySlug(userUUID, slug)
}

// CategoryByUUID returns the category for a given user and UUID.
func (s *Service) CategoryByUUID(userUUID string, categoryUUID string) (Category, error) {
	category := Category{UUID: categoryUUID}

	if err := category.validateUUID(); err != nil {
		return Category{}, err
	}

	return s.r.FeedCategoryGetByUUID(userUUID, categoryUUID)
}

// Categories returns all categories for a given user.
func (s *Service) Categories(userUUID string) ([]Category, error) {
	return s.r.FeedCategoryGetMany(userUUID)
}

// DeleteCategory deletes a Category, related Subscriptions and EntryStatuses.
func (s *Service) DeleteCategory(userUUID string, categoryUUID string) error {
	categoryToDelete := Category{
		UserUUID: userUUID,
		UUID:     categoryUUID,
	}

	if err := categoryToDelete.ValidateForDeletion(s.r); err != nil {
		return err
	}

	return s.r.FeedCategoryDelete(userUUID, categoryUUID)
}

// UpdateCategory updates an existing Category.
func (s *Service) UpdateCategory(category Category) error {
	categoryToUpdate, err := s.CategoryByUUID(category.UserUUID, category.UUID)
	if err != nil {
		return err
	}

	categoryToUpdate.Name = category.Name

	now := time.Now().UTC()
	categoryToUpdate.UpdatedAt = now

	categoryToUpdate.Normalize()

	if err := categoryToUpdate.ValidateForUpdate(s.r); err != nil {
		return err
	}

	return s.r.FeedCategoryUpdate(categoryToUpdate)
}

// FeedBySlug returns the Feed for a given slug.
func (s *Service) FeedBySlug(userUUID string, slug string) (Feed, error) {
	feed := Feed{Slug: slug}

	if err := feed.ValidateSlug(); err != nil {
		return Feed{}, err
	}

	return s.r.FeedGetBySlug(slug)
}

// Subscribe creates a new Feed if needed, and creates the corresponding Subscription
// for a given user.
func (s *Service) Subscribe(userUUID string, categoryUUID string, feedURL string) error {
	feed, _, err := s.GetOrCreateFeedAndEntries(feedURL)
	if err != nil {
		return fmt.Errorf("failed to create or retrieve feed: %w", err)
	}

	subscription := Subscription{
		CategoryUUID: categoryUUID,
		FeedUUID:     feed.UUID,
		UserUUID:     userUUID,
	}

	if _, err := s.createSubscription(subscription); err != nil {
		return fmt.Errorf("failed to create subscription: %w", err)
	}

	return nil
}

// MarkAllEntriesAsRead marks all entries as "read" for a given User.
func (s *Service) MarkAllEntriesAsRead(userUUID string) error {
	return s.r.FeedEntryMarkAllAsRead(userUUID)
}

// MarkAllEntriesAsReadByCategory marks all entries as "read" for a given User and Category.
func (s *Service) MarkAllEntriesAsReadByCategory(userUUID string, categoryUUID string) error {
	return s.r.FeedEntryMarkAllAsReadByCategory(userUUID, categoryUUID)
}

// MarkAllEntriesAsReadBySubscription marks all entries as "read" for a given User and Subscription.
func (s *Service) MarkAllEntriesAsReadBySubscription(userUUID string, subscriptionUUID string) error {
	return s.r.FeedEntryMarkAllAsReadBySubscription(userUUID, subscriptionUUID)
}

// ToggleEntryRead toggles the "read" status for a given User and Entry.
func (s *Service) ToggleEntryRead(userUUID string, entryUID string) error {
	entryMetadata, err := s.r.FeedEntryMetadataGetByUID(userUUID, entryUID)
	if errors.Is(err, ErrEntryMetadataNotFound) {
		newEntryMetadata := EntryMetadata{
			UserUUID: userUUID,
			EntryUID: entryUID,
			Read:     true,
		}

		if err := s.r.FeedEntryMetadataCreate(newEntryMetadata); err != nil {
			return fmt.Errorf("failed to create entry metadata: %w", err)
		}

		return nil

	} else if err != nil {
		return fmt.Errorf("failed to retrieve entry metadata: %w", err)

	}

	entryMetadata.Read = !entryMetadata.Read
	if err := s.r.FeedEntryMetadataUpdate(entryMetadata); err != nil {
		return fmt.Errorf("failed to update entry metadata: %w", err)
	}

	return nil
}

func (s *Service) DeleteSubscription(userUUID string, subscriptionUUID string) error {
	return s.r.FeedSubscriptionDelete(userUUID, subscriptionUUID)
}

func (s *Service) SubscriptionByFeed(userUUID string, feedUUID string) (Subscription, error) {
	return s.r.FeedSubscriptionGetByFeed(userUUID, feedUUID)
}

func (s *Service) UpdateSubscription(subscription Subscription) error {
	subscriptionToUpdate, err := s.r.FeedSubscriptionGetByUUID(subscription.UserUUID, subscription.UUID)
	if err != nil {
		return err
	}

	subscriptionToUpdate.CategoryUUID = subscription.CategoryUUID

	now := time.Now().UTC()
	subscriptionToUpdate.UpdatedAt = now

	return s.r.FeedSubscriptionUpdate(subscriptionToUpdate)
}

func (s *Service) createEntries(feedUUID string, items []*gofeed.Item) error {
	var entries []Entry
	now := time.Now().UTC()

	for _, item := range items {
		entry := NewEntryFromItem(feedUUID, now, item)

		if err := entry.ValidateForAddition(); err != nil {
			log.
				Warn().
				Err(err).
				Str("feed_uuid", entry.FeedUUID).
				Str("entry_url", entry.URL).
				Msg("feeds: skipping invalid entry")
			continue
		}

		entries = append(entries, entry)
	}

	n, err := s.r.FeedEntryCreateMany(entries)
	if err != nil {
		return err
	}
	if n != int64(len(entries)) {
		return fmt.Errorf("feed: %d entries created, %d expected", n, len(entries))
	}

	return nil
}

func (s *Service) createFeedAndEntries(feed Feed) (Feed, error) {
	feedStatus, err := s.client.Fetch(feed.FeedURL, "", time.Time{})
	if err != nil {
		return Feed{}, err
	}

	feed.Title = feedStatus.Feed.Title
	feed.ETag = feedStatus.ETag
	feed.LastModified = feedStatus.LastModified
	feed.FetchedAt = time.Now().UTC()
	feed.Normalize()

	if err := feed.ValidateForCreation(); err != nil {
		return Feed{}, err
	}

	if err := s.r.FeedCreate(feed); err != nil {
		return Feed{}, err
	}

	if err := s.createEntries(feed.UUID, feedStatus.Feed.Items); err != nil {
		return Feed{}, err
	}

	return feed, nil
}

// GetOrCreateFeedAndEntries returns an existing feed, or creates it (along with its entries).
func (s *Service) GetOrCreateFeedAndEntries(feedURL string) (Feed, bool, error) {
	newFeed, err := NewFeed(feedURL)
	if err != nil {
		return Feed{}, false, err
	}

	if err := newFeed.ValidateURL(); err != nil {
		return Feed{}, false, err
	}

	var feed Feed
	var isCreated bool

	// Attempt to retrieve an existing feed
	feed, err = s.r.FeedGetByURL(newFeed.FeedURL)
	if errors.Is(err, ErrFeedNotFound) {
		// Else, create it
		feed, err = s.createFeedAndEntries(newFeed)
		if err != nil {
			return Feed{}, false, err
		}

		isCreated = true

	} else if err != nil {
		return Feed{}, false, err
	}

	return feed, isCreated, nil
}

// GetOrCreateSubscription returns an existing subscription or creates it.
func (s *Service) GetOrCreateSubscription(newSubscription Subscription) (Subscription, bool, error) {
	var isCreated bool

	subscription, err := s.r.FeedSubscriptionGetByFeed(newSubscription.UserUUID, newSubscription.FeedUUID)
	if errors.Is(err, ErrSubscriptionNotFound) {
		subscription, err = s.createSubscription(newSubscription)
		if err != nil {
			return Subscription{}, false, err
		}

		isCreated = true
	} else if err != nil {
		return Subscription{}, false, err
	}

	return subscription, isCreated, nil
}

func (s *Service) createSubscription(newSubscription Subscription) (Subscription, error) {
	subscription, err := NewSubscription(newSubscription.CategoryUUID, newSubscription.FeedUUID, newSubscription.UserUUID)
	if err != nil {
		return Subscription{}, fmt.Errorf("failed to create subscription: %w", err)
	}

	if err := subscription.ValidateForCreation(s.r); err != nil {
		return Subscription{}, err
	}

	return s.r.FeedSubscriptionCreate(subscription)
}
