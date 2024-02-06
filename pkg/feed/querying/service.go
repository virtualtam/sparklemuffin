// Copyright (c) VirtualTam
// SPDX-License-Identifier: MIT

package querying

// Service handles oprtaions related to displaying and paginating feeds.
type Service struct {
	r Repository
}

// NewService initializes and returns a new Service.
func NewService(r Repository) *Service {
	return &Service{
		r: r,
	}
}

// FeedsByPage returns a Page containing a limited and offset number of feeds.
func (s *Service) FeedsByPage(userUUID string) (FeedPage, error) {
	feedCategories, err := s.r.FeedGetCategories(userUUID)
	if err != nil {
		// TODO error handling
		return FeedPage{}, err
	}

	// TODO: paginate results
	// TODO: query by category
	// TODO: query by feed
	feedEntries, err := s.r.FeedGetEntriesByPage(userUUID)
	if err != nil {
		// TODO error handling
		return FeedPage{}, err
	}

	return FeedPage{
		Categories:  feedCategories,
		FeedEntries: feedEntries,
	}, nil
}