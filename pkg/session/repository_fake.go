// Copyright (c) VirtualTam
// SPDX-License-Identifier: MIT

package session

import (
	"context"
)

var _ Repository = &FakeRepository{}

type FakeRepository struct {
	Sessions []Session
}

func (r *FakeRepository) SessionAdd(_ context.Context, session Session) error {
	r.Sessions = append(r.Sessions, session)
	return nil
}

func (r *FakeRepository) SessionGetByRememberTokenHash(_ context.Context, hash string) (Session, error) {
	for _, s := range r.Sessions {
		if s.RememberTokenHash == hash {
			return s, nil
		}
	}
	return Session{}, ErrNotFound
}
