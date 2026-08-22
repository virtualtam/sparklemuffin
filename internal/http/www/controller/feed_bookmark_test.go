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
	"github.com/segmentio/ksuid"

	"github.com/virtualtam/sparklemuffin/internal/http/www/httpcontext"
	"github.com/virtualtam/sparklemuffin/internal/http/www/view"
	"github.com/virtualtam/sparklemuffin/pkg/bookmark"
	"github.com/virtualtam/sparklemuffin/pkg/feed"
	feedquerying "github.com/virtualtam/sparklemuffin/pkg/feed/querying"
	"github.com/virtualtam/sparklemuffin/pkg/taxonomy"
	"github.com/virtualtam/sparklemuffin/pkg/user"
)

// newTestFeedBookmarkController wires a feedBookmarkController against fake
// feed and bookmark repositories, seeded with a single subscribed feed entry
// owned by ctxUser.
func newTestFeedBookmarkController(ctxUser user.User, entry feed.Entry, f feed.Feed, subscription feed.Subscription, bookmarks []bookmark.Bookmark) feedBookmarkController {
	feedQueryingRepo := &feedquerying.FakeRepository{
		Entries:       []feed.Entry{entry},
		Feeds:         []feed.Feed{f},
		Subscriptions: []feed.Subscription{subscription},
	}

	bookmarkRepo := &bookmark.FakeRepository{Bookmarks: bookmarks}

	var tags []taxonomy.Tag
	for _, b := range bookmarks {
		for _, name := range b.Tags {
			tags = append(tags, taxonomy.Tag{UserUUID: ctxUser.UUID, Name: name})
		}
	}
	taxonomyRepo := &taxonomy.FakeRepository{Tags: tags}

	return feedBookmarkController{
		feedQueryingService: feedquerying.NewService(feedQueryingRepo),
		bookmarkService:     bookmark.NewService(bookmarkRepo),
		taxonomyService:     taxonomy.NewService(taxonomyRepo),

		entryBookmarkView: view.New("feed/entry_bookmark.gohtml"),
	}
}

// newFeedEntryBookmarkViewRequest builds a GET request against
// /feeds/entries/{uid}/bookmark.
func newFeedEntryBookmarkViewRequest(t *testing.T, ctxUser user.User, entryUID string, hxRequest bool) *http.Request {
	t.Helper()

	r := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/feeds/entries/"+entryUID+"/bookmark", nil)
	if hxRequest {
		r.Header.Set("HX-Request", "true")
	}

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("uid", entryUID)

	ctx := httpcontext.WithUser(r.Context(), ctxUser)
	ctx = context.WithValue(ctx, chi.RouteCtxKey, rctx)

	return r.WithContext(ctx)
}

// newFeedEntryBookmarkPostRequest builds a POST request against
// /feeds/entries/{uid}/bookmark.
func newFeedEntryBookmarkPostRequest(t *testing.T, ctxUser user.User, entryUID string, form url.Values, hxRequest bool) *http.Request {
	t.Helper()

	r := httptest.NewRequestWithContext(t.Context(), http.MethodPost, "/feeds/entries/"+entryUID+"/bookmark", strings.NewReader(form.Encode()))
	r.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	if hxRequest {
		r.Header.Set("HX-Request", "true")
	}

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("uid", entryUID)

	ctx := httpcontext.WithUser(r.Context(), ctxUser)
	ctx = context.WithValue(ctx, chi.RouteCtxKey, rctx)

	return r.WithContext(ctx)
}

