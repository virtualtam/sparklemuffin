// Copyright (c) VirtualTam
// SPDX-License-Identifier: MIT

package feed

import (
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/virtualtam/sparklemuffin/internal/test/assert"
	"github.com/virtualtam/sparklemuffin/pkg/user"
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
		return user.ErrUUIDRequired
	}
	return nil
}

func (s *Subscription) requireUUID() error {
	if s.UUID == "" {
		return ErrSubscriptionUUIDRequired
	}
	return nil
}

func AssertSubscriptionsEqual(t *testing.T, want []Subscription, got []Subscription) {
	t.Helper()

	if len(got) != len(want) {
		t.Fatalf("want %d subscriptions, got %d", len(want), len(got))
	}

	for i, wantSubscription := range want {
		AssertSubscriptionEquals(t, wantSubscription, got[i])
	}
}

func AssertSubscriptionEquals(t *testing.T, want Subscription, got Subscription) {
	t.Helper()

	if want.UUID != got.UUID {
		t.Errorf("want UUID %q, got %q", want.UUID, got.UUID)
	}
	if want.CategoryUUID != got.CategoryUUID {
		t.Errorf("want CategoryUUID %q, got %q", want.CategoryUUID, got.CategoryUUID)
	}
	if want.FeedUUID != got.FeedUUID {
		t.Errorf("want FeedUUID %q, got %q", want.FeedUUID, got.FeedUUID)
	}
	if want.UserUUID != got.UserUUID {
		t.Errorf("want UserUUID %q, got %q", want.UserUUID, got.UserUUID)
	}

	assert.TimeAlmostEquals(t, "CreatedAt", got.CreatedAt, want.CreatedAt, assert.TimeComparisonDelta)
	assert.TimeAlmostEquals(t, "UpdatedAt", got.UpdatedAt, want.UpdatedAt, assert.TimeComparisonDelta)
}
