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

// Service handles operations related to managing Tags.
type Service struct {
	r Repository
}

// NewService initializes and returns a new Service.
func NewService(r Repository) *Service {
	return &Service{
		r: r,
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

// AllTags returns every tag for a given user, unpaginated.
func (s *Service) AllTags(ctx context.Context, userUUID string) ([]Tag, error) {
	count, err := s.r.TagGetCount(ctx, userUUID)
	if err != nil {
		return nil, err
	}
	if count == 0 {
		return []Tag{}, nil
	}

	return s.r.TagGetN(ctx, userUUID, count, 0)
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

// RenameTag renames a tag for a given user, and returns the resulting tag:
// the renamed tag itself, or, if a tag with the new name already existed and
// the two were merged via Repository.MergeTag, that existing tag.
func (s *Service) RenameTag(ctx context.Context, uq TagUpdateQuery) (Tag, error) {
	uq.Normalize()

	if err := uq.Validate(); err != nil {
		return Tag{}, err
	}

	currentTag, err := s.r.TagGetByName(ctx, uq.UserUUID, uq.CurrentName)
	if errors.Is(err, ErrNotFound) {
		return Tag{}, nil
	}
	if err != nil {
		return Tag{}, err
	}

	if uq.NewName == uq.CurrentName {
		return currentTag, nil
	}

	existingTag, err := s.r.TagGetByName(ctx, uq.UserUUID, uq.NewName)
	if err != nil && !errors.Is(err, ErrNotFound) {
		return Tag{}, err
	}

	if err == nil {
		// The new name matches the name of an existing tag: merge them.
		if err := s.r.MergeTag(ctx, uq.UserUUID, currentTag.UUID, existingTag.UUID); err != nil {
			return Tag{}, err
		}
		return existingTag, nil
	}

	if err := s.r.TagRename(ctx, uq.UserUUID, currentTag.UUID, uq.NewName); err != nil {
		return Tag{}, err
	}

	currentTag.Name = uq.NewName
	return currentTag, nil
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

// TagByUUID returns the tag with a given UUID for a given user.
func (s *Service) TagByUUID(ctx context.Context, userUUID, tagUUID string) (Tag, error) {
	tag := Tag{UUID: tagUUID}

	if err := tag.validateUUID(); err != nil {
		return Tag{}, err
	}

	return s.r.TagGetByUUID(ctx, userUUID, tagUUID)
}
