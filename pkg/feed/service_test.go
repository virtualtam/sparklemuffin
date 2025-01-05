// Copyright (c) VirtualTam
// SPDX-License-Identifier: MIT

package feed

import (
	"errors"
	"net/http"
	"testing"
	"time"

	"github.com/cespare/xxhash/v2"
	"github.com/jaswdr/faker"
	"github.com/mmcdole/gofeed"
	"github.com/segmentio/ksuid"
	"github.com/virtualtam/sparklemuffin/internal/test/feedtest"
	"github.com/virtualtam/sparklemuffin/pkg/feed/fetching"
)

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
			r := &FakeRepository{
				Categories: tc.repositoryCategories,
			}
			s := NewService(r, nil)

			got, err := s.CreateCategory(userUUID, tc.name)

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
			r := &FakeRepository{
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
			r := &FakeRepository{
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

		r := &FakeRepository{
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

		r := &FakeRepository{
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

		// Ensure the subscription deletion is propagated to feeds
		// and entries (each feed only has one subscription)
		if len(r.Feeds) != 0 {
			t.Fatalf("want 0 Feeds, got %d", len(r.Feeds))
		}
		if len(r.Entries) != 0 {
			t.Fatalf("want 0 Entries, got %d", len(r.Entries))
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
			r := &FakeRepository{
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
		// nominal cases
		{
			tname: "entry with long description - expecting a Summary and TextRankTerms",
			feedItems: []*gofeed.Item{
				{
					// The description for this entry is a plain-text extract from the Wikipedia article
					// https://en.wikipedia.org/wiki/History_of_aluminium
					//
					// This article is licensed under the Creative Commons Attribution-Share-Alike License 4.0.
					Link:  "http://test.local/dates",
					Title: "History of aluminium",
					Description: `
Aluminium (or aluminum) metal is very rare in native form, and the process to refine it from ores is complex, so for most of human history it was unknown. However, the compound alum has been known since the 5th century BCE and was used extensively by the ancients for dyeing. During the Middle Ages, its use for dyeing made it a commodity of international commerce. Renaissance scientists believed that alum was a salt of a new earth; during the Age of Enlightenment, it was established that this earth, alumina, was an oxide of a new metal. Discovery of this metal was announced in 1825 by Danish physicist Hans Christian Ørsted, whose work was extended by German chemist Friedrich Wöhler.
Aluminium was difficult to refine and thus uncommon in actual use. Soon after its discovery, the price of aluminium exceeded that of gold. It was reduced only after the initiation of the first industrial production by French chemist Henri Étienne Sainte-Claire Deville in 1856. Aluminium became much more available to the public with the Hall–Héroult process developed independently by French engineer Paul Héroult and American engineer Charles Martin Hall in 1886, and the Bayer process developed by Austrian chemist Carl Joseph Bayer in 1889. These processes have been used for aluminium production up to the present.
The introduction of these methods for the mass production of aluminium led to extensive use of the light, corrosion-resistant metal in industry and everyday life. Aluminium began to be used in engineering and construction. In World Wars I and II, aluminium was a crucial strategic resource for aviation. World production of the metal grew from 6,800 metric tons in 1900 to 2,810,000 metric tons in 1954, when aluminium became the most produced non-ferrous metal, surpassing copper.
In the second half of the 20th century, aluminium gained usage in transportation and packaging. Aluminium production became a source of concern due to its effect on the environment, and aluminium recycling gained ground. The metal became an exchange commodity in the 1970s. Production began to shift from developed countries to developing ones; by 2010, China had accumulated an especially large share in both production and consumption of aluminium. World production continued to rise, reaching 58,500,000 metric tons in 2015. Aluminium production exceeds those of all other non-ferrous metals combined.
`,
					PublishedParsed: &now,
					UpdatedParsed:   &now,
				},
			},
			want: []Entry{
				{
					FeedUUID: feedUUID,
					URL:      "http://test.local/dates",
					Title:    "History of aluminium",
					Summary:  `Aluminium (or aluminum) metal is very rare in native form, and the process to refine it from ores is complex, so for most of human history it was unknown. However, the compound alum has been known since the 5th century BCE and was used extensively by the ancients for dyeing. During the Middle Ages, its use for dyeing made it a commodity of international commerce. Renaissance scientists believed th…`,
					TextRankTerms: []string{
						"production aluminium", "tons metric",
						"non ferrous", "metric 000",
						"developed process", "production world",
						"metals combined", "metals ferrous",
						"500 reaching", "500 000",
					},
					PublishedAt: now,
					UpdatedAt:   now,
				},
			},
		},

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
			r := &FakeRepository{}
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

			AssertEntriesEqual(t, r.Entries, tc.want)
		})
	}
}

func TestServiceGetOrCreateFeedAndEntries(t *testing.T) {
	now := time.Now().UTC()
	yesterday := now.Add(-24 * time.Hour)

	feed := feedtest.GenerateDummyFeed(t, now)
	feedStr, err := feed.ToAtom()
	if err != nil {
		t.Fatalf("failed to encode feed to Atom: %q", err)
	}
	feedETag := feedtest.HashETag(feedStr)
	feedLastModified := now
	feedHash := xxhash.Sum64String(feedStr)

	transport := feedtest.NewRoundTripper(t, feed)

	testHTTPClient := &http.Client{
		Transport: transport,
	}

	cases := []struct {
		tname   string
		feedURL string

		repositoryFeeds   []Feed
		repositoryEntries []Entry

		wantFeed      Feed
		wantIsCreated bool
		wantEntries   []Entry
		wantErr       error
	}{
		// nominal cases
		{
			tname:   "new feed (resolve metadata)",
			feedURL: "http://test.local",
			wantFeed: Feed{
				FeedURL:      "http://test.local",
				Title:        "Local Test",
				Description:  "A simple syndication feed, for testing purposes.",
				Slug:         "local-test",
				ETag:         feedETag,
				LastModified: feedLastModified,
				Hash:         feedHash,
				CreatedAt:    now,
				UpdatedAt:    now,
				FetchedAt:    now,
			},
			wantIsCreated: true,
			wantEntries: []Entry{
				{
					URL:         "http://test.local/first-post",
					Title:       "First post!",
					Summary:     "First post!\n\nThis is the first post!",
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
					UUID:         "a8920612-b469-4729-85f3-2c8c30cb897f",
					FeedURL:      "http://test.local",
					Title:        "Existing Test",
					Description:  "A simple syndication feed, for testing purposes.",
					Slug:         "existing-test",
					ETag:         feedETag,
					LastModified: feedLastModified,
					CreatedAt:    yesterday,
					UpdatedAt:    yesterday,
					FetchedAt:    yesterday,
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
				FeedURL:      "http://test.local",
				Title:        "Existing Test",
				Slug:         "existing-test",
				Description:  "A simple syndication feed, for testing purposes.",
				ETag:         feedETag,
				LastModified: feedLastModified,
				CreatedAt:    yesterday,
				UpdatedAt:    yesterday,
				FetchedAt:    yesterday,
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
			r := &FakeRepository{
				Entries: tc.repositoryEntries,
				Feeds:   tc.repositoryFeeds,
			}
			feedClient := fetching.NewClient(testHTTPClient, "sparklemuffin/test")

			s := NewService(r, feedClient)

			gotFeed, gotIsCreated, err := s.GetOrCreateFeedAndEntries(tc.feedURL)

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

			if gotIsCreated != tc.wantIsCreated {
				t.Errorf("want isCreated %t, got %t", tc.wantIsCreated, gotIsCreated)
			}

			AssertFeedEquals(t, gotFeed, tc.wantFeed)
			AssertEntriesEqual(t, r.Entries, tc.wantEntries)
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
			r := &FakeRepository{
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
			r := &FakeRepository{
				Feeds:         repositoryFeeds,
				Subscriptions: repositorySubscriptions,
			}
			s := NewService(r, nil)

			_, err := s.createSubscription(tc.subscription)

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
			UUID: fake.UUID().V4(),
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

		r := &FakeRepository{
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

		if _, err := r.FeedGetByURL(feed.FeedURL); !errors.Is(err, ErrFeedNotFound) {
			t.Fatalf("want ErrFeedNotFound, got %q", err)
		}
	})
}

func TestServiceUpdateSubscription(t *testing.T) {
	fake := faker.New()
	now := time.Now().UTC()
	yesterday := now.Add(-24 * time.Hour)

	userUUID := fake.UUID().V4()
	category1UUID := fake.UUID().V4()
	category2UUID := fake.UUID().V4()
	feedUUID := fake.UUID().V4()
	subscriptionUUID := fake.UUID().V4()

	testSubscription := Subscription{
		UUID:         subscriptionUUID,
		CategoryUUID: category1UUID,
		FeedUUID:     feedUUID,
		UserUUID:     userUUID,
		CreatedAt:    yesterday,
		UpdatedAt:    yesterday,
	}

	testSubscriptions := []Subscription{
		testSubscription,
	}

	cases := []struct {
		tname                   string
		repositorySubscriptions []Subscription
		subscription            Subscription
		wantSubscriptions       []Subscription
		wantErr                 error
	}{
		// nominal cases
		{
			tname:                   "updated with no change",
			repositorySubscriptions: testSubscriptions,
			subscription: Subscription{
				UUID:         subscriptionUUID,
				CategoryUUID: category1UUID,
				FeedUUID:     feedUUID,
				UserUUID:     userUUID,
			},
			wantSubscriptions: []Subscription{
				{
					UUID:         subscriptionUUID,
					CategoryUUID: category1UUID,
					FeedUUID:     feedUUID,
					UserUUID:     userUUID,
					CreatedAt:    yesterday,
					UpdatedAt:    now,
				},
			},
		},
		{
			tname:                   "update category",
			repositorySubscriptions: testSubscriptions,
			subscription: Subscription{
				UUID:         subscriptionUUID,
				CategoryUUID: category2UUID,
				FeedUUID:     feedUUID,
				UserUUID:     userUUID,
			},
			wantSubscriptions: []Subscription{
				{
					UUID:         subscriptionUUID,
					CategoryUUID: category2UUID,
					FeedUUID:     feedUUID,
					UserUUID:     userUUID,
					CreatedAt:    yesterday,
					UpdatedAt:    now,
				},
			},
		},
		{
			tname:                   "add alias",
			repositorySubscriptions: testSubscriptions,
			subscription: Subscription{
				UUID:         subscriptionUUID,
				CategoryUUID: category1UUID,
				FeedUUID:     feedUUID,
				UserUUID:     userUUID,
				Alias:        " I would prefer this feed to be displayed with this alias ",
			},
			wantSubscriptions: []Subscription{
				{
					UUID:         subscriptionUUID,
					CategoryUUID: category1UUID,
					FeedUUID:     feedUUID,
					UserUUID:     userUUID,
					Alias:        "I would prefer this feed to be displayed with this alias",
					CreatedAt:    yesterday,
					UpdatedAt:    now,
				},
			},
		},

		// error cases
		{
			tname: "not found",
			subscription: Subscription{
				UUID:     fake.UUID().V4(),
				UserUUID: userUUID,
			},
			wantErr: ErrSubscriptionNotFound,
		},
	}

	for _, tc := range cases {
		t.Run(tc.tname, func(t *testing.T) {
			r := &FakeRepository{
				Subscriptions: tc.repositorySubscriptions,
			}
			s := NewService(r, nil)

			err := s.UpdateSubscription(tc.subscription)

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

			AssertSubscriptionsEqual(t, tc.wantSubscriptions, r.Subscriptions)
		})
	}
}
