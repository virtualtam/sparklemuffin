// Copyright (c) VirtualTam
// SPDX-License-Identifier: MIT

package feed

import (
	"time"
)

type Category struct {
	UUID     string
	UserUUID string

	Name string
	Slug string

	CreatedAt time.Time
	UpdatedAt time.Time
}
