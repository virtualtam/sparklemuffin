// Copyright (c) VirtualTam
// SPDX-License-Identifier: MIT

package feed

import (
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/feeds"
	"github.com/jaswdr/faker"
	"github.com/mmcdole/gofeed"
	"github.com/segmentio/ksuid"
	"github.com/virtualtam/sparklemuffin/pkg/feed/fetching"
)

type testRoundTripper struct{}

func (testRoundTripper) RoundTrip(*http.Request) (*http.Response, error) {
	now := time.Now().UTC()
	yesterday := now.Add(-24 * time.Hour)

	feed := &feeds.Feed{
		Title:   "Local Test",
		Updated: now,
		Items: []*feeds.Item{
			{
				Id:    "http://test.local/first-post",
				Title: "First post!",
				Link: &feeds.Link{
					Href: "http://test.local/first-post",
				},
				Created: now,
				Updated: now,
			},
			{
				Id:    "http://test.local/hello-world",
				Title: "Hello World",
				Link: &feeds.Link{
					Href: "http://test.local/hello-world",
				},
				Created: yesterday,
				Updated: yesterday,
			},
		},
	}

	feedStr, err := feed.ToAtom()
	if err != nil {
		panic(err)
	}

	resp := &http.Response{
		StatusCode: 200,
		Header: map[string][]string{
			"Content-Disposition": {"attachment; filename=test.atom"},
			"Content-Type":        {"application/atom+xml"},
		},
		Body: io.NopCloser(strings.NewReader(feedStr)),
	}

	return resp, nil
}

func TestServiceAddCategory(t *testing.T) {
	userUUID := "179206c8-2965-47a7-ba04-bf0a6a0b8d11"
	now := time.Now().UTC()

	cases := []struct {
		tname                string
		repositoryCategories []Category
		name                 string
		want                 Category
		wantErr              error
	}{
		// nominal cases
		{
			tname: "new category",
			name:  "Linux Distributions",
			want: Category{
				UUID:      "-",
				UserUUID:  userUUID,
				Name:      "Linux Distributions",
				Slug:      "linux-distributions",
				CreatedAt: now,
				UpdatedAt: now,
			},
		},
		{
			tname: "new category with accented characters and punctuation",
			name:  "Choses à faire, peut-être aujourd'hui?",
			want: Category{
				UUID:      "-",
				UserUUID:  userUUID,
				Name:      "Choses à faire, peut-être aujourd'hui?",
				Slug:      "choses-a-faire-peut-etre-aujourdhui",
				CreatedAt: now,
				UpdatedAt: now,
			},
		},

		// error cases
		{
			tname:   "empty name",
			wantErr: ErrCategoryNameRequired,
		},
		{
			tname:   "empty name (whitespace)",
			name:    "     ",
			wantErr: ErrCategoryNameRequired,
		},
		{
			tname:   "empty slug (punctuation)",
			name:    "'?",
			wantErr: ErrCategorySlugRequired,
		},
		{
			tname: "existing category",
			repositoryCategories: []Category{
				{
					UserUUID: userUUID,
					Name:     "Duplicate",
					Slug:     "duplicate",
				},
			},
			name:    "Duplicate",
			wantErr: ErrCategoryAlreadyRegistered,
		},
		{
			tname: "existing category (case-insensitive)",
			repositoryCategories: []Category{
				{
					UserUUID: userUUID,
					Name:     "Duplicate",
					Slug:     "duplicate",
				},
			},
			name:    "DupliCate",
			wantErr: ErrCategoryAlreadyRegistered,
		},
	}

	for _, tc := range cases {
		t.Run(tc.tname, func(t *testing.T) {
			r := &fakeRepository{
				Categories: tc.repositoryCategories,
			}
			s := NewService(r, nil)

			got, err := s.AddCategory(userUUID, tc.name)

			if tc.wantErr != nil {
				if errors.Is(err, tc.wantErr) {
					return
				}
				if err == nil {
					t.Fatalf("want error %q, got nil", tc.wantErr)
				}
				t.Fatalf("want error %q, got %q", tc.wantErr, err)
			}

			if err != nil {
				t.Fatalf("want no error, got %q", err)
			}

			assertCategoryEquals(t, got, tc.want)
		})
	}
}

