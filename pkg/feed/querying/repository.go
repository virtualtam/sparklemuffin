// Copyright (c) VirtualTam
// SPDX-License-Identifier: MIT

package querying

type Repository interface {
	FeedGetCategories(userUUID string) ([]Category, error)

	FeedGetEntriesByPage(userUUID string) (FeedEntries, error)
}
