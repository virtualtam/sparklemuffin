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
	"github.com/jaswdr/faker/v2"

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
		feedService:              feed.NewService(feedRepo, nil),
		queryingService:          feedquerying.NewService(queryingRepo),
		feedListView:             view.New("feed/feed_list.gohtml"),
		feedSubscriptionListView: view.New("feed/subscription_list.gohtml"),
		feedCategoryEditView:     view.New("feed/category_edit.gohtml"),
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
	r.Header.Set("HX-Request", "true")

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
	r.Header.Set("HX-Request", "true")

	ctx := httpcontext.WithUser(r.Context(), ctxUser)

	return r.WithContext(ctx)
}

// newFeedSlugPostRequest builds a POST request against a route with a chi
// "slug" URL param (e.g. the category/subscription mark-all-read endpoints).
func newFeedSlugPostRequest(t *testing.T, path string, slug string, ctxUser user.User, form url.Values) *http.Request {
	t.Helper()

	r := httptest.NewRequestWithContext(t.Context(), http.MethodPost, path, strings.NewReader(form.Encode()))
	r.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	r.Header.Set("Referer", "/feeds")
	r.Header.Set("HX-Request", "true")

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("slug", slug)

	ctx := context.WithValue(r.Context(), chi.RouteCtxKey, rctx)
	ctx = httpcontext.WithUser(ctx, ctxUser)

	return r.WithContext(ctx)
}

