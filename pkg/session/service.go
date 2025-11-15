// Copyright (c) VirtualTam
// SPDX-License-Identifier: MIT

package session

import (
	"context"

	"github.com/virtualtam/sparklemuffin/internal/hash"
	"github.com/virtualtam/sparklemuffin/pkg/user"
)

// Service handles operations for the user session domain.
type Service struct {
	r    Repository
	hmac *hash.HMAC
}

// NewService initializes and returns a Session Service.
func NewService(r Repository, hmacKey string) *Service {
	hmac := hash.NewHMAC(hmacKey)
	return &Service{
		r:    r,
		hmac: hmac,
	}
}

// Add saves a new Session.
func (s *Service) Add(ctx context.Context, session Session) error {
	err := s.runValidationFuncs(
		&session,
		s.requireUserUUID,
		s.requireRememberToken,
		s.hashRememberToken,
		s.requireRememberTokenHash,
	)
	if err != nil {
		return err
	}

	return s.r.SessionAdd(ctx, session)
}

// ByRememberToken returns the user Session corresponding to a given RememberToken.
func (s *Service) ByRememberToken(ctx context.Context, rememberToken string) (Session, error) {
	session := Session{RememberToken: rememberToken}

	err := s.runValidationFuncs(
		&session,
		s.requireRememberToken,
		s.hashRememberToken,
		s.requireRememberTokenHash,
	)
	if err != nil {
		return Session{}, err
	}

	return s.r.SessionGetByRememberTokenHash(ctx, session.RememberTokenHash)
}

func (s *Service) hashRememberToken(session *Session) error {
	if session.RememberToken == "" {
		return nil
	}

	h, err := s.hmac.Hash(session.RememberToken)
	if err != nil {
		return err
	}

	session.RememberTokenHash = h

	return nil
}

func (s *Service) requireRememberToken(session *Session) error {
	if session.RememberToken == "" {
		return ErrRememberTokenRequired
	}
	return nil
}

func (s *Service) requireRememberTokenHash(session *Session) error {
	if session.RememberToken == "" {
		return ErrRememberTokenHashRequired
	}
	return nil
}

func (s *Service) requireUserUUID(session *Session) error {
	if session.UserUUID == "" {
		return user.ErrUUIDRequired
	}
	return nil
}

// validationFunc defines a function that can be applied to normalize or
// validate Session data.
type validationFunc func(*Session) error

// runValidationFuncs applies Session normalization and validation functions and
// stops at the first encountered error.
func (s *Service) runValidationFuncs(session *Session, fns ...validationFunc) error {
	for _, fn := range fns {
		if err := fn(session); err != nil {
			return err
		}
	}
	return nil
}
