// Copyright (c) VirtualTam
// SPDX-License-Identifier: MIT

package feed

import (
	"time"

	"github.com/google/uuid"
)

type Subscription struct {
	UUID         string
	CategoryUUID string
	FeedUUID     string
	UserUUID     string

	CreatedAt time.Time
	UpdatedAt time.Time
}

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

func (s *Subscription) ValidateForCreation(v ValidationRepository) error {
	if err := s.ensureSubscriptionIsNotRegistered(v); err != nil {
		return err
	}

	return nil
}

func (s *Subscription) ensureSubscriptionIsNotRegistered(v ValidationRepository) error {
	registered, err := v.FeedIsSubscriptionRegistered(s.UserUUID, s.FeedUUID)
	if err != nil {
		return err
	}

	if registered {
		return ErrFeedSubscriptionAlreadyRegistered
	}

	return nil
}
