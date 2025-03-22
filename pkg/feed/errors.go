// Copyright (c) VirtualTam
// SPDX-License-Identifier: MIT

package feed

import "errors"

var (
	ErrFeedHashRequired         error = errors.New("feed: hash required")
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

	ErrCategoryAlreadyRegistered error = errors.New("category: already registered")
	ErrCategoryNameRequired      error = errors.New("category: name required")
	ErrCategoryNotFound          error = errors.New("category: Not Found")
	ErrCategorySlugInvalid       error = errors.New("category: invalid slug")
	ErrCategorySlugRequired      error = errors.New("category: slug required")
	ErrCategoryUserUUIDRequired  error = errors.New("category: UserUUID required")
	ErrCategoryUUIDInvalid       error = errors.New("category: invalid UUID")
	ErrCategoryUUIDRequired      error = errors.New("category: UUID required")

	ErrEntryNotFound               error = errors.New("entry: not found")
	ErrEntryTitleRequired          error = errors.New("entry: title required")
	ErrEntryPublishedAtInTheFuture error = errors.New("entry: publication date is in the future")
	ErrEntryPublishedAtIsZero      error = errors.New("entry: publication date is zero")
	ErrEntryUpdatedAtInTheFuture   error = errors.New("entry: update date is in the future")
	ErrEntryUpdatedAtIsZero        error = errors.New("entry: update date is zero")
	ErrEntryURLInvalid             error = errors.New("entry: invalid URL")
	ErrEntryURLNoScheme            error = errors.New("entry: missing URL scheme")
	ErrEntryURLNoHost              error = errors.New("entry: missing URL host")
	ErrEntryURLRequired            error = errors.New("entry: URL required")
	ErrEntryURLUnsupportedScheme   error = errors.New("entry: unsupported URL scheme")
	ErrEntryUIDInvalid             error = errors.New("entry: invalid UID")

	ErrEntryMetadataNotFound error = errors.New("entry-metadata: not found")

	ErrSubscriptionAlreadyRegistered error = errors.New("subscription: already registered")
	ErrSubscriptionNotFound          error = errors.New("subscription: not found")
	ErrSubscriptionUUIDRequired      error = errors.New("subscription: UUID required")
)
