// Copyright (c) VirtualTam
// SPDX-License-Identifier: MIT

package feed

var _ Repository = &fakeRepository{}

type fakeRepository struct {
	Categories    []Category
	Entries       []Entry
	Feeds         []Feed
	Subscriptions []Subscription
}

func (r *fakeRepository) FeedAdd(feed Feed) error {
	r.Feeds = append(r.Feeds, feed)
	return nil
}

func (r *fakeRepository) FeedGetByURL(feedURL string) (Feed, error) {
	for _, f := range r.Feeds {
		if f.FeedURL == feedURL {
			return f, nil
		}
	}

	return Feed{}, ErrFeedNotFound
}

func (r *fakeRepository) FeedCategoryGetBySlug(userUUID string, slug string) (Category, error) {
	for _, category := range r.Categories {
		if category.UserUUID == userUUID && category.Slug == slug {
			return category, nil
		}
	}

	return Category{}, ErrCategoryNotFound
}

func (r *fakeRepository) FeedCategoryGetMany(userUUID string) ([]Category, error) {
	panic("unimplemented")
}

func (r *fakeRepository) FeedCategoryAdd(category Category) error {
	r.Categories = append(r.Categories, category)
	return nil
}

func (r *fakeRepository) FeedCategoryIsRegistered(userUUID string, name string, slug string) (bool, error) {
	for _, category := range r.Categories {
		if category.UserUUID != userUUID {
			continue
		}

		if category.Name == name || category.Slug == slug {
			return true, nil
		}
	}

	return false, nil
}

func (r *fakeRepository) FeedEntryAddMany(entries []Entry) (int64, error) {
	r.Entries = append(r.Entries, entries...)
	return int64(len(entries)), nil
}

func (r *fakeRepository) FeedEntryGetN(feedUUID string, n uint) ([]Entry, error) {
	var entries []Entry
	var count uint

	for _, entry := range r.Entries {
		if entry.FeedUUID != feedUUID {
			continue
		}

		count++
		entries = append(entries, entry)

		if count == n {
			break
		}
	}

	return entries, nil
}

func (r *fakeRepository) FeedSubscriptionIsRegistered(userUUID string, feedUUID string) (bool, error) {
	for _, s := range r.Subscriptions {
		if s.UserUUID == userUUID && s.FeedUUID == feedUUID {
			return true, nil
		}
	}

	return false, nil
}

func (r *fakeRepository) FeedSubscriptionAdd(subscription Subscription) error {
	r.Subscriptions = append(r.Subscriptions, subscription)
	return nil
}