func TestServiceCategoryBySlug(t *testing.T) {
	userUUID := "8c9a910a-fcda-4eef-8394-8d580a969643"
	now := time.Now().UTC()

	cases := []struct {
		tname                string
		repositoryCategories []Category
		slug                 string
		want                 Category
		wantErr              error
	}{
		// nominal cases
		{
			tname:   "not found",
			slug:    "nonexistent",
			wantErr: ErrCategoryNotFound,
		},
		{
			tname: "found",
			repositoryCategories: []Category{
				{
					UUID:      "d3033032-23c0-4f78-9b7d-f4135477b5c3",
					UserUUID:  userUUID,
					Name:      "Existing Category",
					Slug:      "existingcat",
					CreatedAt: now,
					UpdatedAt: now,
				},
			},
			slug: "existingcat",
			want: Category{
				UUID:      "d3033032-23c0-4f78-9b7d-f4135477b5c3",
				UserUUID:  userUUID,
				Name:      "Existing Category",
				Slug:      "existingcat",
				CreatedAt: now,
				UpdatedAt: now,
			},
		},

		// error cases
		{
			tname:   "empty slug",
			wantErr: ErrCategorySlugInvalid,
		},
		{
			tname:   "empty slug (whitespace)",
			slug:    "    ",
			wantErr: ErrCategorySlugInvalid,
		},
		{
			tname:   "invalid slug (characters)",
			slug:    "ABC",
			wantErr: ErrCategorySlugInvalid,
		},
		{
			tname:   "invalid slug (punctuation)",
			slug:    "?.+",
			wantErr: ErrCategorySlugInvalid,
		},
	}

	for _, tc := range cases {
		t.Run(tc.tname, func(t *testing.T) {
			r := &fakeRepository{
				Categories: tc.repositoryCategories,
			}
			s := NewService(r, nil)

			got, err := s.CategoryBySlug(userUUID, tc.slug)

			if tc.wantErr != nil {
				if errors.Is(err, tc.wantErr) {
					return
				}
				if err == nil {
					t.Fatalf("want error %q, got nil", tc.wantErr)
				}
				t.Fatalf("want error %q, got %q", tc.wantErr, err)
			}

			if err != nil {
				t.Fatalf("want no error, got %q", err)
			}

			assertCategoryEquals(t, got, tc.want)
		})
	}
}

func TestServiceCategoryByUUID(t *testing.T) {
	userUUID := "8c9a910a-fcda-4eef-8394-8d580a969643"
	now := time.Now().UTC()

	cases := []struct {
		tname                string
		repositoryCategories []Category
		categoryUUID         string
		want                 Category
		wantErr              error
	}{
		// nominal cases
		{
			tname:        "not found",
			categoryUUID: "76a87b94-2e60-457e-a9e0-9deaebd761aa",
			wantErr:      ErrCategoryNotFound,
		},
		{
			tname: "found",
			repositoryCategories: []Category{
				{
					UUID:      "d3033032-23c0-4f78-9b7d-f4135477b5c3",
					UserUUID:  userUUID,
					Name:      "Existing Category",
					Slug:      "existingcat",
					CreatedAt: now,
					UpdatedAt: now,
				},
			},
			categoryUUID: "d3033032-23c0-4f78-9b7d-f4135477b5c3",
			want: Category{
				UUID:      "d3033032-23c0-4f78-9b7d-f4135477b5c3",
				UserUUID:  userUUID,
				Name:      "Existing Category",
				Slug:      "existingcat",
				CreatedAt: now,
				UpdatedAt: now,
			},
		},

		// error cases
		{
			tname:   "empty UUID",
			wantErr: ErrCategoryUUIDInvalid,
		},
		{
			tname:        "empty UUID (whitespace)",
			categoryUUID: "    ",
			wantErr:      ErrCategoryUUIDInvalid,
		},
		{
			tname:        "invalid UUID",
			categoryUUID: "A-BC",
			wantErr:      ErrCategoryUUIDInvalid,
		},
	}

	for _, tc := range cases {
		t.Run(tc.tname, func(t *testing.T) {
			r := &fakeRepository{
				Categories: tc.repositoryCategories,
			}
			s := NewService(r, nil)

			got, err := s.CategoryByUUID(userUUID, tc.categoryUUID)

			if tc.wantErr != nil {
				if errors.Is(err, tc.wantErr) {
					return
				}
				if err == nil {
					t.Fatalf("want error %q, got nil", tc.wantErr)
				}
				t.Fatalf("want error %q, got %q", tc.wantErr, err)
			}

			if err != nil {
				t.Fatalf("want no error, got %q", err)
			}

			assertCategoryEquals(t, got, tc.want)
		})
	}
}

