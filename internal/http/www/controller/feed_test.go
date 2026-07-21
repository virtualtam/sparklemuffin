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

var (
	testCtxUser      = user.User{UUID: "user-1"}
	testCategory     = feed.Category{UUID: "category-1", UserUUID: testCtxUser.UUID, Name: "Tech", Slug: "tech"}
	testFeed         = feed.Feed{UUID: "feed-1", Title: "Blog", Slug: "blog"}
	testSubscription = feed.Subscription{UUID: "sub-1", UserUUID: testCtxUser.UUID, CategoryUUID: testCategory.UUID, FeedUUID: testFeed.UUID}
	testEntry        = feed.Entry{UID: "entry-1", FeedUUID: testFeed.UUID, URL: "https://example.com/1", Title: "Post 1"}
)

// newTestFeedController wires a feedController against feed and querying fake
// repositories sharing the same fixtures (one category, one feed, one
// subscription, one entry).
//
// entriesMetadata is shared, by reference, between the feed and querying fake
// repositories: in production both services read/write the same
// feed_entries_metadata table, so a mutation made through feedService must be
// observable through queryingService. Passing the same non-empty slice to both
// fakes keeps read/unread updates on their in-place update path (rather than
// the append-a-new-row path), so the mutation is visible through both fakes.
func newTestFeedController(preferences feed.Preferences, entriesMetadata []feed.EntryMetadata) feedController {
	feedRepo := &feed.FakeRepository{
		Categories:      []feed.Category{testCategory},
		Entries:         []feed.Entry{testEntry},
		EntriesMetadata: entriesMetadata,
		Feeds:           []feed.Feed{testFeed},
		Preferences:     map[string]feed.Preferences{testCtxUser.UUID: preferences},
		Subscriptions:   []feed.Subscription{testSubscription},
	}

	queryingRepo := &feedquerying.FakeRepository{
		Categories:      []feed.Category{testCategory},
		Entries:         []feed.Entry{testEntry},
		EntriesMetadata: entriesMetadata,
		Feeds:           []feed.Feed{testFeed},
		Subscriptions:   []feed.Subscription{testSubscription},
	}

	return feedController{
		feedService:     feed.NewService(feedRepo, nil),
		queryingService: feedquerying.NewService(queryingRepo),
		feedListView:    view.New("feed/feed_list.gohtml"),
	}
}

func testUnreadMetadata() []feed.EntryMetadata {
	return []feed.EntryMetadata{{UserUUID: testCtxUser.UUID, EntryUID: testEntry.UID, Read: false}}
}

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

// newFeedPostRequest builds a POST request against a route with no chi URL
// params (e.g. the preferences endpoints), with the given user set in context.
func newFeedPostRequest(t *testing.T, path string, ctxUser user.User, form url.Values) *http.Request {
	t.Helper()

	r := httptest.NewRequestWithContext(t.Context(), http.MethodPost, path, strings.NewReader(form.Encode()))
	r.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	r.Header.Set("Referer", "/feeds")

	ctx := httpcontext.WithUser(r.Context(), ctxUser)

	return r.WithContext(ctx)
}

