// Copyright (c) VirtualTam
// SPDX-License-Identifier: MIT

package feed

import "errors"

var (
	// Feed
	ErrFeedNotFound             error = errors.New("feed: not found")
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
	ErrCategorySlugRequired      error = errors.New("category: slug required")
	ErrCategoryUUIDRequired      error = errors.New("category: UUID required")

	// Feed Entry
	ErrEntryTitleRequired        error = errors.New("entry: title required")
	ErrEntryURLInvalid           error = errors.New("entry: invalid URL")
	ErrEntryURLNoScheme          error = errors.New("entry: missing URL scheme")
	ErrEntryURLNoHost            error = errors.New("entry: missing URL host")
	ErrEntryURLRequired          error = errors.New("entry: URL required")
	ErrEntryURLUnsupportedScheme error = errors.New("entry: unsupported URL scheme")
	ErrEntryUUIDInvalid          error = errors.New("entry: invalid UUID")

	// Feed Subscription
	ErrSubscriptionAlreadyRegistered error = errors.New("subscription: already registered")
	ErrSubscriptionUUIDRequired      error = errors.New("subscription: UUID required")

	// Feed user
	ErrUserUUIDRequired error = errors.New("user: UUID required")
)
