// Copyright (c) VirtualTam
// SPDX-License-Identifier: MIT

package feed

import (
	"fmt"
	"net/url"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/mmcdole/gofeed"
	"github.com/segmentio/ksuid"

	"github.com/virtualtam/sparklemuffin/internal/test/assert"
	"github.com/virtualtam/sparklemuffin/internal/textkit"
)

const (
	EntryTextRankMaxTerms = 10

	// Entry PublishedAt or UpdatedAt may be accepted as late as (time.Now() + entryTimeFutureBound).
	entryTimeFutureBound = 2 * 24 * time.Hour
)

// Entry represents an entry of a syndication feed (Atom or RSS).
type Entry struct {
	UID      string
	FeedUUID string

	URL   string
	Title string

	description   string
	content       string
	Summary       string
	TextRankTerms []string

	PublishedAt time.Time
	UpdatedAt   time.Time
}

// NewEntryFromItem creates and initializes a new Entry from a gofeed.Item.
func NewEntryFromItem(feedUUID string, now time.Time, item *gofeed.Item) Entry {
	uid := ksuid.New().String()

	publishedAt := now
	if item.PublishedParsed != nil {
		publishedAt = *item.PublishedParsed
	}

	updatedAt := publishedAt
	if item.UpdatedParsed != nil {
		updatedAt = *item.UpdatedParsed
	}

	entry := Entry{
		UID:         uid,
		FeedUUID:    feedUUID,
		URL:         item.Link,
		Title:       item.Title,
		description: item.Description,
		content:     item.Content,
		PublishedAt: publishedAt,
		UpdatedAt:   updatedAt,
	}
	entry.Normalize()

	return entry
}

// Normalize sanitizes and normalizes all fields.
func (e *Entry) Normalize() {
	e.normalizeURL()
	e.normalizeTitle()
	e.normalizeDescription()
	e.normalizeContent()
	e.summarize()
	e.normalizePublishedAt()
	e.normalizeUpdatedAt()
}

// ExtractTextRankTerms uses TextRank to extract the top ranking phrases from the Entry content or description.
//
// Given the TextRank implementation we use, ranked phrases (pairs of terms) seem semantically more relevant
// than single ranked words.
func (e *Entry) ExtractTextRankTerms(ranker *textkit.TextRanker, topN int) {
	if len(e.content) > 0 {
		e.TextRankTerms = ranker.RankTopNPhrases(e.content, topN)
	} else if len(e.description) > 0 {
		e.TextRankTerms = ranker.RankTopNPhrases(e.description, topN)
	}
}

