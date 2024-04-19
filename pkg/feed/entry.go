// Copyright (c) VirtualTam
// SPDX-License-Identifier: MIT

package feed

import (
	"net/url"
	"slices"
	"strings"
	"time"

	"github.com/segmentio/ksuid"
)

// Entry represents an entry of a syndication feed (Atom or RSS).
type Entry struct {
	UID      string
	FeedUUID string

	URL   string
	Title string

	PublishedAt time.Time
	UpdatedAt   time.Time
}

// NewEntry creates and initializes a new Entry
func NewEntry(feedUUID string, URL string, Title string, PublishedAt time.Time, UpdatedAt time.Time) Entry {
	uid := ksuid.New().String()

	entry := Entry{
		UID:         uid,
		FeedUUID:    feedUUID,
		URL:         URL,
		Title:       Title,
		PublishedAt: PublishedAt,
		UpdatedAt:   UpdatedAt,
	}
	entry.Normalize()

	return entry
}

// Normalize sanitizes and normalizes all fields.
func (e *Entry) Normalize() {
	e.normalizeURL()
	e.normalizeTitle()
}

// ValidateForAddition ensures mandatory fields are properly set when adding an
// new Entry.
func (e *Entry) ValidateForAddition() error {
	fns := []func() error{
		e.requireURL,
		e.ensureURLIsValid,
		e.requireTitle,
		e.validateUID,
	}

	for _, fn := range fns {
		if err := fn(); err != nil {
			return err
		}
	}

	return nil
}

func (e *Entry) normalizeTitle() {
	e.Title = strings.TrimSpace(e.Title)
}

func (e *Entry) normalizeURL() {
	e.URL = strings.TrimSpace(e.URL)
}

func (e *Entry) ensureURLIsValid() error {
	parsedURL, err := url.Parse(e.URL)
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

func (e *Entry) requireTitle() error {
	if e.Title == "" {
		return ErrFeedTitleRequired
	}
	return nil
}

func (e *Entry) requireURL() error {
	if e.URL == "" {
		return ErrFeedURLRequired
	}
	return nil
}

func (e *Entry) validateUID() error {
	_, err := ksuid.Parse(e.UID)
	if err != nil {
		return ErrFeedUIDInvalid
	}

	return nil
}
