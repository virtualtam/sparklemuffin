// Copyright (c) VirtualTam
// SPDX-License-Identifier: MIT

package feed

import "errors"

var (
	// Feed
	ErrFeedNotFound             error = errors.New("feed: not found")
	ErrFeedSlugInvalid          error = errors.New("feed: invalid slug")
	ErrFeedSlugRequired         error = errors.New("feed: slug required")
	ErrFeedTitleRequired        error = errors.New("feed: title required")
	ErrFeedUUIDInvalid          error = errors.New("feed: invalid UUID")
	ErrFeedUUIDRequired         error = errors.New("feed: UUID required")
	ErrFeedURLInvalid           error = errors.New("feed: invalid URL")
	ErrFeedURLNoScheme          error = errors.New("feed: missing URL scheme")
	ErrFeedURLNoHost            error = errors.New("feed: missing URL host")
	ErrFeedURLRequired          error = errors.New("feed: URL required")
	ErrFeedURLUnsupportedScheme error = errors.New("feed: unsupported URL scheme")

	// Feed Category
	ErrCategoryAlreadyRegistered error = errors.New("category: already registered")
	ErrCategoryNameRequired      error = errors.New("category: name required")
	ErrCategoryNotFound          error = errors.New("category: Not Found")
	ErrCategorySlugInvalid       error = errors.New("category: invalid slug")
	ErrCategorySlugRequired      error = errors.New("category: slug required")
	ErrCategoryUserUUIDRequired  error = errors.New("category: UserUUID required")
	ErrCategoryUUIDInvalid       error = errors.New("category: invalid UUID")
	ErrCategoryUUIDRequired      error = errors.New("category: UUID required")

	// Feed Entry
	ErrEntryNotFound             error = errors.New("entry: not found")
	ErrEntryTitleRequired        error = errors.New("entry: title required")
	ErrEntryURLInvalid           error = errors.New("entry: invalid URL")
	ErrEntryURLNoScheme          error = errors.New("entry: missing URL scheme")
	ErrEntryURLNoHost            error = errors.New("entry: missing URL host")
	ErrEntryURLRequired          error = errors.New("entry: URL required")
	ErrEntryURLUnsupportedScheme error = errors.New("entry: unsupported URL scheme")
	ErrEntryUUIDInvalid          error = errors.New("entry: invalid UUID")

	// Feed Entry Metadata
	ErrEntryMetadataNotFound error = errors.New("entry-metadata: not found")

	// Feed Subscription
	ErrSubscriptionAlreadyRegistered error = errors.New("subscription: already registered")
	ErrSubscriptionNotFound          error = errors.New("subscription: not found")
	ErrSubscriptionUUIDRequired      error = errors.New("subscription: UUID required")

	// Feed user
	ErrUserUUIDRequired error = errors.New("user: UUID required")
)
