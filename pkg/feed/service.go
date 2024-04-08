// Copyright (c) VirtualTam
// SPDX-License-Identifier: MIT

package feed

import (
	"errors"
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

func (s *Service) createFeed(feed Feed) (Feed, error) {
	syndicationFeed, err := s.feedParser.ParseURL(feed.URL)
	if err != nil {
		return Feed{}, err
	}

	feed.Title = syndicationFeed.Title

	if err := s.r.FeedCreate(feed); err != nil {
		return Feed{}, err
	}

	return s.r.FeedGetByURL(feed.URL)
}

func (s *Service) getOrCreateFeed(feedURL string) (Feed, error) {
	feed, err := NewFeed(feedURL)
	if err != nil {
		return Feed{}, err
	}

	if err := feed.ValidateURL(); err != nil {
		return Feed{}, err
	}

	// Attempt to retrieve an existing feed
	newFeed, err := s.r.FeedGetByURL(feed.URL)
	if errors.Is(err, ErrFeedNotFound) {
		// Else, create it
		newFeed, err = s.createFeed(feed)
		if err != nil {
			return Feed{}, err
		}

	} else if err != nil {
		return Feed{}, err
	}

	return newFeed, nil
}

// Subscribe creates a new Feed if needed, and adds the corresponding Subscription
// for a given user.
func (s *Service) Subscribe(userUUID string, CategoryUUID string, feedURL string) error {
	_, err := s.getOrCreateFeed(feedURL)
	if err != nil {
		return err
	}

	// TODO instantiate subscription, validate, save
	// TODO fetch entries
	// TODO sync entry statuses (at most 10)

	return nil
}
