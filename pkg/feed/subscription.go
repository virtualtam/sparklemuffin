// Copyright (c) VirtualTam
// SPDX-License-Identifier: MIT

package feed

import "time"

type Subscription struct {
	UUID         string
	CategoryUUID string
	FeedUUID     string
	UserUUID     string

	CreatedAt time.Time
	UpdatedAt time.Time
}
