// Copyright (c) VirtualTam
// SPDX-License-Identifier: MIT

package feed

import (
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/gosimple/slug"
)

// Categories allow users to group feed subscriptions.
type Category struct {
	UUID     string
	UserUUID string

	Name string
	Slug string

	CreatedAt time.Time
	UpdatedAt time.Time
}

// NewCategory initializes and returns a new feed category.
func NewCategory(userUUID string, name string) (Category, error) {
	now := time.Now().UTC()

	generatedUUID, err := uuid.NewRandom()
	if err != nil {
		return Category{}, err
	}

	category := Category{
		UUID:      generatedUUID.String(),
		UserUUID:  userUUID,
		Name:      name,
		CreatedAt: now,
		UpdatedAt: now,
	}

	category.Normalize()

	return category, nil
}

// Normalize sanitizes and normalizes all fields.
func (c *Category) Normalize() {
	c.normalizeName()
	c.slugify()
}

// ValidateForAddition ensures mandatory fields are properly set when adding a Category.
func (c *Category) ValidateForAddition(v ValidationRepository) error {
	fns := []func() error{
		c.requireName,
		c.requireSlug,
		c.ensureNameAndSlugAreNotRegistered(v),
	}

	for _, fn := range fns {
		if err := fn(); err != nil {
			return err
		}
	}

	return nil
}

// ValidateForDeletion ensures mandatory fields are properly set when deleting a Category.
func (c *Category) ValidateForDeletion(v ValidationRepository) error {
	fns := []func() error{
		c.requireUUID,
		c.validateUUID,
		c.requireUserUUID,
	}

	for _, fn := range fns {
		if err := fn(); err != nil {
			return err
		}
	}

	return nil
}

// ValidateForUpdate ensures mandatory fields are properly set when editing a Category.
func (c *Category) ValidateForUpdate(v ValidationRepository) error {
	fns := []func() error{
		c.requireName,
		c.requireSlug,
		c.ensureNameAndSlugAreNotRegisteredToAnotherCategory(v),
	}

	for _, fn := range fns {
		if err := fn(); err != nil {
			return err
		}
	}

	return nil
}

func (c *Category) normalizeName() {
	c.Name = strings.TrimSpace(c.Name)
}

func (c *Category) slugify() {
	c.Slug = slug.Make(strings.ToLower(c.Name))
}

func (c *Category) requireName() error {
	if c.Name == "" {
		return ErrCategoryNameRequired
	}
	return nil
}

func (c *Category) requireSlug() error {
	if c.Slug == "" {
		return ErrCategorySlugRequired
	}
	return nil
}

func (c *Category) requireUserUUID() error {
	if c.UserUUID == "" {
		return ErrCategoryUserUUIDRequired
	}
	return nil
}

func (c *Category) requireUUID() error {
	if c.UUID == "" {
		return ErrCategoryUUIDRequired
	}
	return nil
}

func (c *Category) ensureNameAndSlugAreNotRegistered(v ValidationRepository) func() error {
	return func() error {
		registered, err := v.FeedCategoryNameAndSlugAreRegistered(c.UserUUID, c.Name, c.Slug)
		if err != nil {
			return err
		}

		if registered {
			return ErrCategoryAlreadyRegistered
		}

		return nil
	}
}

func (c *Category) ensureNameAndSlugAreNotRegisteredToAnotherCategory(v ValidationRepository) func() error {
	return func() error {
		registered, err := v.FeedCategoryNameAndSlugAreRegisteredToAnotherCategory(c.UserUUID, c.UUID, c.Name, c.Slug)
		if err != nil {
			return err
		}

		if registered {
			return ErrCategoryAlreadyRegistered
		}

		return nil
	}
}

func (c *Category) validateSlug() error {
	if !slug.IsSlug(c.Slug) {
		return ErrCategorySlugInvalid
	}

	return nil
}

func (c *Category) validateUUID() error {
	if err := uuid.Validate(c.UUID); err != nil {
		return ErrCategoryUUIDInvalid
	}

	return nil
}
