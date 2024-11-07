// Copyright (c) VirtualTam
// SPDX-License-Identifier: MIT

package exporting

var _ Repository = &fakeRepository{}

type fakeRepository struct {
	categoriesSubscriptions []CategorySubscriptions
}

func (r *fakeRepository) FeedCategorySubscriptionsGetAll(userUUID string) ([]CategorySubscriptions, error) {
	return r.categoriesSubscriptions, nil
}
