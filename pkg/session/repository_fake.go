// Copyright VirtualTam 2022, 2026
// SPDX-License-Identifier: MIT

package session

import (
	"context"
	"slices"
	"time"
)

var _ Repository = &FakeRepository{}

type FakeRepository struct {
	Sessions []Session

	SessionDeleteByUserUUIDErr error
}

func (r *FakeRepository) SessionAdd(_ context.Context, session Session) error {
	r.Sessions = append(r.Sessions, session)
	return nil
}

func (r *FakeRepository) SessionGetByRememberTokenHash(_ context.Context, hash string) (Session, error) {
	for _, s := range r.Sessions {
		if s.RememberTokenHash == hash && s.RememberTokenExpiresAt.After(time.Now()) {
			return s, nil
		}
	}
	return Session{}, ErrNotFound
}

func (r *FakeRepository) SessionDeleteByRememberTokenHash(_ context.Context, hash string) error {
	r.Sessions = slices.DeleteFunc(r.Sessions, func(s Session) bool {
		return s.RememberTokenHash == hash
	})
	return nil
}

func (r *FakeRepository) SessionDeleteByUserUUID(_ context.Context, userUUID string) error {
	if r.SessionDeleteByUserUUIDErr != nil {
		return r.SessionDeleteByUserUUIDErr
	}

	r.Sessions = slices.DeleteFunc(r.Sessions, func(s Session) bool {
		return s.UserUUID == userUUID
	})
	return nil
}
