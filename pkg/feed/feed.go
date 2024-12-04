// Copyright (c) VirtualTam
// SPDX-License-Identifier: MIT

package feed

import (
	"net/url"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/gosimple/slug"
	"github.com/virtualtam/sparklemuffin/internal/test/assert"
)

var (
	allowedFeedURLSchemes = []string{"http", "https"}
)

// Feed represents a Web syndication feed (Atom or RSS).
type Feed struct {
	// UUID is a unique identifier for the feed.
	UUID string

	FeedURL     string
	Title       string
	Description string
	Slug        string

	ETag         string
	LastModified time.Time

	Hash uint64

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
	f.normalizeDescription()
	f.slugify()
}

func (f *Feed) normalizeTitle() {
	f.Title = strings.TrimSpace(f.Title)
}

func (f *Feed) normalizeDescription() {
	f.Description = strings.TrimSpace(f.Description)
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
		f.requireHash,
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

func (f *Feed) requireHash() error {
	if f.Hash == 0 {
		return ErrFeedHashRequired
	}
	return nil
}

func AssertFeedEquals(t *testing.T, got, want Feed) {
	t.Helper()

	if got.Slug != want.Slug {
		t.Errorf("want Slug %q, got %q", want.Slug, got.Slug)
	}
	if got.Title != want.Title {
		t.Errorf("want Title %q, got %q", want.Title, got.Title)
	}
	if got.Description != want.Description {
		t.Errorf("want Description %q, got %q", want.Description, got.Description)
	}
	if got.FeedURL != want.FeedURL {
		t.Errorf("want FeedURL %q, got %q", want.FeedURL, got.FeedURL)
	}

	if got.ETag != want.ETag {
		t.Errorf("want ETag %q, got %q", want.ETag, got.ETag)
	}

	if got.Hash != want.Hash {
		t.Errorf("want Hash %d, got %d", want.Hash, got.Hash)
	}

	assert.TimeAlmostEquals(t, "LastModified", got.LastModified, want.LastModified, assert.TimeComparisonDelta)
	assert.TimeAlmostEquals(t, "CreatedAt", got.CreatedAt, want.CreatedAt, assert.TimeComparisonDelta)
	assert.TimeAlmostEquals(t, "UpdatedAt", got.UpdatedAt, want.UpdatedAt, assert.TimeComparisonDelta)
	assert.TimeAlmostEquals(t, "FetchedAt", got.FetchedAt, want.FetchedAt, assert.TimeComparisonDelta)
}

func AssertFeedsEqual(t *testing.T, gotFeeds, wantFeeds []Feed) {
	t.Helper()

	if len(gotFeeds) != len(wantFeeds) {
		t.Fatalf("want %d feeds, got %d", len(wantFeeds), len(gotFeeds))
	}

	for i, wantFeed := range wantFeeds {
		gotFeed := gotFeeds[i]
		AssertFeedEquals(t, gotFeed, wantFeed)
	}
}