// assertHXRedirectOnError checks the error-fallback contract shared by every
// htmx fragment endpoint in this file: on failure, the response must force a
// full client-side navigation via HX-Redirect (a plain 3xx would instead get
// silently followed by the browser and have the *target* page's full HTML
// swapped into the fragment's hx-target, corrupting the DOM), carrying a
// flash message the user will see once that navigation lands.
func assertHXRedirectOnError(t *testing.T, w *httptest.ResponseRecorder, wantRedirectTo string) {
	t.Helper()

	if w.Code != http.StatusOK {
		t.Fatalf("want status 200, got %d, body:\n%s", w.Code, w.Body.String())
	}
	if got := w.Header().Get("HX-Redirect"); got != wantRedirectTo {
		t.Errorf("want HX-Redirect to %q, got %q", wantRedirectTo, got)
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
}

// assertRejectsNonHxRequest checks that a request missing HX-Request is
// rejected outright: these handlers only ever render an HTML fragment, never
// a full page layout, so a request that isn't provably htmx-issued has no
// well-defined response to send.
func assertRejectsNonHxRequest(t *testing.T, w *httptest.ResponseRecorder) {
	t.Helper()

	if w.Code != http.StatusBadRequest {
		t.Fatalf("want status 400, got %d, body:\n%s", w.Code, w.Body.String())
	}
}

func TestHandleHxFeedEntryToggleRead(t *testing.T) {
	ctxUser := testCtxUser
	entry := testEntry

	t.Run("success, entry stays visible under the current filter", func(t *testing.T) {
		fc := newTestFeedController(feed.Preferences{UserUUID: ctxUser.UUID, ShowEntries: feed.EntryVisibilityAll}, testUnreadMetadata())

		form := url.Values{"urlPath": {"/feeds"}, "search": {""}, "page": {"1"}}
		r := newToggleReadRequest(t, entry.UID, ctxUser, "/feeds", form)
		w := httptest.NewRecorder()

		fc.handleHxFeedEntryToggleRead()(w, r)

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

		fc.handleHxFeedEntryToggleRead()(w, r)

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

		fc.handleHxFeedEntryToggleRead()(w, r)

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

		fc.handleHxFeedEntryToggleRead()(w, r)

		assertHXRedirectOnError(t, w, "/feeds")
	})

	t.Run("error retrieving the entry falls back to the referring page", func(t *testing.T) {
		// The entry lookup failure isn't caused by which view the user was on,
		// so unlike a feedPageForContext failure, this must redirect back to
		// wherever the user came from rather than bouncing to the feeds root.
		fc := newTestFeedController(feed.Preferences{UserUUID: ctxUser.UUID, ShowEntries: feed.EntryVisibilityAll}, nil)

		form := url.Values{"urlPath": {"/feeds/categories/tech"}, "search": {""}, "page": {"1"}}
		r := newToggleReadRequest(t, "does-not-exist", ctxUser, "/feeds/categories/tech", form)
		w := httptest.NewRecorder()

		fc.handleHxFeedEntryToggleRead()(w, r)

		assertHXRedirectOnError(t, w, "/feeds/categories/tech")
	})

	t.Run("error resolving the view context falls back to the feeds root", func(t *testing.T) {
		// A bad category slug in urlPath means the view itself is broken, so
		// redirecting back to it (via Referer) would just fail again: this
		// must always bounce to the feeds root instead, regardless of Referer.
		fc := newTestFeedController(feed.Preferences{UserUUID: ctxUser.UUID, ShowEntries: feed.EntryVisibilityAll}, testUnreadMetadata())

		form := url.Values{"urlPath": {"/feeds/categories/does-not-exist"}, "search": {""}, "page": {"1"}}
		r := newToggleReadRequest(t, entry.UID, ctxUser, "/feeds/categories/does-not-exist", form)
		w := httptest.NewRecorder()

		fc.handleHxFeedEntryToggleRead()(w, r)

		assertHXRedirectOnError(t, w, "/feeds")
	})

	t.Run("rejects requests missing HX-Request", func(t *testing.T) {
		fc := newTestFeedController(feed.Preferences{UserUUID: ctxUser.UUID, ShowEntries: feed.EntryVisibilityAll}, testUnreadMetadata())

		form := url.Values{"urlPath": {"/feeds"}, "search": {""}, "page": {"1"}}
		r := newToggleReadRequest(t, entry.UID, ctxUser, "/feeds", form)
		r.Header.Del("HX-Request")
		w := httptest.NewRecorder()

		fc.handleHxFeedEntryToggleRead()(w, r)

		assertRejectsNonHxRequest(t, w)
	})
}

func TestHandleHxPreferencesFeedShowEntriesUpdate(t *testing.T) {
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

		fc.handleHxPreferencesFeedShowEntriesUpdate()(w, r)

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

		fc.handleHxPreferencesFeedShowEntriesUpdate()(w, r)

		assertHXRedirectOnError(t, w, "/feeds")
	})

	t.Run("rejects requests missing HX-Request", func(t *testing.T) {
		fc := newTestFeedController(feed.Preferences{UserUUID: ctxUser.UUID, ShowEntries: feed.EntryVisibilityAll}, nil)

		form := url.Values{"show": {"UNREAD"}, "urlPath": {"/feeds"}, "search": {""}}
		r := newFeedPostRequest(t, "/feeds/preferences/show-entries", ctxUser, form)
		r.Header.Del("HX-Request")
		w := httptest.NewRecorder()

		fc.handleHxPreferencesFeedShowEntriesUpdate()(w, r)

		assertRejectsNonHxRequest(t, w)
	})
}

func TestHandleHxPreferencesToggleShowEntrySummaries(t *testing.T) {
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

		fc.handleHxPreferencesToggleShowEntrySummaries()(w, r)

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

		fc.handleHxPreferencesToggleShowEntrySummaries()(w, r)

		assertHXRedirectOnError(t, w, "/feeds")
	})

	t.Run("rejects requests missing HX-Request", func(t *testing.T) {
		fc := newTestFeedController(feed.Preferences{UserUUID: ctxUser.UUID}, nil)

		form := url.Values{"urlPath": {"/feeds"}, "search": {""}, "page": {"1"}}
		r := newFeedPostRequest(t, "/feeds/preferences/toggle-show-entry-summaries", ctxUser, form)
		r.Header.Del("HX-Request")
		w := httptest.NewRecorder()

		fc.handleHxPreferencesToggleShowEntrySummaries()(w, r)

		assertRejectsNonHxRequest(t, w)
	})
}

