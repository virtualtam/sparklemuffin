// Copyright VirtualTam 2022, 2026
// SPDX-License-Identifier: MIT

package feed

import (
	"regexp"
	"strings"

	"github.com/virtualtam/sparklemuffin/pkg/user"
)

var (
	tagWhitespaceRegexp = regexp.MustCompile(`\s`)
)

// TagDeleteQuery represents a tag deletion for all Feed subscriptions of an authenticated user.
type TagDeleteQuery struct {
	UserUUID string
	Name     string
}

func (dq *TagDeleteQuery) normalize() {
	dq.Name = strings.TrimSpace(dq.Name)
}

func (dq *TagDeleteQuery) requireUserUUID() error {
	if dq.UserUUID == "" {
		return user.ErrUUIDRequired
	}
	return nil
}

func (dq *TagDeleteQuery) ensureNameHasNoWhitespace() error {
	if tagWhitespaceRegexp.MatchString(dq.Name) {
		return ErrTagNameContainsWhitespace
	}
	return nil
}

func (dq *TagDeleteQuery) requireName() error {
	if dq.Name == "" {
		return ErrTagNameRequired
	}
	return nil
}

// TagUpdateQuery represents a tag name update for all Feed subscriptions of an authenticated user.
type TagUpdateQuery struct {
	UserUUID    string
	CurrentName string
	NewName     string
}

func (uq *TagUpdateQuery) normalize() {
	uq.CurrentName = strings.TrimSpace(uq.CurrentName)
	uq.NewName = strings.TrimSpace(uq.NewName)
}

func (uq *TagUpdateQuery) requireUserUUID() error {
	if uq.UserUUID == "" {
		return user.ErrUUIDRequired
	}
	return nil
}

func (uq *TagUpdateQuery) ensureCurrentNameHasNoWhitespace() error {
	if tagWhitespaceRegexp.MatchString(uq.CurrentName) {
		return newValidationError("current", ErrTagNameContainsWhitespace)
	}
	return nil
}

func (uq *TagUpdateQuery) ensureNewNameHasNoWhitespace() error {
	if tagWhitespaceRegexp.MatchString(uq.NewName) {
		return newValidationError("new", ErrTagNameContainsWhitespace)
	}
	return nil
}

func (uq *TagUpdateQuery) requireCurrentName() error {
	if uq.CurrentName == "" {
		return newValidationError("current", ErrTagNameRequired)
	}
	return nil
}

func (uq *TagUpdateQuery) requireNewName() error {
	if uq.NewName == "" {
		return newValidationError("new", ErrTagNameRequired)
	}
	return nil
}
