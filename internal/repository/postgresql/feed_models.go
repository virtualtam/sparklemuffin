// Copyright (c) VirtualTam
// SPDX-License-Identifier: MIT

package postgresql

import (
	"fmt"
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

	URL     string `db:"url"`
	Title   string `db:"title"`
	Summary string `db:"summary"`

	PublishedAt time.Time `db:"published_at"`
	UpdatedAt   time.Time `db:"updated_at"`
}

func (e *DBEntry) asEntry() feed.Entry {
	return feed.Entry{
		UID:         e.UID,
		FeedUUID:    e.FeedUUID,
		URL:         e.URL,
		Title:       e.Title,
		Summary:     e.Summary,
		PublishedAt: e.PublishedAt,
		UpdatedAt:   e.UpdatedAt,
	}
}

func feedEntryToFullTextSearchString(e feed.Entry) string {
	return fullTextSearchReplacer.Replace(e.Title)
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
	FeedDescription  string `db:"description"`
}

func (st *DBSubscriptionTitle) asSubscriptionTitle() feedquerying.SubscriptionTitle {
	return feedquerying.SubscriptionTitle{
		SubscriptionUUID: st.SubscriptionUUID,
		FeedTitle:        st.FeedTitle,
		FeedDescription:  st.FeedDescription,
	}
}
