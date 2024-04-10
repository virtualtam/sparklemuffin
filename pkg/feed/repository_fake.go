// Copyright (c) VirtualTam
// SPDX-License-Identifier: MIT

package feed

var _ Repository = &fakeRepository{}

type fakeRepository struct {
	Feeds         []Feed
	Subscriptions []Subscription
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

func (f *fakeRepository) FeedIsSubscriptionRegistered(userUUID string, feedUUID string) (bool, error) {
	for _, s := range f.Subscriptions {
		if s.UserUUID == userUUID && s.FeedUUID == feedUUID {
			return true, nil
		}
	}

	return false, nil
}

func (r *fakeRepository) FeedSubscriptionCreate(subscription Subscription) error {
	r.Subscriptions = append(r.Subscriptions, subscription)
	return nil
}
