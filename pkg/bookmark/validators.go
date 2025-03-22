// Copyright (c) VirtualTam
// SPDX-License-Identifier: MIT

package bookmark

import "github.com/virtualtam/sparklemuffin/pkg/user"

func requireUserUUID(userUUID string) error {
	if userUUID == "" {
		return user.ErrUUIDRequired
	}
	return nil
}
