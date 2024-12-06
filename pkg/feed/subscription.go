// Copyright (c) VirtualTam
// SPDX-License-Identifier: MIT

package feed

import (
	"strings"
	"time"

	"github.com/google/uuid"
)

// Subscription represents a given user's subscription to a Feed.
type Subscription struct {
	UUID         string
	CategoryUUID string
	FeedUUID     string
	UserUUID     string

	Alias string

	CreatedAt time.Time
	UpdatedAt time.Time
}

// NewSubscription initializes and returns a new Subscription.
func NewSubscription(categoryUUID string, feedUUID string, userUUID string) (Subscription, error) {
	now := time.Now().UTC()

	generatedUUID, err := uuid.NewRandom()
	if err != nil {
		return Subscription{}, err
	}

	s := Subscription{
		UUID:         generatedUUID.String(),
		CategoryUUID: categoryUUID,
		FeedUUID:     feedUUID,
		UserUUID:     userUUID,
		CreatedAt:    now,
		UpdatedAt:    now,
	}

	return s, nil
}

func (s *Subscription) Normalize() {
	s.normalizeAlias()
}

func (s *Subscription) ValidateForCreation(v ValidationRepository) error {
	fns := []func() error{
		s.requireUUID,
		s.requireCategoryUUID,
		s.requireFeedUUID,
		s.requireUserUUID,
		s.ensureSubscriptionIsNotRegistered(v),
	}

	for _, fn := range fns {
		if err := fn(); err != nil {
			return err
		}
	}

	return nil
}

func (s *Subscription) ensureSubscriptionIsNotRegistered(v ValidationRepository) func() error {
	return func() error {
		registered, err := v.FeedSubscriptionIsRegistered(s.UserUUID, s.FeedUUID)
		if err != nil {
			return err
		}

		if registered {
			return ErrSubscriptionAlreadyRegistered
		}

		return nil
	}
}

func (s *Subscription) normalizeAlias() {
	s.Alias = strings.TrimSpace(s.Alias)
}

func (s *Subscription) requireCategoryUUID() error {
	if s.CategoryUUID == "" {
		return ErrCategoryUUIDRequired
	}
	return nil
}

func (s *Subscription) requireFeedUUID() error {
	if s.FeedUUID == "" {
		return ErrFeedUUIDRequired
	}
	return nil
}

func (s *Subscription) requireUserUUID() error {
	if s.UserUUID == "" {
		return ErrUserUUIDRequired
	}
	return nil
}

func (s *Subscription) requireUUID() error {
	if s.UUID == "" {
		return ErrSubscriptionUUIDRequired
	}
	return nil
}
