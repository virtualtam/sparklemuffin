// Copyright VirtualTam 2022, 2026
// SPDX-License-Identifier: MIT

package taxonomy

import (
	"context"
	"slices"
	"sort"
	"strings"
)

var _ Repository = &FakeRepository{}

type FakeRepository struct {
	Tags []Tag

	// MergeTagErr, if set, is returned by MergeTag without any effect.
	MergeTagErr error

	// MergeTagCalls records every call made to MergeTag.
	MergeTagCalls []MergeTagCall
}

// MergeTagCall records the arguments passed to a single MergeTag call.
type MergeTagCall struct {
	UserUUID   string
	OldTagUUID string
	NewTagUUID string
}

func (r *FakeRepository) MergeTag(_ context.Context, userUUID, oldTagUUID, newTagUUID string) error {
	r.MergeTagCalls = append(r.MergeTagCalls, MergeTagCall{UserUUID: userUUID, OldTagUUID: oldTagUUID, NewTagUUID: newTagUUID})

	if r.MergeTagErr != nil {
		return r.MergeTagErr
	}

	for index, tag := range r.Tags {
		if tag.UserUUID == userUUID && tag.UUID == oldTagUUID {
			r.Tags = slices.Delete(r.Tags, index, index+1)
			break
		}
	}

	return nil
}

func (r *FakeRepository) TagAdd(_ context.Context, tag Tag) error {
	r.Tags = append(r.Tags, tag)
	return nil
}

func (r *FakeRepository) TagNameIsRegistered(_ context.Context, userUUID, name string) (bool, error) {
	for _, tag := range r.Tags {
		if tag.UserUUID == userUUID && tag.Name == name {
			return true, nil
		}
	}

	return false, nil
}

func (r *FakeRepository) TagDelete(_ context.Context, userUUID, name string) error {
	for index, tag := range r.Tags {
		if tag.UserUUID == userUUID && tag.Name == name {
			r.Tags = slices.Delete(r.Tags, index, index+1)
			return nil
		}
	}

	return nil
}

func (r *FakeRepository) TagGetByName(_ context.Context, userUUID, name string) (Tag, error) {
	for _, tag := range r.Tags {
		if tag.UserUUID == userUUID && tag.Name == name {
			return tag, nil
		}
	}

	return Tag{}, ErrNotFound
}

func (r *FakeRepository) TagGetByUUID(_ context.Context, userUUID, tagUUID string) (Tag, error) {
	for _, tag := range r.Tags {
		if tag.UserUUID == userUUID && tag.UUID == tagUUID {
			return tag, nil
		}
	}

	return Tag{}, ErrNotFound
}

func (r *FakeRepository) TagGetCount(_ context.Context, userUUID string) (uint, error) {
	return uint(len(r.tagsFor(userUUID))), nil
}

func (r *FakeRepository) TagGetN(_ context.Context, userUUID string, n, offset uint) ([]Tag, error) {
	return paginateTags(r.tagsFor(userUUID), n, offset), nil
}

func (r *FakeRepository) TagSearchCount(_ context.Context, userUUID, searchTerms string) (uint, error) {
	return uint(len(r.tagsMatching(userUUID, searchTerms))), nil
}

func (r *FakeRepository) TagSearchN(_ context.Context, userUUID, searchTerms string, n, offset uint) ([]Tag, error) {
	return paginateTags(r.tagsMatching(userUUID, searchTerms), n, offset), nil
}

func (r *FakeRepository) tagsFor(userUUID string) []Tag {
	var tags []Tag

	for _, tag := range r.Tags {
		if tag.UserUUID == userUUID {
			tags = append(tags, tag)
		}
	}

	sort.Slice(tags, func(i, j int) bool { return tags[i].Name < tags[j].Name })

	return tags
}

func (r *FakeRepository) tagsMatching(userUUID, searchTerms string) []Tag {
	var tags []Tag

	for _, tag := range r.tagsFor(userUUID) {
		if strings.Contains(strings.ToLower(tag.Name), strings.ToLower(searchTerms)) {
			tags = append(tags, tag)
		}
	}

	return tags
}

func paginateTags(tags []Tag, n, offset uint) []Tag {
	if offset >= uint(len(tags)) {
		return []Tag{}
	}

	end := min(offset+n, uint(len(tags)))

	return tags[offset:end]
}

func (r *FakeRepository) TagRename(_ context.Context, userUUID, tagUUID, newName string) error {
	for index, tag := range r.Tags {
		if tag.UserUUID == userUUID && tag.UUID == tagUUID {
			tag.Name = newName
			r.Tags[index] = tag
			return nil
		}
	}

	return ErrNotFound
}