func TestServiceDeleteCategory(t *testing.T) {
	fake := faker.New()

	userUUID := fake.UUID().V4()

	t.Run("empty category (no subscriptions)", func(t *testing.T) {
		emptyCategory := Category{
			UserUUID: userUUID,
			UUID:     fake.UUID().V4(),
			Name:     fake.Lorem().Text(10),
			Slug:     fake.Internet().Slug(),
		}

		r := &fakeRepository{
			Categories: []Category{emptyCategory},
		}
		s := NewService(r, nil)

		if err := s.DeleteCategory(userUUID, emptyCategory.UUID); err != nil {
			t.Fatalf("want no error, got %q", err)
		}

		if len(r.Categories) != 0 {
			t.Fatalf("want 0 Categories, got %d", len(r.Categories))
		}
	})

	t.Run("category with subscriptions", func(t *testing.T) {
		category := Category{
			UserUUID: userUUID,
			UUID:     fake.UUID().V4(),
			Name:     fake.Lorem().Text(10),
			Slug:     fake.Internet().Slug(),
		}
		categories := []Category{category}

		feeds := []Feed{}
		entries := []Entry{}
		subscriptions := []Subscription{}

		for i := 0; i < 5; i++ {
			feed := Feed{
				UUID:    fake.UUID().V4(),
				FeedURL: fake.Internet().URL(),
				Title:   fake.Lorem().Text(10),
				Slug:    fake.Internet().Slug(),
			}
			feeds = append(feeds, feed)

			subscription := Subscription{
				UUID:         fake.UUID().V4(),
				CategoryUUID: category.UUID,
				FeedUUID:     feed.UUID,
				UserUUID:     userUUID,
			}
			subscriptions = append(subscriptions, subscription)

			for j := 0; j < 10; j++ {
				entry := Entry{
					UID:      ksuid.New().String(),
					FeedUUID: feed.UUID,
					URL:      fake.Internet().URL(),
					Title:    fake.Lorem().Text(10),
				}
				entries = append(entries, entry)
			}
		}

		r := &fakeRepository{
			Categories:    categories,
			Feeds:         feeds,
			Entries:       entries,
			Subscriptions: subscriptions,
		}
		s := NewService(r, nil)

		if err := s.DeleteCategory(userUUID, category.UUID); err != nil {
			t.Fatalf("want no error, got %q", err)
		}

		if len(r.Categories) != 0 {
			t.Fatalf("want 0 Categories, got %d", len(r.Categories))
		}
		if len(r.Subscriptions) != 0 {
			t.Fatalf("want 0 Subscriptions, got %d", len(r.Subscriptions))
		}
	})
}

