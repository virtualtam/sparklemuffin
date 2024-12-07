// Copyright (c) VirtualTam
// SPDX-License-Identifier: MIT

package pguser

import "time"

type DBUser struct {
	UUID         string `db:"uuid"`
	Email        string `db:"email"`
	NickName     string `db:"nick_name"`
	DisplayName  string `db:"display_name"`
	PasswordHash string `db:"password_hash"`
	IsAdmin      bool   `db:"is_admin"`

	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
}
