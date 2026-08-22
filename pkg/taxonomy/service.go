// Copyright VirtualTam 2022, 2026
// SPDX-License-Identifier: MIT

package taxonomy

import (
	"context"
	"errors"
	"strings"

	"github.com/virtualtam/sparklemuffin/internal/paginate"
	"github.com/virtualtam/sparklemuffin/pkg/user"
)

// OnTagMergeFn reassigns a domain's tag associations from one tag UUID to another.
type OnTagMergeFn func(ctx context.Context, userUUID, oldTagUUID, newTagUUID string) error

// Service handles operations related to managing Tags.
type Service struct {
	r             Repository
	onTagMergeFns []OnTagMergeFn
}

// NewService initializes and returns a new Service.
func NewService(r Repository, onTagMergeFns ...OnTagMergeFn) *Service {
	return &Service{
		r:             r,
		onTagMergeFns: onTagMergeFns,
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

// ListTags returns a Page containing a limited and offset number of tags for a given user.
func (s *Service) ListTags(ctx context.Context, userUUID string, number uint) (TagPage, error) {
	if number < 1 {
		return TagPage{}, paginate.ErrPageNumberOutOfBounds
	}

	tagCount, err := s.r.TagGetCount(ctx, userUUID)
	if err != nil {
		return TagPage{}, err
	}

	totalPages := paginate.PageCount(tagCount, tagsPerPage)

	if number > totalPages {
		return TagPage{}, paginate.ErrPageNumberOutOfBounds
	}

	if tagCount == 0 {
		return NewTagPage(1, 1, 0, []Tag{}), nil
	}

	offset := (number - 1) * tagsPerPage

	tags, err := s.r.TagGetN(ctx, userUUID, tagsPerPage, offset)
	if err != nil {
		return TagPage{}, err
	}

	return NewTagPage(number, totalPages, tagCount, tags), nil
}

// RenameTag renames a tag for a given user.
//
// If a tag with the new name already exists, the two tags are merged: every
// registered OnTagMergeFn is called to move its associations from the
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
		// The new name matches the name of an existing tag.

		// 1. Update all references to point to its UUID.
		for _, onTagMergeFn := range s.onTagMergeFns {
			if err := onTagMergeFn(ctx, uq.UserUUID, currentTag.UUID, existingTag.UUID); err != nil {
				return err
			}
		}

		// 2. Delete the old tag.
		return s.r.TagDelete(ctx, uq.UserUUID, uq.CurrentName)
	}

	return s.r.TagRename(ctx, uq.UserUUID, currentTag.UUID, uq.NewName)
}

// SearchTags returns a Page containing a limited and offset number of tags
// for a given user and search terms.
func (s *Service) SearchTags(ctx context.Context, userUUID, searchTerms string, number uint) (TagPage, error) {
	if number < 1 {
		return TagPage{}, paginate.ErrPageNumberOutOfBounds
	}

	tagCount, err := s.r.TagSearchCount(ctx, userUUID, searchTerms)
	if err != nil {
		return TagPage{}, err
	}

	totalPages := paginate.PageCount(tagCount, tagsPerPage)

	if number > totalPages {
		return TagPage{}, paginate.ErrPageNumberOutOfBounds
	}

	if tagCount == 0 {
		return NewTagSearchResultPage(searchTerms, 0, 1, 1, []Tag{}), nil
	}

	offset := (number - 1) * tagsPerPage

	tags, err := s.r.TagSearchN(ctx, userUUID, searchTerms, tagsPerPage, offset)
	if err != nil {
		return TagPage{}, err
	}

	return NewTagSearchResultPage(searchTerms, tagCount, number, totalPages, tags), nil
}