func TestHandleFeedEntryBookmarkView(t *testing.T) {
	fake := faker.New()
	ctxUser := user.User{UUID: fake.UUID().V4(), NickName: fake.Internet().User(), DisplayName: fake.Person().Name()}

	f := feed.Feed{UUID: fake.UUID().V4(), Slug: "example-feed", Title: "Example Feed"}
	subscription := feed.Subscription{
		UUID:     fake.UUID().V4(),
		UserUUID: ctxUser.UUID,
		FeedUUID: f.UUID,
		Tags:     []string{"tech"},
	}
	entry := feed.Entry{
		UID:      fake.UUID().V4(),
		FeedUUID: f.UUID,
		URL:      "https://example.com/posts/1",
		Title:    "First Post",
	}

	t.Run("htmx request, no existing bookmark, renders the pre-filled add form fragment", func(t *testing.T) {
		fbc := newTestFeedBookmarkController(ctxUser, entry, f, subscription, nil)
		r := newFeedEntryBookmarkViewRequest(t, ctxUser, entry.UID, true)
		w := httptest.NewRecorder()

		fbc.handleFeedEntryBookmarkView()(w, r)

		if w.Code != http.StatusOK {
			t.Fatalf("want status 200, got %d, body:\n%s", w.Code, w.Body.String())
		}

		body := w.Body.String()
		if strings.Contains(body, "<!DOCTYPE html>") {
			t.Errorf("want a fragment with no layout, got:\n%s", body)
		}
		if !strings.Contains(body, `value="`+entry.URL+`"`) {
			t.Errorf("want the entry URL pre-filled, got:\n%s", body)
		}
		if !strings.Contains(body, `value="`+entry.Title+`"`) {
			t.Errorf("want the entry title pre-filled, got:\n%s", body)
		}
		if !strings.Contains(body, `value="tech"`) {
			t.Errorf("want the subscription tags pre-filled, got:\n%s", body)
		}
		if !strings.Contains(body, `hx-post="/feeds/entries/`+entry.UID+`/bookmark"`) {
			t.Errorf("want the form to post back to the bridge route, got:\n%s", body)
		}
	})

	t.Run("htmx request, existing bookmark for this URL, renders the pre-filled edit form fragment", func(t *testing.T) {
		existing := bookmark.Bookmark{
			UID:         ksuid.New().String(),
			UserUUID:    ctxUser.UUID,
			URL:         entry.URL,
			Title:       "My saved title",
			Description: "My notes",
			Tags:        []string{"saved"},
			Private:     true,
		}
		fbc := newTestFeedBookmarkController(ctxUser, entry, f, subscription, []bookmark.Bookmark{existing})
		r := newFeedEntryBookmarkViewRequest(t, ctxUser, entry.UID, true)
		w := httptest.NewRecorder()

		fbc.handleFeedEntryBookmarkView()(w, r)

		if w.Code != http.StatusOK {
			t.Fatalf("want status 200, got %d, body:\n%s", w.Code, w.Body.String())
		}

		body := w.Body.String()
		if !strings.Contains(body, `value="My saved title"`) {
			t.Errorf("want the existing bookmark's title pre-filled, not the entry's title, got:\n%s", body)
		}
		if !strings.Contains(body, "My notes") {
			t.Errorf("want the existing bookmark's description pre-filled, got:\n%s", body)
		}
		if !strings.Contains(body, `value="saved"`) {
			t.Errorf("want the existing bookmark's tags pre-filled, not the subscription's tags, got:\n%s", body)
		}
		if !strings.Contains(body, `id="private" name="private" checked`) {
			t.Errorf("want the existing bookmark's private flag pre-filled, got:\n%s", body)
		}
		if !strings.Contains(body, "already bookmarked") {
			t.Errorf("want a notice that this entry is already bookmarked, got:\n%s", body)
		}
		if !strings.Contains(body, `hx-post="/feeds/entries/`+entry.UID+`/bookmark"`) {
			t.Errorf("want the form to still post back to the bridge route, got:\n%s", body)
		}
	})

	t.Run("plain browser request, no existing bookmark, renders the full page", func(t *testing.T) {
		fbc := newTestFeedBookmarkController(ctxUser, entry, f, subscription, nil)
		r := newFeedEntryBookmarkViewRequest(t, ctxUser, entry.UID, false)
		w := httptest.NewRecorder()

		fbc.handleFeedEntryBookmarkView()(w, r)

		if w.Code != http.StatusOK {
			t.Fatalf("want status 200, got %d, body:\n%s", w.Code, w.Body.String())
		}

		body := w.Body.String()
		if !strings.Contains(body, "<!DOCTYPE html>") {
			t.Errorf("want a full page (with layout), got:\n%s", body)
		}
		if !strings.Contains(body, `value="`+entry.URL+`"`) {
			t.Errorf("want the entry URL pre-filled, got:\n%s", body)
		}
	})

	t.Run("unknown entry, htmx request uses HX-Redirect", func(t *testing.T) {
		fbc := newTestFeedBookmarkController(ctxUser, entry, f, subscription, nil)
		unknownUID := ksuid.New().String()
		r := newFeedEntryBookmarkViewRequest(t, ctxUser, unknownUID, true)
		w := httptest.NewRecorder()

		fbc.handleFeedEntryBookmarkView()(w, r)

		assertHXRedirectOnError(t, w, "/feeds/entries/"+unknownUID+"/bookmark")
	})

	t.Run("entry belonging to another user's subscription, htmx request uses HX-Redirect", func(t *testing.T) {
		otherUser := user.User{UUID: fake.UUID().V4(), NickName: fake.Internet().User(), DisplayName: fake.Person().Name()}
		fbc := newTestFeedBookmarkController(otherUser, entry, f, subscription, nil)
		r := newFeedEntryBookmarkViewRequest(t, otherUser, entry.UID, true)
		w := httptest.NewRecorder()

		fbc.handleFeedEntryBookmarkView()(w, r)

		assertHXRedirectOnError(t, w, "/feeds/entries/"+entry.UID+"/bookmark")
	})
}

