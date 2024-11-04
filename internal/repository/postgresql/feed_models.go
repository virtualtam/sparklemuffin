// Copyright (c) VirtualTam
// SPDX-License-Identifier: MIT

package postgresql

import (
	"time"

	"github.com/virtualtam/sparklemuffin/pkg/feed"
	feedquerying "github.com/virtualtam/sparklemuffin/pkg/feed/querying"
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

	FeedURL string `db:"feed_url"`
	Title   string `db:"title"`
	Slug    string `db:"slug"`

	ETag         string    `db:"etag"`
	LastModified time.Time `db:"last_modified"`

	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
	FetchedAt time.Time `db:"fetched_at"`
}

func (f *DBFeed) asFeed() feed.Feed {
	return feed.Feed{
		UUID:         f.UUID,
		FeedURL:      f.FeedURL,
		Title:        f.Title,
		Slug:         f.Slug,
		ETag:         f.ETag,
		LastModified: f.LastModified,
		CreatedAt:    f.CreatedAt,
		UpdatedAt:    f.UpdatedAt,
		FetchedAt:    f.FetchedAt,
	}
}

type DBEntry struct {
	UID      string `db:"uid"`
	FeedUUID string `db:"feed_uuid"`

	URL   string `db:"url"`
	Title string `db:"title"`

	PublishedAt time.Time `db:"published_at"`
	UpdatedAt   time.Time `db:"updated_at"`
}

func (e *DBEntry) asEntry() feed.Entry {
	return feed.Entry{
		UID:         e.UID,
		FeedUUID:    e.FeedUUID,
		URL:         e.URL,
		Title:       e.Title,
		PublishedAt: e.PublishedAt,
		UpdatedAt:   e.UpdatedAt,
	}
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

type DBQueryingEntry struct {
	DBEntry

	Read bool `db:"read"`
}

func (qe *DBQueryingEntry) asQueryingEntry() feedquerying.SubscribedFeedEntry {
	return feedquerying.SubscribedFeedEntry{
		Entry: qe.asEntry(),
		Read:  qe.Read,
	}
}

type DBSubscribedFeed struct {
	UUID    string `db:"uuid"`
	FeedURL string `db:"feed_url"`
	Title   string `db:"title"`
	Slug    string `db:"slug"`

	Unread uint `db:"unread"`

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
		Unread: f.Unread,
	}
}

type DBSubscription struct {
	UUID         string `db:"uuid"`
	CategoryUUID string `db:"category_uuid"`
	FeedUUID     string `db:"feed_uuid"`
	UserUUID     string `db:"user_uuid"`

	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
}

func (s *DBSubscription) asSubscription() feed.Subscription {
	return feed.Subscription{
		UUID:         s.UUID,
		CategoryUUID: s.CategoryUUID,
		FeedUUID:     s.FeedUUID,
		UserUUID:     s.UserUUID,
		CreatedAt:    s.CreatedAt,
		UpdatedAt:    s.UpdatedAt,
	}
}

type DBSubscriptionTitle struct {
	SubscriptionUUID string `db:"uuid"`
	FeedTitle        string `db:"title"`
}

func (st *DBSubscriptionTitle) asSubscriptionTitle() feedquerying.SubscriptionTitle {
	return feedquerying.SubscriptionTitle{
		SubscriptionUUID: st.SubscriptionUUID,
		FeedTitle:        st.FeedTitle,
	}
}