func TestHandleFeedEntryToggleRead(t *testing.T) {
	ctxUser := testCtxUser
	entry := testEntry

	t.Run("success, entry stays visible under the current filter", func(t *testing.T) {
		fc := newTestFeedController(feed.Preferences{UserUUID: ctxUser.UUID, ShowEntries: feed.EntryVisibilityAll}, testUnreadMetadata())

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
		fc := newTestFeedController(feed.Preferences{UserUUID: ctxUser.UUID, ShowEntries: feed.EntryVisibilityUnread}, testUnreadMetadata())

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
		fc := newTestFeedController(feed.Preferences{UserUUID: ctxUser.UUID, ShowEntries: feed.EntryVisibilityAll}, testUnreadMetadata())

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
		fc := newTestFeedController(feed.Preferences{UserUUID: ctxUser.UUID, ShowEntries: feed.EntryVisibilityAll}, nil)

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

func TestHandlePreferencesFeedShowEntriesUpdate(t *testing.T) {
	ctxUser := testCtxUser
	entry := testEntry

	t.Run("success, list is re-rendered for the new filter", func(t *testing.T) {
		// The entry starts read and visible under "All"; switching to "Unread
		// only" must drop it from the re-rendered list.
		readMetadata := []feed.EntryMetadata{{UserUUID: ctxUser.UUID, EntryUID: entry.UID, Read: true}}
		fc := newTestFeedController(feed.Preferences{UserUUID: ctxUser.UUID, ShowEntries: feed.EntryVisibilityAll}, readMetadata)

		form := url.Values{"show": {"UNREAD"}, "urlPath": {"/feeds"}, "search": {""}}
		r := newFeedPostRequest(t, "/feeds/preferences/show-entries", ctxUser, form)
		w := httptest.NewRecorder()

		fc.handlePreferencesFeedShowEntriesUpdate()(w, r)

		if w.Code != http.StatusOK {
			t.Fatalf("want status 200, got %d, body:\n%s", w.Code, w.Body.String())
		}
		if got := w.Header().Get("HX-Refresh"); got != "" {
			t.Errorf("want no HX-Refresh header, got %q", got)
		}

		body := w.Body.String()

		if strings.Contains(body, `id="feed-entry-entry-1"`) {
			t.Errorf("want the read entry dropped from the Unread-only view, got:\n%s", body)
		}

		wantContains := []string{
			`<ol id="entry-list"`,
			`id="unread-count-all"`,
			`id="unread-count-category-tech"`,
			`id="unread-count-feed-blog"`,
			`id="entry-count"`,
			`id="pagination-top"`,
			`id="pagination-bottom"`,
			`hx-vals='{&#34;search&#34;:&#34;&#34;,&#34;show&#34;:&#34;UNREAD&#34;,&#34;urlPath&#34;:&#34;/feeds&#34;}'`,
		}
		for _, want := range wantContains {
			if !strings.Contains(body, want) {
				t.Errorf("want body to contain %q, got:\n%s", want, body)
			}
		}

		// 2: the Unread filter button, plus the Compact button (also always
		// re-rendered by the shared helper, active by default since
		// ShowEntrySummaries is zero-valued/false in this fixture)
		if got := strings.Count(body, "is-active"); got != 2 {
			t.Errorf("want exactly 2 active buttons (Unread + Compact), got %d, body:\n%s", got, body)
		}
	})

	t.Run("error falls back to a full reload", func(t *testing.T) {
		fc := newTestFeedController(feed.Preferences{UserUUID: ctxUser.UUID, ShowEntries: feed.EntryVisibilityAll}, nil)

		unknownUser := user.User{UUID: "no-such-user"}
		form := url.Values{"show": {"UNREAD"}, "urlPath": {"/feeds"}, "search": {""}}
		r := newFeedPostRequest(t, "/feeds/preferences/show-entries", unknownUser, form)
		w := httptest.NewRecorder()

		fc.handlePreferencesFeedShowEntriesUpdate()(w, r)

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

func TestHandlePreferencesToggleShowEntrySummaries(t *testing.T) {
	ctxUser := testCtxUser

	t.Run("success, list is re-rendered and the current page is preserved", func(t *testing.T) {
		fc := newTestFeedController(feed.Preferences{
			UserUUID:           ctxUser.UUID,
			ShowEntries:        feed.EntryVisibilityAll,
			ShowEntrySummaries: true,
		}, nil)

		form := url.Values{"urlPath": {"/feeds"}, "search": {""}, "page": {"1"}}
		r := newFeedPostRequest(t, "/feeds/preferences/toggle-show-entry-summaries", ctxUser, form)
		w := httptest.NewRecorder()

		fc.handlePreferencesToggleShowEntrySummaries()(w, r)

		if w.Code != http.StatusOK {
			t.Fatalf("want status 200, got %d, body:\n%s", w.Code, w.Body.String())
		}
		if got := w.Header().Get("HX-Refresh"); got != "" {
			t.Errorf("want no HX-Refresh header, got %q", got)
		}

		body := w.Body.String()

		wantContains := []string{
			`<ol id="entry-list"`,
			`id="compact-button"`,
			`id="unread-count-all"`,
			`id="entry-count"`,
			`id="pagination-top"`,
			`id="pagination-bottom"`,
		}
		for _, want := range wantContains {
			if !strings.Contains(body, want) {
				t.Errorf("want body to contain %q, got:\n%s", want, body)
			}
		}

		// ShowEntrySummaries flips from true to false, so the Compact button
		// (now hiding summaries) becomes active; the "All" filter button is
		// also active since ShowEntries stays ALL. 2 total.
		if got := strings.Count(body, "is-active"); got != 2 {
			t.Errorf("want exactly 2 active buttons (All + Compact), got %d, body:\n%s", got, body)
		}
	})

	t.Run("error falls back to a full reload", func(t *testing.T) {
		fc := newTestFeedController(feed.Preferences{UserUUID: ctxUser.UUID}, nil)

		unknownUser := user.User{UUID: "no-such-user"}
		form := url.Values{"urlPath": {"/feeds"}, "search": {""}, "page": {"1"}}
		r := newFeedPostRequest(t, "/feeds/preferences/toggle-show-entry-summaries", unknownUser, form)
		w := httptest.NewRecorder()

		fc.handlePreferencesToggleShowEntrySummaries()(w, r)

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