func TestHandleFeedEntryBookmark(t *testing.T) {
	fake := faker.New()
	ctxUser := user.User{UUID: fake.UUID().V4(), NickName: fake.Internet().User(), DisplayName: fake.Person().Name()}

	f := feed.Feed{UUID: fake.UUID().V4(), Slug: "example-feed", Title: "Example Feed"}
	subscription := feed.Subscription{
		UUID:     fake.UUID().V4(),
		UserUUID: ctxUser.UUID,
		FeedUUID: f.UUID,
		Tags:     []string{"tech"},
	}
	entry := feed.Entry{
		UID:      fake.UUID().V4(),
		FeedUUID: f.UUID,
		URL:      "https://example.com/posts/1",
		Title:    "First Post",
	}

	t.Run("htmx request, no existing bookmark, creates it and renders the confirmation fragment", func(t *testing.T) {
		fbc := newTestFeedBookmarkController(ctxUser, entry, f, subscription, nil)
		form := url.Values{
			"url":         {entry.URL},
			"title":       {entry.Title},
			"description": {""},
			"tags":        {"tech"},
		}
		r := newFeedEntryBookmarkPostRequest(t, ctxUser, entry.UID, form, true)
		w := httptest.NewRecorder()

		fbc.handleFeedEntryBookmark()(w, r)

		if w.Code != http.StatusOK {
			t.Fatalf("want status 200, got %d, body:\n%s", w.Code, w.Body.String())
		}

		body := w.Body.String()
		if !strings.Contains(body, entry.Title) {
			t.Errorf("want the confirmation to mention the bookmark's title, got:\n%s", body)
		}
		if !strings.Contains(body, "/bookmarks/") {
			t.Errorf("want a link to the saved bookmark, got:\n%s", body)
		}

		saved, err := fbc.bookmarkService.ByURL(t.Context(), ctxUser.UUID, entry.URL)
		if err != nil {
			t.Fatalf("want the bookmark to have been created, got error: %s", err)
		}
		if saved.Title != entry.Title {
			t.Errorf("want the created bookmark's title %q, got %q", entry.Title, saved.Title)
		}
		if len(saved.Tags) != 1 || saved.Tags[0] != "tech" {
			t.Errorf("want the created bookmark's tags [tech], got %v", saved.Tags)
		}
	})

	t.Run("htmx request, existing bookmark for this URL, updates it instead of duplicating it", func(t *testing.T) {
		existing := bookmark.Bookmark{
			UID:      ksuid.New().String(),
			UserUUID: ctxUser.UUID,
			URL:      entry.URL,
			Title:    "Old title",
			Tags:     []string{"old"},
		}
		fbc := newTestFeedBookmarkController(ctxUser, entry, f, subscription, []bookmark.Bookmark{existing})
		form := url.Values{
			"url":         {entry.URL},
			"title":       {"New title"},
			"description": {"New notes"},
			"tags":        {"new"},
		}
		r := newFeedEntryBookmarkPostRequest(t, ctxUser, entry.UID, form, true)
		w := httptest.NewRecorder()

		fbc.handleFeedEntryBookmark()(w, r)

		if w.Code != http.StatusOK {
			t.Fatalf("want status 200, got %d, body:\n%s", w.Code, w.Body.String())
		}

		body := w.Body.String()
		if !strings.Contains(body, "New title") {
			t.Errorf("want the confirmation to mention the updated title, got:\n%s", body)
		}

		all, err := fbc.bookmarkService.All(t.Context(), ctxUser.UUID)
		if err != nil {
			t.Fatalf("failed to list bookmarks: %s", err)
		}
		if len(all) != 1 {
			t.Fatalf("want exactly 1 bookmark (updated, not duplicated), got %d", len(all))
		}
		if all[0].UID != existing.UID {
			t.Errorf("want the same bookmark UID %q to be reused, got %q", existing.UID, all[0].UID)
		}
		if all[0].Title != "New title" {
			t.Errorf("want the bookmark's title updated to %q, got %q", "New title", all[0].Title)
		}
		if len(all[0].Tags) != 1 || all[0].Tags[0] != "new" {
			t.Errorf("want the bookmark's tags updated to [new], got %v", all[0].Tags)
		}
	})

	t.Run("missing title, htmx request uses HX-Redirect and does not create a bookmark", func(t *testing.T) {
		fbc := newTestFeedBookmarkController(ctxUser, entry, f, subscription, nil)
		form := url.Values{
			"url": {entry.URL}, // no title: fails ValidateForAddition
		}
		r := newFeedEntryBookmarkPostRequest(t, ctxUser, entry.UID, form, true)
		w := httptest.NewRecorder()

		fbc.handleFeedEntryBookmark()(w, r)

		assertHXRedirectOnError(t, w, "/feeds/entries/"+entry.UID+"/bookmark")

		all, err := fbc.bookmarkService.All(t.Context(), ctxUser.UUID)
		if err != nil {
			t.Fatalf("failed to list bookmarks: %s", err)
		}
		if len(all) != 0 {
			t.Errorf("want no bookmark to have been created, got %d", len(all))
		}
	})

	t.Run("missing URL, htmx request uses HX-Redirect and does not create a bookmark", func(t *testing.T) {
		fbc := newTestFeedBookmarkController(ctxUser, entry, f, subscription, nil)
		form := url.Values{
			"title": {entry.Title}, // no url: fails the existing-bookmark check and ValidateForAddition
		}
		r := newFeedEntryBookmarkPostRequest(t, ctxUser, entry.UID, form, true)
		w := httptest.NewRecorder()

		fbc.handleFeedEntryBookmark()(w, r)

		assertHXRedirectOnError(t, w, "/feeds/entries/"+entry.UID+"/bookmark")

		all, err := fbc.bookmarkService.All(t.Context(), ctxUser.UUID)
		if err != nil {
			t.Fatalf("failed to list bookmarks: %s", err)
		}
		if len(all) != 0 {
			t.Errorf("want no bookmark to have been created, got %d", len(all))
		}
	})

	t.Run("plain browser request, no existing bookmark, creates it and redirects to the bookmark list", func(t *testing.T) {
		fbc := newTestFeedBookmarkController(ctxUser, entry, f, subscription, nil)
		form := url.Values{
			"url":         {entry.URL},
			"title":       {entry.Title},
			"description": {""},
			"tags":        {"tech"},
		}
		r := newFeedEntryBookmarkPostRequest(t, ctxUser, entry.UID, form, false)
		w := httptest.NewRecorder()

		fbc.handleFeedEntryBookmark()(w, r)

		if w.Code != http.StatusSeeOther {
			t.Fatalf("want status 303, got %d, body:\n%s", w.Code, w.Body.String())
		}
		if got := w.Header().Get("Location"); got != "/bookmarks" {
			t.Errorf("want redirect to /bookmarks, got %q", got)
		}

		saved, err := fbc.bookmarkService.ByURL(t.Context(), ctxUser.UUID, entry.URL)
		if err != nil {
			t.Fatalf("want the bookmark to have been created, got error: %s", err)
		}
		if saved.Title != entry.Title {
			t.Errorf("want the created bookmark's title %q, got %q", entry.Title, saved.Title)
		}
	})
}
