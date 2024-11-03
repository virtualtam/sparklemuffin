// Copyright (c) VirtualTam
// SPDX-License-Identifier: MIT

package feed

import (
	"net/url"
	"slices"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/gosimple/slug"
)

var (
	allowedFeedURLSchemes = []string{"http", "https"}
)

// Feed represents a Web syndication feed (Atom or RSS).
type Feed struct {
	UUID string

	FeedURL string
	Title   string
	Slug    string

	ETag string

	CreatedAt time.Time
	UpdatedAt time.Time
	FetchedAt time.Time
}

// NewFeed initializes and returns a new Feed.
func NewFeed(feedURL string) (Feed, error) {
	now := time.Now().UTC()

	generatedUUID, err := uuid.NewRandom()
	if err != nil {
		return Feed{}, err
	}

	f := Feed{
		UUID:      generatedUUID.String(),
		FeedURL:   feedURL,
		CreatedAt: now,
		UpdatedAt: now,
	}
	f.normalizeURL()

	return f, nil
}

// Normalize sanitizes and normalizes all fields.
func (f *Feed) Normalize() {
	f.normalizeURL()
	f.normalizeTitle()
	f.slugify()
}

func (f *Feed) normalizeTitle() {
	f.Title = strings.TrimSpace(f.Title)
}

func (f *Feed) normalizeURL() {
	f.FeedURL = strings.TrimSpace(f.FeedURL)
}

func (f *Feed) slugify() {
	f.Slug = slug.Make(f.Title)
}

// ValidateForCreation ensures mandatory fields are set
// when creating a new Feed.
func (f *Feed) ValidateForCreation() error {
	fns := []func() error{
		f.requireURL,
		f.ensureURLIsValid,
		f.requireTitle,
		f.requireSlug,
	}

	for _, fn := range fns {
		if err := fn(); err != nil {
			return err
		}
	}

	return nil
}

// ValidateSlug ensures the slug is normalized and valid.
func (f *Feed) ValidateSlug() error {
	if !slug.IsSlug(f.Slug) {
		return ErrFeedSlugInvalid
	}

	return nil
}

// ValidateURL validates the URL is properly formed and uses a supported scheme.
func (f *Feed) ValidateURL() error {
	if err := f.requireURL(); err != nil {
		return err
	}

	if err := f.ensureURLIsValid(); err != nil {
		return err
	}

	return nil
}

func (f *Feed) requireURL() error {
	if f.FeedURL == "" {
		return ErrFeedURLInvalid
	}

	return nil
}

func (f *Feed) ensureURLIsValid() error {
	parsedURL, err := url.Parse(f.FeedURL)
	if err != nil {
		return ErrFeedURLInvalid
	}

	if parsedURL.Scheme == "" {
		return ErrFeedURLNoScheme
	}

	if !slices.Contains(allowedFeedURLSchemes, parsedURL.Scheme) {
		return ErrFeedURLUnsupportedScheme
	}

	if parsedURL.Host == "" {
		return ErrFeedURLNoHost
	}

	return nil
}

func (f *Feed) requireSlug() error {
	if f.Slug == "" {
		return ErrFeedSlugRequired
	}
	return nil
}

func (f *Feed) requireTitle() error {
	if f.Title == "" {
		return ErrFeedTitleRequired
	}
	return nil
}
