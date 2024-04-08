// Copyright (c) VirtualTam
// SPDX-License-Identifier: MIT

package feed

var _ Repository = &fakeRepository{}

type fakeRepository struct {
	Feeds []Feed
}

func (r *fakeRepository) FeedCreate(feed Feed) error {
	r.Feeds = append(r.Feeds, feed)
	return nil
}

func (r *fakeRepository) FeedGetByURL(feedURL string) (Feed, error) {
	for _, f := range r.Feeds {
		if f.URL == feedURL {
			return f, nil
		}
	}

	return Feed{}, ErrFeedNotFound
}

func (r *fakeRepository) FeedGetCategories(userUUID string) ([]Category, error) {
	panic("unimplemented")
}
