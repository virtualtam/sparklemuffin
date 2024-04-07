// Copyright (c) VirtualTam
// SPDX-License-Identifier: MIT

package feed

// Repository provides access to user feeds.
type Repository interface {
	// FeedGetCategories returns all categories for a giver user.
	FeedGetCategories(userUUID string) ([]Category, error)
}
