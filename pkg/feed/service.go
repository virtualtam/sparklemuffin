// Copyright (c) VirtualTam
// SPDX-License-Identifier: MIT

package feed

// Service handles operations for the feed domain.
type Service struct {
	r Repository
}

// NewService initializes and returns a Feed Service.
func NewService(r Repository) *Service {
	return &Service{
		r: r,
	}
}

// Categories returns all categories for a given user.
func (s *Service) Categories(userUUID string) ([]Category, error) {
	return s.r.FeedGetCategories(userUUID)
}
