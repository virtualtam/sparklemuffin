// Copyright (c) VirtualTam
// SPDX-License-Identifier: MIT

package session

import (
	"github.com/virtualtam/sparklemuffin/pkg/hash"
)

// Service handles operations for the user session domain.
type Service struct {
	r    Repository
	hmac *hash.HMAC
}

// NewServices initializes and returns a Session Service.
func NewService(r Repository, hmacKey string) *Service {
	hmac := hash.NewHMAC(hmacKey)
	return &Service{
		r:    r,
		hmac: hmac,
	}
}

// Add saves a new Session.
func (s *Service) Add(session Session) error {
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

	return s.r.SessionAdd(session)
}

// ByRememberToken returns the user Session corresponding to a given RememberToken.
func (s *Service) ByRememberToken(rememberToken string) (Session, error) {
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

	return s.r.SessionGetByRememberTokenHash(session.RememberTokenHash)
}

func (s *Service) hashRememberToken(session *Session) error {
	if session.RememberToken == "" {
		return nil
	}

	hash, err := s.hmac.Hash(session.RememberToken)
	if err != nil {
		return err
	}

	session.RememberTokenHash = hash

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
		return ErrUserUUIDRequired
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
