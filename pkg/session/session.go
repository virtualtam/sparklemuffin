// Copyright (c) VirtualTam
// SPDX-License-Identifier: MIT

package session

import "time"

// Session represents a Web User session.
type Session struct {
	UserUUID               string
	RememberToken          string
	RememberTokenHash      string
	RememberTokenExpiresAt time.Time
}
