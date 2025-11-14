// Copyright (c) VirtualTam
// SPDX-License-Identifier: MIT

package feed

import "errors"

var (
	ErrFeedHashRequired         = errors.New("feed: hash required")
	ErrFeedNotFound             = errors.New("feed: not found")
	ErrFeedSlugInvalid          = errors.New("feed: invalid slug")
	ErrFeedSlugRequired         = errors.New("feed: slug required")
	ErrFeedTitleRequired        = errors.New("feed: title required")
	ErrFeedUUIDRequired         = errors.New("feed: UUID required")
	ErrFeedURLInvalid           = errors.New("feed: invalid URL")
	ErrFeedURLNoScheme          = errors.New("feed: missing URL scheme")
	ErrFeedURLNoHost            = errors.New("feed: missing URL host")
	ErrFeedURLRequired          = errors.New("feed: URL required")
	ErrFeedURLUnsupportedScheme = errors.New("feed: unsupported URL scheme")

	ErrCategoryAlreadyRegistered = errors.New("category: already registered")
	ErrCategoryNameRequired      = errors.New("category: name required")
	ErrCategoryNotFound          = errors.New("category: Not Found")
	ErrCategorySlugInvalid       = errors.New("category: invalid slug")
	ErrCategorySlugRequired      = errors.New("category: slug required")
	ErrCategoryUserUUIDRequired  = errors.New("category: UserUUID required")
	ErrCategoryUUIDInvalid       = errors.New("category: invalid UUID")
	ErrCategoryUUIDRequired      = errors.New("category: UUID required")

	ErrEntryNotFound               = errors.New("entry: not found")
	ErrEntryTitleRequired          = errors.New("entry: title required")
	ErrEntryPublishedAtInTheFuture = errors.New("entry: publication date is in the future")
	ErrEntryPublishedAtIsZero      = errors.New("entry: publication date is zero")
	ErrEntryUpdatedAtInTheFuture   = errors.New("entry: update date is in the future")
	ErrEntryUpdatedAtIsZero        = errors.New("entry: update date is zero")
	ErrEntryURLInvalid             = errors.New("entry: invalid URL")
	ErrEntryURLNoScheme            = errors.New("entry: missing URL scheme")
	ErrEntryURLNoHost              = errors.New("entry: missing URL host")
	ErrEntryURLRequired            = errors.New("entry: URL required")
	ErrEntryURLUnsupportedScheme   = errors.New("entry: unsupported URL scheme")
	ErrEntryUIDInvalid             = errors.New("entry: invalid UID")

	ErrEntryMetadataNotFound = errors.New("entry-metadata: not found")

	ErrPreferencesEntryVisibilityUnknown = errors.New("preferences: unknown entry visibility")

	ErrSubscriptionAlreadyRegistered = errors.New("subscription: already registered")
	ErrSubscriptionNotFound          = errors.New("subscription: not found")
	ErrSubscriptionUUIDRequired      = errors.New("subscription: UUID required")
)
