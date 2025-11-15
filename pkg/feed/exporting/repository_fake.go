// Copyright (c) VirtualTam
// SPDX-License-Identifier: MIT

package exporting

import (
	"context"
)

var _ Repository = &fakeRepository{}

type fakeRepository struct {
	categoriesSubscriptions []CategorySubscriptions
}

func (r *fakeRepository) FeedCategorySubscriptionsGetAll(_ context.Context, userUUID string) ([]CategorySubscriptions, error) {
	return r.categoriesSubscriptions, nil
}