func TestServiceUpdateCategory(t *testing.T) {
	userUUID := "179206c8-2965-47a7-ba04-bf0a6a0b8d11"
	now := time.Now().UTC()
	yesterday := now.Add(-24 * time.Hour)

	existingCategory := Category{
		UUID:      "d3033032-23c0-4f78-9b7d-f4135477b5c3",
		UserUUID:  userUUID,
		Name:      "Existing Category",
		Slug:      "existing-category",
		CreatedAt: yesterday,
		UpdatedAt: yesterday,
	}

	cases := []struct {
		tname           string
		updatedCategory Category
		want            Category
		wantErr         error
	}{
		// nominal cases
		{
			tname: "update category with new name and slug",
			updatedCategory: Category{
				UserUUID: existingCategory.UserUUID,
				UUID:     existingCategory.UUID,
				Name:     "New Cat",
			},
			want: Category{
				UserUUID:  existingCategory.UserUUID,
				UUID:      existingCategory.UUID,
				Name:      "New Cat",
				Slug:      "new-cat",
				CreatedAt: yesterday,
				UpdatedAt: now,
			},
		},
		{
			tname: "update category with no changes",
			updatedCategory: Category{
				UserUUID: existingCategory.UserUUID,
				UUID:     existingCategory.UUID,
				Name:     existingCategory.Name,
			},
			want: Category{
				UserUUID:  existingCategory.UserUUID,
				UUID:      existingCategory.UUID,
				Name:      existingCategory.Name,
				Slug:      existingCategory.Slug,
				CreatedAt: yesterday,
				UpdatedAt: now,
			},
		},

		// error cases
		{
			tname: "not found",
			updatedCategory: Category{
				UserUUID: existingCategory.UserUUID,
				UUID:     "f46b4c2c-b7b5-495e-8a92-d58e2d2ae1d0",
			},
			wantErr: ErrCategoryNotFound,
		},
		{
			tname: "invalid UUID",
			updatedCategory: Category{
				UserUUID: existingCategory.UserUUID,
				UUID:     "a-Bc123z",
			},
			wantErr: ErrCategoryUUIDInvalid,
		},
		{
			tname: "empty name",
			updatedCategory: Category{
				UserUUID: existingCategory.UserUUID,
				UUID:     existingCategory.UUID,
			},
			wantErr: ErrCategoryNameRequired,
		},
		{
			tname: "empty name (whitespace)",
			updatedCategory: Category{
				UserUUID: existingCategory.UserUUID,
				UUID:     existingCategory.UUID,
				Name:     "    ",
			},
			wantErr: ErrCategoryNameRequired,
		},
		{
			tname: "empty slug (punctuation)",
			updatedCategory: Category{
				UserUUID: existingCategory.UserUUID,
				UUID:     existingCategory.UUID,
				Name:     "?",
			},
			wantErr: ErrCategorySlugRequired,
		},
	}

	for _, tc := range cases {
		t.Run(tc.tname, func(t *testing.T) {
			r := &fakeRepository{
				Categories: []Category{
					existingCategory,
				},
			}
			s := NewService(r, nil)

			err := s.UpdateCategory(tc.updatedCategory)

			if tc.wantErr != nil {
				if errors.Is(err, tc.wantErr) {
					return
				}
				if err == nil {
					t.Fatalf("want error %q, got nil", tc.wantErr)
				}
				t.Fatalf("want error %q, got %q", tc.wantErr, err)
			}

			if err != nil {
				t.Fatalf("want no error, got %q", err)
			}

			got, err := s.CategoryByUUID(userUUID, tc.updatedCategory.UUID)
			if err != nil {
				t.Fatalf("failed to retrieve category: %q", err)
			}

			assertCategoryEquals(t, got, tc.want)
		})
	}
}

