// Copyright (c) VirtualTam
// SPDX-License-Identifier: MIT

package feed

import (
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/mmcdole/gofeed"
	"github.com/rs/zerolog/log"
)

// Service handles operations for the feed domain.
type Service struct {
	r Repository

	feedParser *gofeed.Parser
}

// NewService initializes and returns a Feed Service.
func NewService(r Repository, httpClient *http.Client) *Service {
	feedParser := gofeed.NewParser()
	feedParser.Client = httpClient

	return &Service{
		r:          r,
		feedParser: feedParser,
	}
}

// AddCategory adds a new Category for a given User.
func (s *Service) AddCategory(userUUID string, name string) (Category, error) {
	category, err := NewCategory(userUUID, name)
	if err != nil {
		return Category{}, err
	}

	if err := category.ValidateForAddition(s.r); err != nil {
		return Category{}, err
	}

	if err := s.r.FeedCategoryAdd(category); err != nil {
		return Category{}, err
	}

	return category, nil
}

// CategoryBySlug returns the category for a given user and slug.
func (s *Service) CategoryBySlug(userUUID string, slug string) (Category, error) {
	category := Category{Slug: slug}

	if err := category.ValidateSlug(); err != nil {
		return Category{}, err
	}

	return s.r.FeedCategoryGetBySlug(userUUID, slug)
}

// CategoryByUUID returns the category for a given user and UUID.
func (s *Service) CategoryByUUID(userUUID string, categoryUUID string) (Category, error) {
	category := Category{UUID: categoryUUID}

	if err := category.ValidateUUID(); err != nil {
		return Category{}, err
	}

	return s.r.FeedCategoryGetByUUID(userUUID, categoryUUID)
}

// Categories returns all categories for a given user.
func (s *Service) Categories(userUUID string) ([]Category, error) {
	return s.r.FeedCategoryGetMany(userUUID)
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

// Subscribe creates a new Feed if needed, and adds the corresponding Subscription
// for a given user.
func (s *Service) Subscribe(userUUID string, categoryUUID string, feedURL string) error {
	feed, _, err := s.getOrCreateFeedAndEntries(feedURL)
	if err != nil {
		return fmt.Errorf("failed to create or retrieve feed: %w", err)
	}

	subscription, err := NewSubscription(categoryUUID, feed.UUID, userUUID)
	if err != nil {
		return fmt.Errorf("failed to create subscription: %w", err)
	}

	if err := s.createSubscription(subscription); err != nil {
		return fmt.Errorf("failed to create subscription: %w", err)
	}

	return nil
}

func (s *Service) SubscriptionByFeed(userUUID string, feedUUID string) (Subscription, error) {
	return s.r.FeedSubscriptionGetByFeed(userUUID, feedUUID)
}

func (s *Service) createEntries(feedUUID string, items []*gofeed.Item) ([]Entry, error) {
	var entries []Entry
	now := time.Now().UTC()

	for _, item := range items {
		publishedAt := now
		if item.PublishedParsed != nil {
			publishedAt = *item.PublishedParsed
		}

		updatedAt := publishedAt
		if item.UpdatedParsed != nil {
			updatedAt = *item.UpdatedParsed
		}

		entry := NewEntry(
			feedUUID,
			item.Link,
			item.Title,
			publishedAt,
			updatedAt,
		)
		if err := entry.ValidateForAddition(); err != nil {
			log.Warn().Err(err).Msg("skipping invalid entry")
			continue
		}

		entries = append(entries, entry)
	}

	n, err := s.r.FeedEntryAddMany(entries)
	if err != nil {
		return []Entry{}, err
	}
	if n != int64(len(entries)) {
		return []Entry{}, fmt.Errorf("feed: %d entries created, %d expected", n, len(entries))
	}

	return entries, nil
}

func (s *Service) createFeedAndEntries(feed Feed) (Feed, []Entry, error) {
	syndicationFeed, err := s.feedParser.ParseURL(feed.FeedURL)
	if err != nil {
		return Feed{}, []Entry{}, err
	}

	feed.Title = syndicationFeed.Title
	feed.FetchedAt = time.Now().UTC()
	feed.Normalize()

	if err := feed.ValidateForCreation(); err != nil {
		return Feed{}, []Entry{}, err
	}

	if err := s.r.FeedAdd(feed); err != nil {
		return Feed{}, []Entry{}, err
	}

	entries, err := s.createEntries(feed.UUID, syndicationFeed.Items)
	if err != nil {
		return Feed{}, []Entry{}, err
	}

	return feed, entries, nil
}

func (s *Service) getOrCreateFeedAndEntries(feedURL string) (Feed, []Entry, error) {
	newFeed, err := NewFeed(feedURL)
	if err != nil {
		return Feed{}, []Entry{}, err
	}

	if err := newFeed.ValidateURL(); err != nil {
		return Feed{}, []Entry{}, err
	}

	var feed Feed
	var entries []Entry

	// Attempt to retrieve an existing feed
	feed, err = s.r.FeedGetByURL(newFeed.FeedURL)
	if err == nil {
		// Retrieve at most 10 existing entries for this feed
		var err2 error
		entries, err2 = s.r.FeedEntryGetN(feed.UUID, 10)
		if err2 != nil {
			return Feed{}, []Entry{}, err2
		}

	} else if errors.Is(err, ErrFeedNotFound) {
		// Else, create it
		feed, entries, err = s.createFeedAndEntries(newFeed)
		if err != nil {
			return Feed{}, []Entry{}, err
		}

	} else {
		return Feed{}, []Entry{}, err
	}

	return feed, entries, nil
}

func (s *Service) createSubscription(subscription Subscription) error {
	if err := subscription.ValidateForCreation(s.r); err != nil {
		return err
	}

	return s.r.FeedSubscriptionAdd(subscription)
}