// ValidateForAddition ensures mandatory fields are properly set when adding a new Entry.
func (e *Entry) ValidateForAddition(now time.Time) error {
	entryTimeMustBeBefore := now.Add(entryTimeFutureBound)

	fns := []func() error{
		e.requireURL,
		e.ensureURLIsValid,
		e.requireTitle,
		e.validateUID,
		e.requirePublishedAt,
		e.ensurePublishedAtIsBefore(entryTimeMustBeBefore),
		e.requireUpdatedAt,
		e.ensureUpdatedAtIsBefore(entryTimeMustBeBefore),
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

func (e *Entry) normalizeDescription() {
	e.description = textkit.NormalizeHTMLToText(e.description)
}

func (e *Entry) normalizeContent() {
	e.content = textkit.NormalizeHTMLToText(e.content)
}

func (e *Entry) normalizePublishedAt() {
	if e.PublishedAt.IsZero() {
		e.PublishedAt = time.Now().UTC()
	}
}

func (e *Entry) normalizeUpdatedAt() {
	if e.UpdatedAt.IsZero() {
		e.UpdatedAt = e.PublishedAt
	}
}

const (
	entrySummaryKeepIfUnder   = 200 // Length to consider text "short enough" as is
	entrySummaryTruncateAfter = 400 // Maximum length for multi-paragraph summary
)

func (e *Entry) summarize() {
	if e.content != "" {
		e.Summary = textkit.Summarize(e.content, entrySummaryKeepIfUnder, entrySummaryTruncateAfter)
		return
	}

	if e.description != "" {
		e.Summary = textkit.Summarize(e.description, entrySummaryKeepIfUnder, entrySummaryTruncateAfter)
	}
}

func (e *Entry) ensureURLIsValid() error {
	parsedURL, err := url.Parse(e.URL)
	if err != nil {
		return ErrEntryURLInvalid
	}

	if parsedURL.Scheme == "" {
		return ErrEntryURLNoScheme
	}

	if !slices.Contains(allowedFeedURLSchemes, parsedURL.Scheme) {
		return ErrEntryURLUnsupportedScheme
	}

	if parsedURL.Host == "" {
		return ErrEntryURLNoHost
	}

	return nil
}

func (e *Entry) requireTitle() error {
	if e.Title == "" {
		return ErrEntryTitleRequired
	}
	return nil
}

func (e *Entry) requireURL() error {
	if e.URL == "" {
		return ErrEntryURLRequired
	}
	return nil
}

func (e *Entry) validateUID() error {
	_, err := ksuid.Parse(e.UID)
	if err != nil {
		return ErrEntryUIDInvalid
	}

	return nil
}

func (e *Entry) requirePublishedAt() error {
	if e.PublishedAt.IsZero() {
		return ErrEntryPublishedAtIsZero
	}
	return nil
}

func (e *Entry) ensurePublishedAtIsBefore(before time.Time) func() error {
	return func() error {
		if e.PublishedAt.After(before) {
			return ErrEntryPublishedAtInTheFuture
		}
		return nil
	}
}

func (e *Entry) requireUpdatedAt() error {
	if e.UpdatedAt.IsZero() {
		return ErrEntryUpdatedAtIsZero
	}
	return nil
}

func (e *Entry) ensureUpdatedAtIsBefore(before time.Time) func() error {
	return func() error {
		if e.UpdatedAt.After(before) {
			return ErrEntryUpdatedAtInTheFuture
		}
		return nil
	}
}

func AssertEntriesEqual(t *testing.T, gotEntries []Entry, wantEntries []Entry) {
	t.Helper()

	if len(gotEntries) != len(wantEntries) {
		t.Fatalf("want %d entries, got %d", len(wantEntries), len(gotEntries))
	}

	for i, wantEntry := range wantEntries {
		AssertEntryEquals(t, i, gotEntries[i], wantEntry)
	}
}

func AssertEntryEquals(t *testing.T, index int, gotEntry Entry, wantEntry Entry) {
	t.Helper()

	if gotEntry.FeedUUID != wantEntry.FeedUUID {
		t.Errorf("want Entry %d FeedUUID %q, got %q", index, wantEntry.FeedUUID, gotEntry.FeedUUID)
	}

	if gotEntry.URL != wantEntry.URL {
		t.Errorf("want Entry %d URL %q, got %q", index, wantEntry.URL, gotEntry.URL)
	}
	if gotEntry.Title != wantEntry.Title {
		t.Errorf("want Entry %d Title %q, got %q", index, wantEntry.Title, gotEntry.Title)
	}
	if gotEntry.Summary != wantEntry.Summary {
		t.Errorf("want Entry %d Summary %q, got %q", index, wantEntry.Summary, gotEntry.Summary)
	}

	if len(gotEntry.TextRankTerms) != len(wantEntry.TextRankTerms) {
		t.Errorf("want Entry %d TextRankTerms %#v, got %#v", index, wantEntry.TextRankTerms, gotEntry.TextRankTerms)
		t.Fatalf("want Entry %d TextRankTerms %d, got %d", index, len(wantEntry.TextRankTerms), len(gotEntry.TextRankTerms))
	}

	for j, wantTerm := range wantEntry.TextRankTerms {
		if gotEntry.TextRankTerms[j] != wantTerm {
			t.Errorf("want Entry %d TextRankTerms[%d] %q, got %q", index, j, wantTerm, gotEntry.TextRankTerms[j])
		}
	}

	assert.TimeAlmostEquals(t, fmt.Sprintf("Entry %d PublishedAt", index), gotEntry.PublishedAt, wantEntry.PublishedAt, assert.TimeComparisonDelta)
	assert.TimeAlmostEquals(t, fmt.Sprintf("Entry %d UpdatedAt", index), gotEntry.UpdatedAt, wantEntry.UpdatedAt, assert.TimeComparisonDelta)
}