func TestServiceCreateEntries(t *testing.T) {
	feedUUID := "26b0aafc-4de6-46be-91ea-0f6e111c660c"
	now := time.Now().UTC()
	yesterday := now.Add(-24 * time.Hour)

	cases := []struct {
		tname     string
		feedItems []*gofeed.Item
		want      []Entry
		wantErr   error
	}{
		// edge cases
		{
			tname: "publication and update dates not set, default to now",
			feedItems: []*gofeed.Item{
				{
					Link:            "http://test.local/dates",
					Title:           "Date test",
					PublishedParsed: nil,
					UpdatedParsed:   nil,
				},
			},
			want: []Entry{
				{
					FeedUUID:    feedUUID,
					URL:         "http://test.local/dates",
					Title:       "Date test",
					PublishedAt: now,
					UpdatedAt:   now,
				},
			},
		},
		{
			tname: "update date not set, default to published date",
			feedItems: []*gofeed.Item{
				{
					Link:            "http://test.local/dates",
					Title:           "Date test",
					PublishedParsed: &yesterday,
					UpdatedParsed:   nil,
				},
			},
			want: []Entry{
				{
					FeedUUID:    feedUUID,
					URL:         "http://test.local/dates",
					Title:       "Date test",
					PublishedAt: yesterday,
					UpdatedAt:   yesterday,
				},
			},
		},
		{
			tname: "title not set, skip",
			feedItems: []*gofeed.Item{
				{
					Link:  "http://test.local/dates",
					Title: "",
				},
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.tname, func(t *testing.T) {
			r := &fakeRepository{}
			s := NewService(r, nil)

			err := s.createEntries(feedUUID, tc.feedItems)

			if tc.wantErr != nil {
				if errors.Is(err, tc.wantErr) {
					return
				}
				if err == nil {
					t.Fatalf("want error %q, got nil", tc.wantErr)
				}
				t.Fatalf("want error %q, got %q", tc.wantErr, err)
			}

			if err != nil {
				t.Fatalf("want no error, got %q", err)
			}

			assertEntriesEqual(t, r.Entries, tc.want)
		})
	}
}

func TestServiceGetOrCreateFeedAndEntries(t *testing.T) {
	testHTTPClient := &http.Client{
		Transport: testRoundTripper{},
	}

	now := time.Now().UTC()
	yesterday := now.Add(-24 * time.Hour)

	cases := []struct {
		tname             string
		feedURL           string
		repositoryFeeds   []Feed
		repositoryEntries []Entry
		wantFeed          Feed
		wantEntries       []Entry
		wantErr           error
	}{
		// nominal cases
		{
			tname:   "new feed (resolve metadata)",
			feedURL: "http://test.local",
			wantFeed: Feed{
				FeedURL:   "http://test.local",
				Title:     "Local Test",
				Slug:      "local-test",
				CreatedAt: now,
				UpdatedAt: now,
				FetchedAt: now,
			},
			wantEntries: []Entry{
				{
					URL:         "http://test.local/first-post",
					Title:       "First post!",
					PublishedAt: now,
					UpdatedAt:   now,
				},
				{
					URL:         "http://test.local/hello-world",
					Title:       "Hello World",
					PublishedAt: yesterday,
					UpdatedAt:   yesterday,
				},
			},
		},
		{
			tname:   "existing feed (leave unchanged, do not resolve metadata)",
			feedURL: "http://test.local",
			repositoryFeeds: []Feed{
				{
					UUID:      "a8920612-b469-4729-85f3-2c8c30cb897f",
					FeedURL:   "http://test.local",
					Title:     "Existing Test",
					Slug:      "existing-test",
					CreatedAt: yesterday,
					UpdatedAt: yesterday,
					FetchedAt: yesterday,
				},
			},
			repositoryEntries: []Entry{
				{
					FeedUUID:    "a8920612-b469-4729-85f3-2c8c30cb897f",
					URL:         "http://test.local/first-post",
					Title:       "First post!",
					PublishedAt: now,
					UpdatedAt:   now,
				},
				{
					FeedUUID:    "a8920612-b469-4729-85f3-2c8c30cb897f",
					URL:         "http://test.local/hello-world",
					Title:       "Hello World",
					PublishedAt: yesterday,
					UpdatedAt:   yesterday,
				},
			},
			wantFeed: Feed{
				FeedURL:   "http://test.local",
				Title:     "Existing Test",
				Slug:      "existing-test",
				CreatedAt: yesterday,
				UpdatedAt: yesterday,
				FetchedAt: yesterday,
			},
			wantEntries: []Entry{
				{
					FeedUUID:    "a8920612-b469-4729-85f3-2c8c30cb897f",
					URL:         "http://test.local/first-post",
					Title:       "First post!",
					PublishedAt: now,
					UpdatedAt:   now,
				},
				{
					FeedUUID:    "a8920612-b469-4729-85f3-2c8c30cb897f",
					URL:         "http://test.local/hello-world",
					Title:       "Hello World",
					PublishedAt: yesterday,
					UpdatedAt:   yesterday,
				},
			},
		},

		// error cases
		{
			tname:   "empty URL",
			wantErr: ErrFeedURLInvalid,
		},
		{
			tname:   "empty URL (whitespace)",
			feedURL: "     ",
			wantErr: ErrFeedURLInvalid,
		},
		{
			tname:   "invalid URL (no host)",
			feedURL: "http://",
			wantErr: ErrFeedURLNoHost,
		},
		{
			tname:   "invalid URL (no scheme)",
			feedURL: "domain.tld",
			wantErr: ErrFeedURLNoScheme,
		},
		{
			tname:   "invalid URL (unsupported scheme)",
			feedURL: "ftp://domain.tld",
			wantErr: ErrFeedURLUnsupportedScheme,
		},
	}

	for _, tc := range cases {
		t.Run(tc.tname, func(t *testing.T) {
			r := &fakeRepository{
				Entries: tc.repositoryEntries,
				Feeds:   tc.repositoryFeeds,
			}
			feedClient := fetching.NewClient(testHTTPClient, "sparklemuffin/test")

			s := NewService(r, feedClient)

			gotFeed, err := s.getOrCreateFeedAndEntries(tc.feedURL)

			if tc.wantErr != nil {
				if errors.Is(err, tc.wantErr) {
					return
				}
				if err == nil {
					t.Fatalf("want error %q, got nil", tc.wantErr)
				}
				t.Fatalf("want error %q, got %q", tc.wantErr, err)
			}

			if err != nil {
				t.Fatalf("want no error, got %q", err)
			}

			// Update expected FeedUUID
			for i := 0; i < len(tc.wantEntries); i++ {
				tc.wantEntries[i].FeedUUID = gotFeed.UUID
			}

			assertFeedEquals(t, gotFeed, tc.wantFeed)
			assertEntriesEqual(t, r.Entries, tc.wantEntries)
		})
	}
}

func TestServiceToggleEntryRead(t *testing.T) {
	fake := faker.New()

	userUUID := fake.UUID().V4()

	entry1 := Entry{
		UID: ksuid.New().String(),
	}
	entry2 := Entry{
		UID: ksuid.New().String(),
	}

	cases := []struct {
		tname                     string
		repositoryEntries         []Entry
		repositoryEntriesMetadata []EntryMetadata
		entryUID                  string
		want                      []EntryMetadata
		wantErr                   error
	}{
		// nominal cases
		{
			tname: "add entry metadata",
			repositoryEntries: []Entry{
				entry1,
				entry2,
			},
			entryUID: entry2.UID,
			want: []EntryMetadata{
				{
					UserUUID: userUUID,
					EntryUID: entry2.UID,
					Read:     true,
				},
			},
		},
		{
			tname: "update entry metadata",
			repositoryEntries: []Entry{
				entry1,
				entry2,
			},
			repositoryEntriesMetadata: []EntryMetadata{
				{
					UserUUID: userUUID,
					EntryUID: entry1.UID,
					Read:     true,
				},
				{
					UserUUID: userUUID,
					EntryUID: entry2.UID,
					Read:     true,
				},
			},
			entryUID: entry2.UID,
			want: []EntryMetadata{
				{
					UserUUID: userUUID,
					EntryUID: entry1.UID,
					Read:     true,
				},
				{
					UserUUID: userUUID,
					EntryUID: entry2.UID,
					Read:     false,
				},
			},
		},

		// error cases
		{
			tname:    "entry not found",
			entryUID: ksuid.New().String(),
			wantErr:  ErrEntryNotFound,
		},
	}

	for _, tc := range cases {
		t.Run(tc.tname, func(t *testing.T) {
			r := &fakeRepository{
				Entries:         tc.repositoryEntries,
				EntriesMetadata: tc.repositoryEntriesMetadata,
			}
			s := NewService(r, nil)

			err := s.ToggleEntryRead(userUUID, tc.entryUID)

			if tc.wantErr != nil {
				if errors.Is(err, tc.wantErr) {
					return
				}
				if err == nil {
					t.Fatalf("want error %q, got nil", tc.wantErr)
				}
				t.Fatalf("want error %q, got %q", tc.wantErr, err)
			}

			if err != nil {
				t.Fatalf("want no error, got %q", err)
			}

			assertEntriesMetadataEqual(t, r.EntriesMetadata, tc.want)
		})
	}
}

func TestServiceCreateSubscription(t *testing.T) {
	repositoryFeeds := []Feed{
		{
			UUID:    "c2b0fc18-c234-456e-a18c-6d453cfe11ab",
			FeedURL: "http://test.local",
			Title:   "Local Test",
			Slug:    "local-test",
		},
	}

	existingSubscription := Subscription{
		UUID:         "e237792f-33f6-4eeb-a397-36764c8eccb8",
		CategoryUUID: "99cee38d-1eab-4f2a-b598-d341b3b147ab",
		UserUUID:     "a8343f4e-dd9e-4c81-bf5c-1a06d19e1ccf",
		FeedUUID:     "65b6633e-dfa5-4d09-9dc7-7054bd5d731e",
	}
	repositorySubscriptions := []Subscription{existingSubscription}

	cases := []struct {
		tname        string
		subscription Subscription
		wantErr      error
	}{
		// nominal cases
		{
			tname: "new subscription",
			subscription: Subscription{
				UUID:         "0779aef5-269d-4ae3-9658-93427dd04581",
				CategoryUUID: "99cee38d-1eab-4f2a-b598-d341b3b147ab",
				UserUUID:     "a8343f4e-dd9e-4c81-bf5c-1a06d19e1ccf",
				FeedUUID:     "c2b0fc18-c234-456e-a18c-6d453cfe11ab",
			},
		},

		// error cases
		{
			tname:        "existing subscription (identical)",
			subscription: existingSubscription,
			wantErr:      ErrSubscriptionAlreadyRegistered,
		},
		{
			tname: "existing subscription (different category)",
			subscription: Subscription{
				UUID:         existingSubscription.UUID,
				CategoryUUID: "e30f44d0-a1c0-4c85-8f25-b574913ff03d",
				FeedUUID:     existingSubscription.FeedUUID,
				UserUUID:     existingSubscription.UserUUID,
			},
			wantErr: ErrSubscriptionAlreadyRegistered,
		},
	}

	for _, tc := range cases {
		t.Run(tc.tname, func(t *testing.T) {
			r := &fakeRepository{
				Feeds:         repositoryFeeds,
				Subscriptions: repositorySubscriptions,
			}
			s := NewService(r, nil)

			err := s.createSubscription(tc.subscription)

			if tc.wantErr != nil {
				if errors.Is(err, tc.wantErr) {
					return
				}
				if err == nil {
					t.Fatalf("want error %q, got nil", tc.wantErr)
				}
				t.Fatalf("want error %q, got %q", tc.wantErr, err)
			}

			if err != nil {
				t.Fatalf("want no error, got %q", err)
			}
		})
	}
}

func TestServiceDeleteSubscription(t *testing.T) {
	fake := faker.New()
	userUUID := fake.UUID().V4()

	t.Run("subscription with entries", func(t *testing.T) {
		category := Category{
			UserUUID: userUUID,
			UUID:     fake.UUID().V4(),
			Name:     fake.Lorem().Text(10),
			Slug:     fake.Internet().Slug(),
		}

		entries := []Entry{}

		feed := Feed{
			UUID:    fake.UUID().V4(),
			FeedURL: fake.Internet().URL(),
			Title:   fake.Lorem().Text(10),
			Slug:    fake.Internet().Slug(),
		}

		subscription := Subscription{
			UUID:         fake.UUID().V4(),
			CategoryUUID: category.UUID,
			FeedUUID:     feed.UUID,
			UserUUID:     userUUID,
		}

		for j := 0; j < 10; j++ {
			entry := Entry{
				UID:      ksuid.New().String(),
				FeedUUID: feed.UUID,
				URL:      fake.Internet().URL(),
				Title:    fake.Lorem().Text(10),
			}
			entries = append(entries, entry)
		}

		r := &fakeRepository{
			Feeds:         []Feed{feed},
			Entries:       entries,
			Subscriptions: []Subscription{subscription},
		}
		s := NewService(r, nil)

		if err := s.DeleteSubscription(userUUID, subscription.UUID); err != nil {
			t.Fatalf("want no error, got %q", err)
		}

		if len(r.Subscriptions) != 0 {
			t.Fatalf("want 0 Subscriptions, got %d", len(r.Subscriptions))
		}
	})
}
