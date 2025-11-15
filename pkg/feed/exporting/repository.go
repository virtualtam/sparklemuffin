// Copyright (c) VirtualTam
// SPDX-License-Identifier: MIT

package exporting

import (
	"context"
)

// Repository provides access to user feed subscriptions for exporting.
type Repository interface {
	// FeedCategorySubscriptionsGetAll returns all CategorySubscriptions for a given user.
	FeedCategorySubscriptionsGetAll(ctx context.Context, userUUID string) ([]CategorySubscriptions, error)
}
