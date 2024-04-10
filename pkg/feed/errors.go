// Copyright (c) VirtualTam
// SPDX-License-Identifier: MIT

package feed

import "errors"

var (
	// Feed
	ErrFeedNotFound             error = errors.New("feed: not found")
	ErrFeedSlugRequired         error = errors.New("feed: slug required")
	ErrFeedTitleRequired        error = errors.New("feed: title required")
	ErrFeedURLInvalid           error = errors.New("feed: invalid URL")
	ErrFeedURLNoScheme          error = errors.New("feed: missing URL scheme")
	ErrFeedURLNoHost            error = errors.New("feed: missing URL host")
	ErrFeedURLUnsupportedScheme error = errors.New("feed: unsupported URL scheme")

	// Feed Subscription
	ErrFeedSubscriptionAlreadyRegistered error = errors.New("feed: subscription already registered")
)
