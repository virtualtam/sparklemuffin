// Copyright VirtualTam 2022, 2026
// SPDX-License-Identifier: MIT

package taxonomy

import (
	"context"
	"regexp"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/virtualtam/sparklemuffin/pkg/user"
)

var tagWhitespaceRegexp = regexp.MustCompile(`\s`)

// Tag represents a user-defined label that can be attached to entities in
// other domains (e.g. bookmarks, feed subscriptions).
type Tag struct {
	UUID     string
	UserUUID string

	Name string

	CreatedAt time.Time
	UpdatedAt time.Time
}

// NewTag initializes and returns a new Tag.
func NewTag(userUUID, name string) (Tag, error) {
	now := time.Now().UTC()

	generatedUUID, err := uuid.NewRandom()
	if err != nil {
		return Tag{}, err
	}

	tag := Tag{
		UUID:      generatedUUID.String(),
		UserUUID:  userUUID,
		Name:      name,
		CreatedAt: now,
		UpdatedAt: now,
	}

	tag.Normalize()

	return tag, nil
}

// Normalize sanitizes and normalizes all fields.
func (t *Tag) Normalize() {
	t.Name = strings.TrimSpace(t.Name)
}

// ValidateForAddition ensures mandatory fields are properly set when adding a Tag.
func (t *Tag) ValidateForAddition(ctx context.Context, v ValidationRepository) error {
	fns := []func() error{
		t.requireUserUUID,
		t.requireName,
		t.ensureNameHasNoWhitespace,
		t.ensureNameIsNotRegistered(ctx, v),
	}

	for _, fn := range fns {
		if err := fn(); err != nil {
			return err
		}
	}

	return nil
}

func (t *Tag) requireUserUUID() error {
	if t.UserUUID == "" {
		return user.ErrUUIDRequired
	}
	return nil
}

func (t *Tag) requireName() error {
	if t.Name == "" {
		return ErrTagNameRequired
	}
	return nil
}

func (t *Tag) ensureNameHasNoWhitespace() error {
	if tagWhitespaceRegexp.MatchString(t.Name) {
		return ErrTagNameContainsWhitespace
	}
	return nil
}

func (t *Tag) ensureNameIsNotRegistered(ctx context.Context, v ValidationRepository) func() error {
	return func() error {
		registered, err := v.TagNameIsRegistered(ctx, t.UserUUID, t.Name)
		if err != nil {
			return err
		}

		if registered {
			return ErrTagAlreadyRegistered
		}

		return nil
	}
}

// TagDeleteQuery represents a tag deletion for a given user.
type TagDeleteQuery struct {
	UserUUID string
	Name     string
}

// Normalize sanitizes and normalizes all fields.
func (dq *TagDeleteQuery) Normalize() {
	dq.Name = strings.TrimSpace(dq.Name)
}

// Validate ensures mandatory fields are properly set.
func (dq *TagDeleteQuery) Validate() error {
	fns := []func() error{
		dq.requireUserUUID,
		dq.requireName,
		dq.ensureNameHasNoWhitespace,
	}

	for _, fn := range fns {
		if err := fn(); err != nil {
			return err
		}
	}

	return nil
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

// TagUpdateQuery represents a tag rename for a given user.
type TagUpdateQuery struct {
	UserUUID    string
	CurrentName string
	NewName     string
}

// Normalize sanitizes and normalizes all fields.
func (uq *TagUpdateQuery) Normalize() {
	uq.CurrentName = strings.TrimSpace(uq.CurrentName)
	uq.NewName = strings.TrimSpace(uq.NewName)
}

// Validate ensures mandatory fields are properly set.
func (uq *TagUpdateQuery) Validate() error {
	fns := []func() error{
		uq.requireUserUUID,
		uq.requireCurrentName,
		uq.ensureCurrentNameHasNoWhitespace,
		uq.requireNewName,
		uq.ensureNewNameHasNoWhitespace,
	}

	for _, fn := range fns {
		if err := fn(); err != nil {
			return err
		}
	}

	return nil
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
