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
	ErrCategoryUUIDRequired error = errors.New("category: UUID required")

	// Feed Subscription
	ErrSubscriptionAlreadyRegistered error = errors.New("subscription: already registered")
	ErrSubscriptionUUIDRequired      error = errors.New("subscription: UUID required")

	// Feed user
	ErrUserUUIDRequired error = errors.New("user: UUID required")
)
