// Copyright (c) VirtualTam
// SPDX-License-Identifier: MIT

package feed

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/mmcdole/gofeed"
	"github.com/rs/zerolog/log"

	"github.com/virtualtam/sparklemuffin/internal/textkit"
	"github.com/virtualtam/sparklemuffin/pkg/feed/fetching"
)

// Service handles operations for the feed domain.
type Service struct {
	r Repository

	client *fetching.Client

	textRanker       *textkit.TextRanker
	textRankMaxTerms int
}

// NewService initializes and returns a Feed Service.
func NewService(r Repository, client *fetching.Client) *Service {
	return &Service{
		r:                r,
		client:           client,
		textRanker:       textkit.NewTextRanker(),
		textRankMaxTerms: EntryTextRankMaxTerms,
	}
}

// CreateCategory creates a new Category for a given User.
func (s *Service) CreateCategory(ctx context.Context, userUUID string, name string) (Category, error) {
	category, err := NewCategory(userUUID, name)
	if err != nil {
		return Category{}, err
	}

	if err := category.ValidateForAddition(ctx, s.r); err != nil {
		return Category{}, err
	}

	if err := s.r.FeedCategoryCreate(ctx, category); err != nil {
		return Category{}, err
	}

	return category, nil
}

// GetOrCreateCategory returns an existing Category or creates it.
func (s *Service) GetOrCreateCategory(ctx context.Context, userUUID string, name string) (Category, bool, error) {
	var isCreated bool

	category, err := s.r.FeedCategoryGetByName(ctx, userUUID, name)
	if errors.Is(err, ErrCategoryNotFound) {
		category, err = s.CreateCategory(ctx, userUUID, name)
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
func (s *Service) CategoryBySlug(ctx context.Context, userUUID string, slug string) (Category, error) {
	category := Category{Slug: slug}

	if err := category.validateSlug(); err != nil {
		return Category{}, err
	}

	return s.r.FeedCategoryGetBySlug(ctx, userUUID, slug)
}

// CategoryByUUID returns the category for a given user and UUID.
func (s *Service) CategoryByUUID(ctx context.Context, userUUID string, categoryUUID string) (Category, error) {
	category := Category{UUID: categoryUUID}

	if err := category.validateUUID(); err != nil {
		return Category{}, err
	}

	return s.r.FeedCategoryGetByUUID(ctx, userUUID, categoryUUID)
}

// Categories returns all categories for a given user.
func (s *Service) Categories(ctx context.Context, userUUID string) ([]Category, error) {
	return s.r.FeedCategoryGetMany(ctx, userUUID)
}

// DeleteCategory deletes a Category, related Subscriptions and EntryStatuses.
func (s *Service) DeleteCategory(ctx context.Context, userUUID string, categoryUUID string) error {
	categoryToDelete := Category{
		UserUUID: userUUID,
		UUID:     categoryUUID,
	}

	if err := categoryToDelete.ValidateForDeletion(); err != nil {
		return err
	}

	return s.r.FeedCategoryDelete(ctx, userUUID, categoryUUID)
}

// UpdateCategory updates an existing Category.
func (s *Service) UpdateCategory(ctx context.Context, category Category) error {
	categoryToUpdate, err := s.CategoryByUUID(ctx, category.UserUUID, category.UUID)
	if err != nil {
		return err
	}

	categoryToUpdate.Name = category.Name

	now := time.Now().UTC()
	categoryToUpdate.UpdatedAt = now

	categoryToUpdate.Normalize()

	if err := categoryToUpdate.ValidateForUpdate(ctx, s.r); err != nil {
		return err
	}

	return s.r.FeedCategoryUpdate(ctx, categoryToUpdate)
}

// FeedBySlug returns the Feed for a given slug.
func (s *Service) FeedBySlug(ctx context.Context, slug string) (Feed, error) {
	feed := Feed{Slug: slug}

	if err := feed.ValidateSlug(); err != nil {
		return Feed{}, err
	}

	return s.r.FeedGetBySlug(ctx, slug)
}

// Subscribe creates a new Feed if needed, and creates the corresponding Subscription
// for a given user.
func (s *Service) Subscribe(ctx context.Context, userUUID string, categoryUUID string, feedURL string) error {
	feed, _, err := s.GetOrCreateFeedAndEntries(ctx, feedURL)
	if err != nil {
		return fmt.Errorf("failed to create or retrieve feed: %w", err)
	}

	subscription := Subscription{
		CategoryUUID: categoryUUID,
		FeedUUID:     feed.UUID,
		UserUUID:     userUUID,
	}

	if _, err := s.createSubscription(ctx, subscription); err != nil {
		return fmt.Errorf("failed to create subscription: %w", err)
	}

	return nil
}

// MarkAllEntriesAsRead marks all entries as "read" for a given User.
func (s *Service) MarkAllEntriesAsRead(ctx context.Context, userUUID string) error {
	return s.r.FeedEntryMarkAllAsRead(ctx, userUUID)
}

// MarkAllEntriesAsReadByCategory marks all entries as "read" for a given User and Category.
func (s *Service) MarkAllEntriesAsReadByCategory(ctx context.Context, userUUID string, categoryUUID string) error {
	return s.r.FeedEntryMarkAllAsReadByCategory(ctx, userUUID, categoryUUID)
}

// MarkAllEntriesAsReadBySubscription marks all entries as "read" for a given User and Subscription.
func (s *Service) MarkAllEntriesAsReadBySubscription(ctx context.Context, userUUID string, subscriptionUUID string) error {
	return s.r.FeedEntryMarkAllAsReadBySubscription(ctx, userUUID, subscriptionUUID)
}

// ToggleEntryRead toggles the "read" status for a given User and Entry.
func (s *Service) ToggleEntryRead(ctx context.Context, userUUID string, entryUID string) error {
	entryMetadata, err := s.r.FeedEntryMetadataGetByUID(ctx, userUUID, entryUID)
	if errors.Is(err, ErrEntryMetadataNotFound) {
		newEntryMetadata := EntryMetadata{
			UserUUID: userUUID,
			EntryUID: entryUID,
			Read:     true,
		}

		if err := s.r.FeedEntryMetadataCreate(ctx, newEntryMetadata); err != nil {
			return fmt.Errorf("failed to create entry metadata: %w", err)
		}

		return nil

	} else if err != nil {
		return fmt.Errorf("failed to retrieve entry metadata: %w", err)

	}

	entryMetadata.Read = !entryMetadata.Read
	if err := s.r.FeedEntryMetadataUpdate(ctx, entryMetadata); err != nil {
		return fmt.Errorf("failed to update entry metadata: %w", err)
	}

	return nil
}

// PreferencesByUserUUID returns the feed Preferences for a given user.
func (s *Service) PreferencesByUserUUID(ctx context.Context, userUUID string) (Preferences, error) {
	return s.r.FeedPreferencesGetByUserUUID(ctx, userUUID)
}

func (s *Service) UpdatePreferences(ctx context.Context, preferences Preferences) error {
	preferences.UpdatedAt = time.Now().UTC()

	if err := preferences.ValidateForUpdate(); err != nil {
		return err
	}

	return s.r.FeedPreferencesUpdate(ctx, preferences)
}

func (s *Service) DeleteSubscription(ctx context.Context, userUUID string, subscriptionUUID string) error {
	subscription, err := s.r.FeedSubscriptionGetByUUID(ctx, userUUID, subscriptionUUID)
	if err != nil {
		return err
	}

	if err := s.r.FeedSubscriptionDelete(ctx, userUUID, subscription.UUID); err != nil {
		return err
	}

	return nil
}

func (s *Service) SubscriptionByFeed(ctx context.Context, userUUID string, feedUUID string) (Subscription, error) {
	return s.r.FeedSubscriptionGetByFeed(ctx, userUUID, feedUUID)
}

func (s *Service) UpdateSubscription(ctx context.Context, subscription Subscription) error {
	subscriptionToUpdate, err := s.r.FeedSubscriptionGetByUUID(ctx, subscription.UserUUID, subscription.UUID)
	if err != nil {
		return err
	}

	subscriptionToUpdate.Alias = subscription.Alias
	subscriptionToUpdate.CategoryUUID = subscription.CategoryUUID

	now := time.Now().UTC()
	subscriptionToUpdate.UpdatedAt = now

	subscriptionToUpdate.Normalize()

	return s.r.FeedSubscriptionUpdate(ctx, subscriptionToUpdate)
}

func (s *Service) createEntries(ctx context.Context, feedUUID string, items []*gofeed.Item) error {
	var entries []Entry
	now := time.Now().UTC()

	for _, item := range items {
		entry := NewEntryFromItem(feedUUID, now, item)
		entry.ExtractTextRankTerms(s.textRanker, s.textRankMaxTerms)

		if err := entry.ValidateForAddition(now); err != nil {
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

	n, err := s.r.FeedEntryCreateMany(ctx, entries)
	if err != nil {
		return err
	}
	if n != int64(len(entries)) {
		return fmt.Errorf("feed: %d entries created, %d expected", n, len(entries))
	}

	return nil
}

func (s *Service) createFeedAndEntries(ctx context.Context, feed Feed) (Feed, error) {
	feedStatus, err := s.client.Fetch(ctx, feed.FeedURL, "", time.Time{})
	if err != nil {
		return Feed{}, err
	}

	feed.Title = feedStatus.Feed.Title
	feed.Description = feedStatus.Feed.Description
	feed.ETag = feedStatus.ETag
	feed.Hash = feedStatus.Hash
	feed.LastModified = feedStatus.LastModified
	feed.FetchedAt = time.Now().UTC()
	feed.Normalize()

	if err := feed.ValidateForCreation(); err != nil {
		return Feed{}, err
	}

	if err := s.r.FeedCreate(ctx, feed); err != nil {
		return Feed{}, err
	}

	if err := s.createEntries(ctx, feed.UUID, feedStatus.Feed.Items); err != nil {
		return Feed{}, err
	}

	return feed, nil
}

// GetOrCreateFeedAndEntries returns an existing feed, or creates it (along with its entries).
func (s *Service) GetOrCreateFeedAndEntries(ctx context.Context, feedURL string) (Feed, bool, error) {
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
	feed, err = s.r.FeedGetByURL(ctx, newFeed.FeedURL)
	if errors.Is(err, ErrFeedNotFound) {
		// Else, create it
		feed, err = s.createFeedAndEntries(ctx, newFeed)
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
func (s *Service) GetOrCreateSubscription(ctx context.Context, newSubscription Subscription) (Subscription, bool, error) {
	var isCreated bool

	subscription, err := s.r.FeedSubscriptionGetByFeed(ctx, newSubscription.UserUUID, newSubscription.FeedUUID)
	if errors.Is(err, ErrSubscriptionNotFound) {
		subscription, err = s.createSubscription(ctx, newSubscription)
		if err != nil {
			return Subscription{}, false, err
		}

		isCreated = true
	} else if err != nil {
		return Subscription{}, false, err
	}

	return subscription, isCreated, nil
}

func (s *Service) createSubscription(ctx context.Context, newSubscription Subscription) (Subscription, error) {
	subscription, err := NewSubscription(newSubscription.CategoryUUID, newSubscription.FeedUUID, newSubscription.UserUUID)
	if err != nil {
		return Subscription{}, fmt.Errorf("failed to create subscription: %w", err)
	}

	if err := subscription.ValidateForCreation(ctx, s.r); err != nil {
		return Subscription{}, err
	}

	return s.r.FeedSubscriptionCreate(ctx, subscription)
}
