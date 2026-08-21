// Copyright VirtualTam 2022, 2026
// SPDX-License-Identifier: MIT

package taxonomy

import (
	"context"
	"slices"
)

var _ Repository = &FakeRepository{}

type FakeRepository struct {
	Tags []Tag
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