func TestHandleHxEntryMetadataMarkAllAsRead(t *testing.T) {
	ctxUser := testCtxUser
	entry := testEntry

	t.Run("success, list is re-rendered and now-read entries drop from the Unread view", func(t *testing.T) {
		fc := newTestFeedController(feed.Preferences{UserUUID: ctxUser.UUID, ShowEntries: feed.EntryVisibilityUnread}, testUnreadMetadata())

		form := url.Values{"urlPath": {"/feeds"}, "search": {""}}
		r := newFeedPostRequest(t, "/feeds/entries/mark-all-read", ctxUser, form)
		w := httptest.NewRecorder()

		fc.handleHxEntryMetadataMarkAllAsRead()(w, r)

		if w.Code != http.StatusOK {
			t.Fatalf("want status 200, got %d, body:\n%s", w.Code, w.Body.String())
		}
		if got := w.Header().Get("HX-Refresh"); got != "" {
			t.Errorf("want no HX-Refresh header, got %q", got)
		}

		body := w.Body.String()

		if strings.Contains(body, `id="feed-entry-`+entry.UID+`"`) {
			t.Errorf("want the now-read entry dropped from the Unread-only view, got:\n%s", body)
		}

		wantContains := []string{
			`<ol id="entry-list"`,
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
	})

	t.Run("error falls back to a full reload", func(t *testing.T) {
		fc := newTestFeedController(feed.Preferences{UserUUID: ctxUser.UUID, ShowEntries: feed.EntryVisibilityAll}, nil)

		unknownUser := user.User{UUID: "no-such-user"}
		form := url.Values{"urlPath": {"/feeds"}, "search": {""}}
		r := newFeedPostRequest(t, "/feeds/entries/mark-all-read", unknownUser, form)
		w := httptest.NewRecorder()

		fc.handleHxEntryMetadataMarkAllAsRead()(w, r)

		assertHXRedirectOnError(t, w, "/feeds")
	})

	t.Run("rejects requests missing HX-Request", func(t *testing.T) {
		fc := newTestFeedController(feed.Preferences{UserUUID: ctxUser.UUID, ShowEntries: feed.EntryVisibilityUnread}, testUnreadMetadata())

		form := url.Values{"urlPath": {"/feeds"}, "search": {""}}
		r := newFeedPostRequest(t, "/feeds/entries/mark-all-read", ctxUser, form)
		r.Header.Del("HX-Request")
		w := httptest.NewRecorder()

		fc.handleHxEntryMetadataMarkAllAsRead()(w, r)

		assertRejectsNonHxRequest(t, w)
	})
}

func TestHandleHxEntryMetadataMarkAllAsReadByCategory(t *testing.T) {
	ctxUser := testCtxUser
	entry := testEntry

	t.Run("success, list is re-rendered for the category view", func(t *testing.T) {
		fc := newTestFeedController(feed.Preferences{UserUUID: ctxUser.UUID, ShowEntries: feed.EntryVisibilityUnread}, testUnreadMetadata())

		form := url.Values{"urlPath": {"/feeds/categories/" + testCategory.Slug}, "search": {""}}
		r := newFeedSlugPostRequest(t, "/feeds/categories/"+testCategory.Slug+"/entries/mark-all-read", testCategory.Slug, ctxUser, form)
		w := httptest.NewRecorder()

		fc.handleHxEntryMetadataMarkAllAsReadByCategory()(w, r)

		if w.Code != http.StatusOK {
			t.Fatalf("want status 200, got %d, body:\n%s", w.Code, w.Body.String())
		}

		body := w.Body.String()

		if strings.Contains(body, `id="feed-entry-`+entry.UID+`"`) {
			t.Errorf("want the now-read entry dropped from the Unread-only view, got:\n%s", body)
		}
		if !strings.Contains(body, `<ol id="entry-list"`) {
			t.Errorf("want the entry list re-rendered, got:\n%s", body)
		}
	})

	t.Run("error falls back to a full reload", func(t *testing.T) {
		fc := newTestFeedController(feed.Preferences{UserUUID: ctxUser.UUID, ShowEntries: feed.EntryVisibilityAll}, nil)

		form := url.Values{"urlPath": {"/feeds/categories/does-not-exist"}, "search": {""}}
		r := newFeedSlugPostRequest(t, "/feeds/categories/does-not-exist/entries/mark-all-read", "does-not-exist", ctxUser, form)
		w := httptest.NewRecorder()

		fc.handleHxEntryMetadataMarkAllAsReadByCategory()(w, r)

		assertHXRedirectOnError(t, w, "/feeds")
	})

	t.Run("rejects requests missing HX-Request", func(t *testing.T) {
		fc := newTestFeedController(feed.Preferences{UserUUID: ctxUser.UUID, ShowEntries: feed.EntryVisibilityUnread}, testUnreadMetadata())

		form := url.Values{"urlPath": {"/feeds/categories/" + testCategory.Slug}, "search": {""}}
		r := newFeedSlugPostRequest(t, "/feeds/categories/"+testCategory.Slug+"/entries/mark-all-read", testCategory.Slug, ctxUser, form)
		r.Header.Del("HX-Request")
		w := httptest.NewRecorder()

		fc.handleHxEntryMetadataMarkAllAsReadByCategory()(w, r)

		assertRejectsNonHxRequest(t, w)
	})
}

func TestHandleHxEntryMetadataMarkAllAsReadByFeed(t *testing.T) {
	ctxUser := testCtxUser
	entry := testEntry

	t.Run("success, list is re-rendered for the subscription view", func(t *testing.T) {
		fc := newTestFeedController(feed.Preferences{UserUUID: ctxUser.UUID, ShowEntries: feed.EntryVisibilityUnread}, testUnreadMetadata())

		form := url.Values{"urlPath": {"/feeds/subscriptions/" + testFeed.Slug}, "search": {""}}
		r := newFeedSlugPostRequest(t, "/feeds/subscriptions/"+testFeed.Slug+"/entries/mark-all-read", testFeed.Slug, ctxUser, form)
		w := httptest.NewRecorder()

		fc.handleHxEntryMetadataMarkAllAsReadByFeed()(w, r)

		if w.Code != http.StatusOK {
			t.Fatalf("want status 200, got %d, body:\n%s", w.Code, w.Body.String())
		}

		body := w.Body.String()

		if strings.Contains(body, `id="feed-entry-`+entry.UID+`"`) {
			t.Errorf("want the now-read entry dropped from the Unread-only view, got:\n%s", body)
		}
		if !strings.Contains(body, `<ol id="entry-list"`) {
			t.Errorf("want the entry list re-rendered, got:\n%s", body)
		}
	})

	t.Run("error falls back to a full reload", func(t *testing.T) {
		fc := newTestFeedController(feed.Preferences{UserUUID: ctxUser.UUID, ShowEntries: feed.EntryVisibilityAll}, nil)

		form := url.Values{"urlPath": {"/feeds/subscriptions/does-not-exist"}, "search": {""}}
		r := newFeedSlugPostRequest(t, "/feeds/subscriptions/does-not-exist/entries/mark-all-read", "does-not-exist", ctxUser, form)
		w := httptest.NewRecorder()

		fc.handleHxEntryMetadataMarkAllAsReadByFeed()(w, r)

		assertHXRedirectOnError(t, w, "/feeds")
	})

	t.Run("rejects requests missing HX-Request", func(t *testing.T) {
		fc := newTestFeedController(feed.Preferences{UserUUID: ctxUser.UUID, ShowEntries: feed.EntryVisibilityUnread}, testUnreadMetadata())

		form := url.Values{"urlPath": {"/feeds/subscriptions/" + testFeed.Slug}, "search": {""}}
		r := newFeedSlugPostRequest(t, "/feeds/subscriptions/"+testFeed.Slug+"/entries/mark-all-read", testFeed.Slug, ctxUser, form)
		r.Header.Del("HX-Request")
		w := httptest.NewRecorder()

		fc.handleHxEntryMetadataMarkAllAsReadByFeed()(w, r)

		assertRejectsNonHxRequest(t, w)
	})
}

// newTestFeedControllerForCategoryEdit wires a feedController against the
// given category (needs a real UUID, unlike testCategory's "category-1",
// to satisfy CategoryByUUID/UpdateCategory's UUID format validation), for
// exercising the category rename handlers.
func newTestFeedControllerForCategoryEdit(category feed.Category) feedController {
	feedRepo := &feed.FakeRepository{
		Categories: []feed.Category{category},
	}

	queryingRepo := &feedquerying.FakeRepository{
		Categories: []feed.Category{category},
	}

	return feedController{
		feedService:              feed.NewService(feedRepo, nil),
		queryingService:          feedquerying.NewService(queryingRepo),
		feedSubscriptionListView: view.New("feed/subscription_list.gohtml"),
		feedCategoryEditView:     view.New("feed/category_edit.gohtml"),
	}
}

// newCategoryEditViewRequest builds a GET request against
// /feeds/categories/{uuid}/edit.
func newCategoryEditViewRequest(t *testing.T, ctxUser user.User, categoryUUID string, hxRequest bool) *http.Request {
	t.Helper()

	r := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/feeds/categories/"+categoryUUID+"/edit", nil)
	if hxRequest {
		r.Header.Set("HX-Request", "true")
	}

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("uuid", categoryUUID)

	ctx := httpcontext.WithUser(r.Context(), ctxUser)
	ctx = context.WithValue(ctx, chi.RouteCtxKey, rctx)

	return r.WithContext(ctx)
}

// newCategoryEditPostRequest builds a POST request against
// /feeds/categories/{uuid}/edit.
func newCategoryEditPostRequest(t *testing.T, ctxUser user.User, categoryUUID string, form url.Values, hxRequest bool) *http.Request {
	t.Helper()

	r := httptest.NewRequestWithContext(t.Context(), http.MethodPost, "/feeds/categories/"+categoryUUID+"/edit", strings.NewReader(form.Encode()))
	r.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	if hxRequest {
		r.Header.Set("HX-Request", "true")
	}

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("uuid", categoryUUID)

	ctx := httpcontext.WithUser(r.Context(), ctxUser)
	ctx = context.WithValue(ctx, chi.RouteCtxKey, rctx)

	return r.WithContext(ctx)
}

func TestHandleFeedCategoryEditView(t *testing.T) {
	fake := faker.New()
	ctxUser := testCtxUser

	t.Run("plain browser request renders the full page", func(t *testing.T) {
		category := feed.Category{UUID: fake.UUID().V4(), UserUUID: ctxUser.UUID, Name: fake.Lorem().Text(10)}
		fc := newTestFeedControllerForCategoryEdit(category)
		r := newCategoryEditViewRequest(t, ctxUser, category.UUID, false)
		w := httptest.NewRecorder()

		fc.handleFeedCategoryEditView()(w, r)

		if w.Code != http.StatusOK {
			t.Fatalf("want status 200, got %d, body:\n%s", w.Code, w.Body.String())
		}

		body := w.Body.String()
		if !strings.Contains(body, "<!DOCTYPE html>") {
			t.Errorf("want a full page (with layout), got:\n%s", body)
		}
		if !strings.Contains(body, `value="`+category.Name+`"`) {
			t.Errorf("want the category name pre-filled, got:\n%s", body)
		}
		if strings.Contains(body, "hx-post") {
			t.Errorf("want a plain form with no htmx attributes on the full page, got:\n%s", body)
		}
	})

	t.Run("htmx request renders only the form fragment", func(t *testing.T) {
		category := feed.Category{UUID: fake.UUID().V4(), UserUUID: ctxUser.UUID, Name: fake.Lorem().Text(10)}
		fc := newTestFeedControllerForCategoryEdit(category)
		r := newCategoryEditViewRequest(t, ctxUser, category.UUID, true)
		w := httptest.NewRecorder()

		fc.handleFeedCategoryEditView()(w, r)

		if w.Code != http.StatusOK {
			t.Fatalf("want status 200, got %d, body:\n%s", w.Code, w.Body.String())
		}

		body := w.Body.String()
		if strings.Contains(body, "<!DOCTYPE html>") {
			t.Errorf("want a fragment with no layout, got:\n%s", body)
		}
		if !strings.Contains(body, `value="`+category.Name+`"`) {
			t.Errorf("want the category name pre-filled, got:\n%s", body)
		}
		if !strings.Contains(body, `hx-post="/feeds/categories/`) {
			t.Errorf("want the modal fragment's form to be htmx-enhanced, got:\n%s", body)
		}
	})

	t.Run("unknown category, htmx request uses HX-Redirect", func(t *testing.T) {
		category := feed.Category{UUID: fake.UUID().V4(), UserUUID: ctxUser.UUID, Name: fake.Lorem().Text(10)}
		fc := newTestFeedControllerForCategoryEdit(category)
		unknownUUID := fake.UUID().V4()
		r := newCategoryEditViewRequest(t, ctxUser, unknownUUID, true)
		w := httptest.NewRecorder()

		fc.handleFeedCategoryEditView()(w, r)

		assertHXRedirectOnError(t, w, "/feeds/categories/"+unknownUUID+"/edit")
	})
}

func TestHandleFeedCategoryEdit(t *testing.T) {
	fake := faker.New()
	ctxUser := testCtxUser

	t.Run("plain browser request renames and redirects", func(t *testing.T) {
		category := feed.Category{UUID: fake.UUID().V4(), UserUUID: ctxUser.UUID, Name: fake.Lorem().Text(10)}
		fc := newTestFeedControllerForCategoryEdit(category)
		form := url.Values{"name": {fake.Lorem().Text(10)}}
		r := newCategoryEditPostRequest(t, ctxUser, category.UUID, form, false)
		w := httptest.NewRecorder()

		fc.handleFeedCategoryEdit()(w, r)

		if w.Code != http.StatusSeeOther {
			t.Fatalf("want status 303, got %d, body:\n%s", w.Code, w.Body.String())
		}
		if got := w.Header().Get("Location"); got != "/feeds/subscriptions" {
			t.Errorf("want redirect to /feeds/subscriptions, got %q", got)
		}
	})

	t.Run("htmx request retargets the response into the category's heading and closes the modal", func(t *testing.T) {
		category := feed.Category{UUID: fake.UUID().V4(), UserUUID: ctxUser.UUID, Name: fake.Lorem().Text(10)}
		fc := newTestFeedControllerForCategoryEdit(category)
		newName := fake.Lorem().Text(10)
		form := url.Values{"name": {newName}}
		r := newCategoryEditPostRequest(t, ctxUser, category.UUID, form, true)
		w := httptest.NewRecorder()

		fc.handleFeedCategoryEdit()(w, r)

		if w.Code != http.StatusOK {
			t.Fatalf("want status 200, got %d, body:\n%s", w.Code, w.Body.String())
		}

		if got := w.Header().Get("HX-Retarget"); got != "#category-"+category.UUID {
			t.Errorf("want HX-Retarget to the category's heading, got %q", got)
		}
		if got := w.Header().Get("HX-Reswap"); got != "outerHTML" {
			t.Errorf("want HX-Reswap outerHTML, got %q", got)
		}
		if got := w.Header().Get("HX-Trigger"); got != "modal:close" {
			t.Errorf("want HX-Trigger modal:close, got %q", got)
		}

		body := w.Body.String()
		if !strings.Contains(body, `id="category-`+category.UUID+`"`) {
			t.Errorf("want the category's heading rendered, got:\n%s", body)
		}
		if !strings.Contains(body, newName) {
			t.Errorf("want the new category name rendered, got:\n%s", body)
		}
	})

	t.Run("unknown category, htmx request uses HX-Redirect", func(t *testing.T) {
		category := feed.Category{UUID: fake.UUID().V4(), UserUUID: ctxUser.UUID, Name: fake.Lorem().Text(10)}
		fc := newTestFeedControllerForCategoryEdit(category)
		unknownUUID := fake.UUID().V4()
		form := url.Values{"name": {fake.Lorem().Text(10)}}
		r := newCategoryEditPostRequest(t, ctxUser, unknownUUID, form, true)
		w := httptest.NewRecorder()

		fc.handleFeedCategoryEdit()(w, r)

		assertHXRedirectOnError(t, w, "/feeds/categories/"+unknownUUID+"/edit")
	})
}
