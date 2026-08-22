// Copyright VirtualTam 2022, 2026
// SPDX-License-Identifier: MIT

package bookmark

import (
	"context"
	"time"
)

// Service handles operations for the bookmark domain.
type Service struct {
	r Repository
}

// NewService initializes and returns a bookmark Service.
func NewService(r Repository) *Service {
	return &Service{
		r: r,
	}
}

// Add creates a new bookmark.
func (s *Service) Add(ctx context.Context, bookmark Bookmark) error {
	now := time.Now().UTC()

	bookmark.generateUID()
	bookmark.CreatedAt = now
	bookmark.UpdatedAt = now

	bookmark.Normalize()

	if err := bookmark.ValidateForAddition(ctx, s.r); err != nil {
		return err
	}

	return s.r.BookmarkAdd(ctx, bookmark)
}

// All returns all bookmarks for a given user.
func (s *Service) All(ctx context.Context, userUUID string) ([]Bookmark, error) {
	return s.r.BookmarkGetAll(ctx, userUUID)
}

// ByUID returns a bookmark for a given user and UID.
func (s *Service) ByUID(ctx context.Context, userUUID string, uid string) (Bookmark, error) {
	b := &Bookmark{
		UserUUID: userUUID,
		UID:      uid,
	}

	fns := []func() error{
		b.requireUID,
		b.validateUID,
		b.requireUserUUID,
	}

	for _, fn := range fns {
		if err := fn(); err != nil {
			return Bookmark{}, err
		}
	}

	return s.r.BookmarkGetByUID(ctx, userUUID, uid)
}

// ByURL returns a bookmark for a given user and URL.
func (s *Service) ByURL(ctx context.Context, userUUID string, u string) (Bookmark, error) {
	b := &Bookmark{
		UserUUID: userUUID,
		URL:      u,
	}

	b.Normalize()

	fns := []func() error{
		b.requireURL,
		b.requireUserUUID,
	}

	for _, fn := range fns {
		if err := fn(); err != nil {
			return Bookmark{}, err
		}
	}

	return s.r.BookmarkGetByURL(ctx, userUUID, b.URL)
}

// Delete permanently deletes a bookmark.
func (s *Service) Delete(ctx context.Context, userUUID, uid string) error {
	b := Bookmark{
		UserUUID: userUUID,
		UID:      uid,
	}

	fns := []func() error{
		b.requireUID,
		b.validateUID,
		b.requireUserUUID,
	}

	for _, fn := range fns {
		if err := fn(); err != nil {
			return err
		}
	}

	return s.r.BookmarkDelete(ctx, userUUID, uid)
}

// Update updates all data for an existing bookmark.
func (s *Service) Update(ctx context.Context, bookmark Bookmark) error {
	now := time.Now().UTC()
	bookmark.UpdatedAt = now

	bookmark.Normalize()

	if err := bookmark.ValidateForUpdate(ctx, s.r); err != nil {
		return err
	}

	return s.r.BookmarkUpdate(ctx, bookmark)
}
