// Copyright (c) VirtualTam
// SPDX-License-Identifier: MIT

package postgresql

import (
	"fmt"
	"strings"
	"time"

	"github.com/virtualtam/sparklemuffin/pkg/feed"
	feedquerying "github.com/virtualtam/sparklemuffin/pkg/feed/querying"
	feedsynchronizing "github.com/virtualtam/sparklemuffin/pkg/feed/synchronizing"
)

type DBCategory struct {
	UUID     string `db:"uuid"`
	UserUUID string `db:"user_uuid"`

	Name string `db:"name"`
	Slug string `db:"slug"`

	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
}

func (c *DBCategory) asCategory() feed.Category {
	return feed.Category{
		UUID:      c.UUID,
		UserUUID:  c.UserUUID,
		Name:      c.Name,
		Slug:      c.Slug,
		CreatedAt: c.CreatedAt,
		UpdatedAt: c.UpdatedAt,
	}
}

type DBFeed struct {
	UUID string `db:"uuid"`

	FeedURL     string `db:"feed_url"`
	Title       string `db:"title"`
	Description string `db:"description"`
	Slug        string `db:"slug"`

	ETag         string    `db:"etag"`
	LastModified time.Time `db:"last_modified"`

	// xxHash64 returns a 64-bit unsigned integer hash value,
	// whereas it is stored as a 64-bit signed integer in the database (BIGINT).
	Hash int64 `db:"hash_xxhash64"`

	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
	FetchedAt time.Time `db:"fetched_at"`
}

func (f *DBFeed) asFeed() feed.Feed {
	return feed.Feed{
		UUID:         f.UUID,
		FeedURL:      f.FeedURL,
		Title:        f.Title,
		Description:  f.Description,
		Slug:         f.Slug,
		ETag:         f.ETag,
		Hash:         uint64(f.Hash), // int64 (BIGINT) -> uint64
		LastModified: f.LastModified,
		CreatedAt:    f.CreatedAt,
		UpdatedAt:    f.UpdatedAt,
		FetchedAt:    f.FetchedAt,
	}
}

func feedToFullTextSearchString(f feed.Feed) string {
	return fmt.Sprintf(
		"%s %s",
		fullTextSearchReplacer.Replace(f.Title),
		fullTextSearchReplacer.Replace(f.Description),
	)
}

func feedMetadataToFullTextSearchString(feedMetadata feedsynchronizing.FeedMetadata) string {
	return fmt.Sprintf(
		"%s %s",
		fullTextSearchReplacer.Replace(feedMetadata.Title),
		fullTextSearchReplacer.Replace(feedMetadata.Description),
	)
}

type DBEntry struct {
	UID      string `db:"uid"`
	FeedUUID string `db:"feed_uuid"`

	URL           string   `db:"url"`
	Title         string   `db:"title"`
	Summary       string   `db:"summary"`
	TextRankTerms []string `db:"textrank_terms"`

	PublishedAt time.Time `db:"published_at"`
	UpdatedAt   time.Time `db:"updated_at"`
}

func (e *DBEntry) asEntry() feed.Entry {
	return feed.Entry{
		UID:           e.UID,
		FeedUUID:      e.FeedUUID,
		URL:           e.URL,
		Title:         e.Title,
		Summary:       e.Summary,
		TextRankTerms: e.TextRankTerms,
		PublishedAt:   e.PublishedAt,
		UpdatedAt:     e.UpdatedAt,
	}
}

func feedEntryToFullTextSearchString(e feed.Entry) string {
	return fmt.Sprintf(
		"%s %s",
		fullTextSearchReplacer.Replace(e.Title),
		fullTextSearchReplacer.Replace(strings.Join(e.TextRankTerms, " ")),
	)
}

type DBEntryMetadata struct {
	UserUUID string `db:"user_uuid"`
	EntryUID string `db:"entry_uid"`
	Read     bool   `db:"read"`
}

func (em *DBEntryMetadata) asEntryMetadata() feed.EntryMetadata {
	return feed.EntryMetadata{
		UserUUID: em.UserUUID,
		EntryUID: em.EntryUID,
		Read:     em.Read,
	}
}

type DBQueryingSubscribedFeedEntry struct {
	DBEntry

	FeedTitle string `db:"feed_title"`
	Read      bool   `db:"read"`
}

func (qe *DBQueryingSubscribedFeedEntry) asQueryingSubscribedFeedEntry() feedquerying.SubscribedFeedEntry {
	return feedquerying.SubscribedFeedEntry{
		Entry:     qe.asEntry(),
		FeedTitle: qe.FeedTitle,
		Read:      qe.Read,
	}
}

type DBSubscribedFeed struct {
	UUID    string `db:"uuid"`
	FeedURL string `db:"feed_url"`
	Title   string `db:"title"`
	Slug    string `db:"slug"`

	Alias  string `db:"alias"`
	Unread uint   `db:"unread"`

	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
	FetchedAt time.Time `db:"fetched_at"`
}

func (f *DBSubscribedFeed) asSubscribedFeed() feedquerying.SubscribedFeed {
	return feedquerying.SubscribedFeed{
		Feed: feed.Feed{
			UUID:      f.UUID,
			FeedURL:   f.FeedURL,
			Title:     f.Title,
			Slug:      f.Slug,
			CreatedAt: f.CreatedAt,
			UpdatedAt: f.UpdatedAt,
			FetchedAt: f.FetchedAt,
		},
		Alias:  f.Alias,
		Unread: f.Unread,
	}
}

type DBSubscription struct {
	UUID         string `db:"uuid"`
	CategoryUUID string `db:"category_uuid"`
	FeedUUID     string `db:"feed_uuid"`
	UserUUID     string `db:"user_uuid"`

	Alias string `db:"alias"`

	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
}

func (s *DBSubscription) asSubscription() feed.Subscription {
	return feed.Subscription{
		UUID:         s.UUID,
		CategoryUUID: s.CategoryUUID,
		FeedUUID:     s.FeedUUID,
		UserUUID:     s.UserUUID,
		Alias:        s.Alias,
		CreatedAt:    s.CreatedAt,
		UpdatedAt:    s.UpdatedAt,
	}
}

type DBQueryingSubscription struct {
	SubscriptionUUID  string `db:"uuid"`
	SubscriptionAlias string `db:"alias"`
	CategoryUUID      string `db:"category_uuid"`

	FeedTitle       string `db:"title"`
	FeedDescription string `db:"description"`
}

func (s *DBQueryingSubscription) asQueryingSubscription() feedquerying.Subscription {
	return feedquerying.Subscription{
		UUID:            s.SubscriptionUUID,
		Alias:           s.SubscriptionAlias,
		CategoryUUID:    s.CategoryUUID,
		FeedTitle:       s.FeedTitle,
		FeedDescription: s.FeedDescription,
	}
}
