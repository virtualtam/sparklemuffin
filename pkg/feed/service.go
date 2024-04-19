// Copyright (c) VirtualTam
// SPDX-License-Identifier: MIT

package feed

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/mmcdole/gofeed"
)

// Service handles operations for the feed domain.
type Service struct {
	r Repository

	feedParser *gofeed.Parser
}

// NewService initializes and returns a Feed Service.
func NewService(r Repository, client *http.Client) *Service {
	feedParser := gofeed.NewParser()
	feedParser.Client = client

	return &Service{
		r:          r,
		feedParser: feedParser,
	}
}

// Categories returns all categories for a given user.
func (s *Service) Categories(userUUID string) ([]Category, error) {
	return s.r.FeedGetCategories(userUUID)
}

// Subscribe creates a new Feed if needed, and adds the corresponding Subscription
// for a given user.
func (s *Service) Subscribe(userUUID string, categoryUUID string, feedURL string) error {
	feed, entries, err := s.getOrCreateFeedAndEntries(feedURL)
	if err != nil {
		return err
	}

	subscription, err := NewSubscription(categoryUUID, feed.UUID, userUUID)
	if err != nil {
		return err
	}

	if err := s.createSubscription(subscription); err != nil {
		return err
	}

	// TODO sync entry statuses (at most 10)
	_ = entries

	return nil
}

func (s *Service) createEntries(feedUUID string, items []*gofeed.Item) ([]Entry, error) {
	entries := make([]Entry, len(items))

	for i, item := range items {
		entry := NewEntry(
			feedUUID,
			item.Link,
			item.Title,
			*item.PublishedParsed,
			*item.UpdatedParsed,
		)
		if err := entry.ValidateForAddition(); err != nil {
			return []Entry{}, err
		}

		entries[i] = entry
	}

	n, err := s.r.FeedEntryCreateMany(entries)
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

	if err := s.r.FeedCreate(feed); err != nil {
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
		entries, err = s.r.FeedEntryGetN(feed.UUID, 10)
		if err != nil {
			return Feed{}, []Entry{}, err
		}

	} else if errors.Is(err, ErrFeedNotFound) {
		// Else, create it
		feed, entries, err = s.createFeedAndEntries(newFeed)
		if err != nil {
			return Feed{}, []Entry{}, err
		}

	} else if err != nil {
		return Feed{}, []Entry{}, err
	}

	return feed, entries, nil
}

func (s *Service) createSubscription(subscription Subscription) error {
	if err := subscription.ValidateForCreation(s.r); err != nil {
		return err
	}

	return s.r.FeedSubscriptionCreate(subscription)
}
