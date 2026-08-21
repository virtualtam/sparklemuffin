// Copyright VirtualTam 2022, 2026
// SPDX-License-Identifier: MIT

package taxonomy

import (
	"context"
	"errors"
	"strings"

	"github.com/virtualtam/sparklemuffin/pkg/user"
)

// OnTagRenameFn reassigns a domain's tag associations from one tag UUID to another.
type OnTagRenameFn func(ctx context.Context, userUUID, oldTagUUID, newTagUUID string) error

// Service handles operations related to managing Tags.
type Service struct {
	r              Repository
	onTagRenameFns []OnTagRenameFn
}

// NewService initializes and returns a new Service.
func NewService(r Repository, onTagRenameFns ...OnTagRenameFn) *Service {
	return &Service{
		r:              r,
		onTagRenameFns: onTagRenameFns,
	}
}

// AddTag adds a new Tag for a given user.
func (s *Service) AddTag(ctx context.Context, userUUID, name string) (Tag, error) {
	tag, err := NewTag(userUUID, name)
	if err != nil {
		return Tag{}, err
	}

	if err := tag.ValidateForAddition(ctx, s.r); err != nil {
		return Tag{}, err
	}

	if err := s.r.TagAdd(ctx, tag); err != nil {
		return Tag{}, err
	}

	return tag, nil
}

// DeleteTag deletes a tag for a given user.
func (s *Service) DeleteTag(ctx context.Context, dq TagDeleteQuery) error {
	dq.Normalize()

	if err := dq.Validate(); err != nil {
		return err
	}

	return s.r.TagDelete(ctx, dq.UserUUID, dq.Name)
}

// GetOrCreateTags returns the tags for a given user matching the given
// names, in the same order, creating any tag that does not already exist.
func (s *Service) GetOrCreateTags(ctx context.Context, userUUID string, names []string) ([]Tag, error) {
	if userUUID == "" {
		return nil, user.ErrUUIDRequired
	}

	tags := make([]Tag, 0, len(names))

	for _, name := range names {
		tag, err := s.r.TagGetByName(ctx, userUUID, strings.TrimSpace(name))

		if errors.Is(err, ErrNotFound) {
			tag, err = s.AddTag(ctx, userUUID, name)
		}

		if err != nil {
			return nil, err
		}

		tags = append(tags, tag)
	}

	return tags, nil
}

// RenameTag renames a tag for a given user.
//
// If a tag with the new name already exists, the two tags are merged: every
// registered OnTagRenameFn is called to move its associations from the
// current tag to the existing one, which is then deleted.
func (s *Service) RenameTag(ctx context.Context, uq TagUpdateQuery) error {
	uq.Normalize()

	if err := uq.Validate(); err != nil {
		return err
	}

	if uq.NewName == uq.CurrentName {
		return nil
	}

	currentTag, err := s.r.TagGetByName(ctx, uq.UserUUID, uq.CurrentName)
	if errors.Is(err, ErrNotFound) {
		return nil
	}
	if err != nil {
		return err
	}

	existingTag, err := s.r.TagGetByName(ctx, uq.UserUUID, uq.NewName)
	if err != nil && !errors.Is(err, ErrNotFound) {
		return err
	}

	if err == nil {
		for _, onTagRenameFn := range s.onTagRenameFns {
			if err := onTagRenameFn(ctx, uq.UserUUID, currentTag.UUID, existingTag.UUID); err != nil {
				return err
			}
		}

		return s.r.TagDelete(ctx, uq.UserUUID, uq.CurrentName)
	}

	return s.r.TagRename(ctx, uq.UserUUID, currentTag.UUID, uq.NewName)
}
