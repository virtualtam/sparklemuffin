// Copyright (c) VirtualTam
// SPDX-License-Identifier: MIT

package feed

// Repository provides access to user feeds.
type Repository interface {
	// FeedCreate creates a new Feed.
	FeedCreate(feed Feed) error

	// FeedGetByURL returns the Feed for a given URL.
	FeedGetByURL(feedURL string) (Feed, error)

	// FeedGetCategories returns all categories for a giver user.
	FeedGetCategories(userUUID string) ([]Category, error)
}
