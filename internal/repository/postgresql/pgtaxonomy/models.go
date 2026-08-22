// Copyright VirtualTam 2022, 2026
// SPDX-License-Identifier: MIT

package pgtaxonomy

import "time"

type DBTag struct {
	TagUUID  string `db:"tag_uuid"`
	UserUUID string `db:"user_uuid"`
	TagName  string `db:"tag_name"`

	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
}
