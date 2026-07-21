// Copyright VirtualTam 2022, 2026
// SPDX-License-Identifier: MIT

package controller

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"

	"github.com/virtualtam/sparklemuffin/internal/http/www/httpcontext"
	"github.com/virtualtam/sparklemuffin/internal/http/www/view"
	"github.com/virtualtam/sparklemuffin/pkg/feed"
	feedquerying "github.com/virtualtam/sparklemuffin/pkg/feed/querying"
	"github.com/virtualtam/sparklemuffin/pkg/user"
)

func newToggleReadRequest(t *testing.T, entryUID string, ctxUser user.User, referer string, form url.Values) *http.Request {
	t.Helper()

	r := httptest.NewRequestWithContext(t.Context(), http.MethodPost, "/feeds/entries/"+entryUID+"/toggle-read", strings.NewReader(form.Encode()))
	r.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	r.Header.Set("Referer", referer)

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("uid", entryUID)

	ctx := context.WithValue(r.Context(), chi.RouteCtxKey, rctx)
	ctx = httpcontext.WithUser(ctx, ctxUser)

	return r.WithContext(ctx)
}

func TestHandleFeedEntryToggleRead(t *testing.T) {
	ctxUser := user.User{UUID: "user-1"}

	category := feed.Category{UUID: "category-1", UserUUID: ctxUser.UUID, Name: "Tech", Slug: "tech"}
	aFeed := feed.Feed{UUID: "feed-1", Title: "Blog", Slug: "blog"}
	subscription := feed.Subscription{UUID: "sub-1", UserUUID: ctxUser.UUID, CategoryUUID: category.UUID, FeedUUID: aFeed.UUID}
	entry := feed.Entry{UID: "entry-1", FeedUUID: aFeed.UUID, URL: "https://example.com/1", Title: "Post 1"}

	// entriesMetadata is shared, by reference, between the feed and querying fake
	// repositories: in production both services read/write the same
	// feed_entries_metadata table, so a toggle made through feedService must be
	// observable through queryingService. Passing the same non-empty slice to both
	// fakes keeps ToggleEntryRead on its in-place update path (rather than the
	// append-a-new-row path), so the mutation is visible through both fakes.
	newController := func(preferences feed.Preferences, entriesMetadata []feed.EntryMetadata) feedController {
		feedRepo := &feed.FakeRepository{
			Categories:      []feed.Category{category},
			Entries:         []feed.Entry{entry},
			EntriesMetadata: entriesMetadata,
			Feeds:           []feed.Feed{aFeed},
			Preferences:     map[string]feed.Preferences{ctxUser.UUID: preferences},
			Subscriptions:   []feed.Subscription{subscription},
		}

		queryingRepo := &feedquerying.FakeRepository{
			Categories:      []feed.Category{category},
			Entries:         []feed.Entry{entry},
			EntriesMetadata: entriesMetadata,
			Feeds:           []feed.Feed{aFeed},
			Subscriptions:   []feed.Subscription{subscription},
		}

		return feedController{
			feedService:     feed.NewService(feedRepo, nil),
			queryingService: feedquerying.NewService(queryingRepo),
			feedListView:    view.New("feed/feed_list.gohtml"),
		}
	}

	unreadMetadata := func() []feed.EntryMetadata {
		return []feed.EntryMetadata{{UserUUID: ctxUser.UUID, EntryUID: entry.UID, Read: false}}
	}

	t.Run("success, entry stays visible under the current filter", func(t *testing.T) {
		fc := newController(feed.Preferences{UserUUID: ctxUser.UUID, ShowEntries: feed.EntryVisibilityAll}, unreadMetadata())

		form := url.Values{"urlPath": {"/feeds"}, "search": {""}, "page": {"1"}}
		r := newToggleReadRequest(t, entry.UID, ctxUser, "/feeds", form)
		w := httptest.NewRecorder()

		fc.handleFeedEntryToggleRead()(w, r)

		if w.Code != http.StatusOK {
			t.Fatalf("want status 200, got %d, body:\n%s", w.Code, w.Body.String())
		}

		body := w.Body.String()
		wantContains := []string{
			`id="feed-entry-entry-1"`,
			"has-text-grey-light", // now marked as read
			"Mark as unread",
			`id="unread-count-all"`,
			`id="unread-count-category-tech"`,
			`id="unread-count-feed-blog"`,
			`id="entry-count"`,
			`hx-swap-oob="true"`,
			// the re-rendered button must carry the same context forward, so a
			// second click still recomputes counts against the right view
			`hx-post="/feeds/entries/entry-1/toggle-read"`,
			`urlPath&#34;:&#34;/feeds`,
			`page&#34;:1`,
		}
		for _, want := range wantContains {
			if !strings.Contains(body, want) {
				t.Errorf("want body to contain %q, got:\n%s", want, body)
			}
		}
	})

	t.Run("success, entry no longer matches the current filter", func(t *testing.T) {
		fc := newController(feed.Preferences{UserUUID: ctxUser.UUID, ShowEntries: feed.EntryVisibilityUnread}, unreadMetadata())

		form := url.Values{"urlPath": {"/feeds"}, "search": {""}, "page": {"1"}}
		r := newToggleReadRequest(t, entry.UID, ctxUser, "/feeds", form)
		w := httptest.NewRecorder()

		fc.handleFeedEntryToggleRead()(w, r)

		if w.Code != http.StatusOK {
			t.Fatalf("want status 200, got %d, body:\n%s", w.Code, w.Body.String())
		}

		body := w.Body.String()
		if strings.Contains(body, `id="feed-entry-entry-1"`) {
			t.Errorf("want entry removed from the response, got:\n%s", body)
		}
		if !strings.Contains(body, `id="unread-count-all"`) {
			t.Errorf("want unread counts to still be refreshed, got:\n%s", body)
		}
	})

	t.Run("success, category context recomputes the matching entry count", func(t *testing.T) {
		fc := newController(feed.Preferences{UserUUID: ctxUser.UUID, ShowEntries: feed.EntryVisibilityAll}, unreadMetadata())

		form := url.Values{"urlPath": {"/feeds/categories/tech"}, "search": {""}, "page": {"1"}}
		r := newToggleReadRequest(t, entry.UID, ctxUser, "/feeds/categories/tech", form)
		w := httptest.NewRecorder()

		fc.handleFeedEntryToggleRead()(w, r)

		if w.Code != http.StatusOK {
			t.Fatalf("want status 200, got %d, body:\n%s", w.Code, w.Body.String())
		}

		if body := w.Body.String(); !strings.Contains(body, `id="entry-count"`) {
			t.Errorf("want entry count fragment rendered, got:\n%s", body)
		}
	})

	t.Run("error falls back to a full reload", func(t *testing.T) {
		fc := newController(feed.Preferences{UserUUID: ctxUser.UUID, ShowEntries: feed.EntryVisibilityAll}, nil)

		form := url.Values{"urlPath": {"/feeds"}, "search": {""}, "page": {"1"}}
		r := newToggleReadRequest(t, "does-not-exist", ctxUser, "/feeds", form)
		w := httptest.NewRecorder()

		fc.handleFeedEntryToggleRead()(w, r)

		if w.Code != http.StatusSeeOther {
			t.Fatalf("want status 303, got %d", w.Code)
		}
		if got := w.Header().Get("Location"); got != "/feeds" {
			t.Errorf("want redirect to %q, got %q", "/feeds", got)
		}

		foundFlashCookie := false
		for _, cookie := range w.Result().Cookies() {
			if cookie.Name == "flash" {
				foundFlashCookie = true
			}
		}
		if !foundFlashCookie {
			t.Error("want a flash cookie to be set")
		}
	})
}
