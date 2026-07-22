// Copyright VirtualTam 2022, 2026
// SPDX-License-Identifier: MIT

package controller

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/virtualtam/sparklemuffin/internal/http/www/httpcontext"
	"github.com/virtualtam/sparklemuffin/internal/http/www/view"
	"github.com/virtualtam/sparklemuffin/pkg/bookmark"
	bookmarkquerying "github.com/virtualtam/sparklemuffin/pkg/bookmark/querying"
	"github.com/virtualtam/sparklemuffin/pkg/user"
)

var (
	testBookmarkCtxUser = user.User{UUID: "user-1", NickName: "alice", DisplayName: "Alice"}
	testBookmarkEntry   = bookmark.Bookmark{
		UID:       "bookmark-1",
		UserUUID:  testBookmarkCtxUser.UUID,
		URL:       "https://example.com/1",
		Title:     "Example Domain",
		Tags:      []string{"example"},
		CreatedAt: time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
	}
)

// newTestBookmarkController wires a bookmarkController against a querying
// fake repository seeded with the given bookmarks, owned by
// testBookmarkCtxUser.
func newTestBookmarkController(bookmarks []bookmark.Bookmark) bookmarkController {
	queryingRepo := &bookmarkquerying.FakeRepository{
		Bookmarks: bookmarks,
		Users:     []user.User{testBookmarkCtxUser},
	}

	return bookmarkController{
		queryingService:  bookmarkquerying.NewService(queryingRepo),
		bookmarkListView: view.New("bookmark/bookmark_list.gohtml"),
	}
}

// newBookmarkListRequest builds a GET request against /bookmarks, optionally
// carrying HX-Request, with the given user set in context.
func newBookmarkListRequest(t *testing.T, ctxUser user.User, rawQuery string, hxRequest bool) *http.Request {
	t.Helper()

	target := "/bookmarks"
	if rawQuery != "" {
		target += "?" + rawQuery
	}

	r := httptest.NewRequestWithContext(t.Context(), http.MethodGet, target, nil)
	if hxRequest {
		r.Header.Set("HX-Request", "true")
	}

	ctx := httpcontext.WithUser(r.Context(), ctxUser)
	return r.WithContext(ctx)
}

func TestHandleBookmarkListView(t *testing.T) {
	ctxUser := testBookmarkCtxUser

	t.Run("plain browser request renders the full page", func(t *testing.T) {
		bc := newTestBookmarkController([]bookmark.Bookmark{testBookmarkEntry})
		r := newBookmarkListRequest(t, ctxUser, "", false)
		w := httptest.NewRecorder()

		bc.handleBookmarkListView()(w, r)

		if w.Code != http.StatusOK {
			t.Fatalf("want status 200, got %d, body:\n%s", w.Code, w.Body.String())
		}

		body := w.Body.String()
		if !strings.Contains(body, "<!DOCTYPE html>") {
			t.Errorf("want a full page (with layout), got:\n%s", body)
		}
		if !strings.Contains(body, `id="bookmark-list-content"`) {
			t.Errorf("want the bookmark list content, got:\n%s", body)
		}
		if !strings.Contains(body, testBookmarkEntry.Title) {
			t.Errorf("want the bookmark title rendered, got:\n%s", body)
		}
	})

	t.Run("htmx request renders only the fragment", func(t *testing.T) {
		bc := newTestBookmarkController([]bookmark.Bookmark{testBookmarkEntry})
		r := newBookmarkListRequest(t, ctxUser, "", true)
		w := httptest.NewRecorder()

		bc.handleBookmarkListView()(w, r)

		if w.Code != http.StatusOK {
			t.Fatalf("want status 200, got %d, body:\n%s", w.Code, w.Body.String())
		}

		body := w.Body.String()
		if strings.Contains(body, "<!DOCTYPE html>") {
			t.Errorf("want a fragment with no layout, got:\n%s", body)
		}
		if !strings.Contains(body, `id="bookmark-list-content"`) {
			t.Errorf("want the bookmark list content, got:\n%s", body)
		}
		if !strings.Contains(body, testBookmarkEntry.Title) {
			t.Errorf("want the bookmark title rendered, got:\n%s", body)
		}
	})

	t.Run("htmx search request filters the fragment", func(t *testing.T) {
		other := bookmark.Bookmark{
			UID:       "bookmark-2",
			UserUUID:  ctxUser.UUID,
			URL:       "https://different-domain.test/2",
			Title:     "Something else entirely",
			CreatedAt: time.Date(2026, 1, 2, 0, 0, 0, 0, time.UTC),
		}
		bc := newTestBookmarkController([]bookmark.Bookmark{testBookmarkEntry, other})
		r := newBookmarkListRequest(t, ctxUser, "search=Example", true)
		w := httptest.NewRecorder()

		bc.handleBookmarkListView()(w, r)

		if w.Code != http.StatusOK {
			t.Fatalf("want status 200, got %d, body:\n%s", w.Code, w.Body.String())
		}

		body := w.Body.String()
		if !strings.Contains(body, testBookmarkEntry.Title) {
			t.Errorf("want the matching bookmark rendered, got:\n%s", body)
		}
		if strings.Contains(body, other.Title) {
			t.Errorf("want the non-matching bookmark excluded, got:\n%s", body)
		}
	})

	t.Run("invalid page number, plain request falls back to a real redirect", func(t *testing.T) {
		bc := newTestBookmarkController([]bookmark.Bookmark{testBookmarkEntry})
		r := newBookmarkListRequest(t, ctxUser, "page=notanumber", false)
		w := httptest.NewRecorder()

		bc.handleBookmarkListView()(w, r)

		if w.Code != http.StatusSeeOther {
			t.Fatalf("want status 303, got %d", w.Code)
		}
		if got := w.Header().Get("Location"); got != "/bookmarks" {
			t.Errorf("want redirect to /bookmarks, got %q", got)
		}
	})

	t.Run("invalid page number, htmx request uses HX-Redirect", func(t *testing.T) {
		bc := newTestBookmarkController([]bookmark.Bookmark{testBookmarkEntry})
		r := newBookmarkListRequest(t, ctxUser, "page=notanumber", true)
		w := httptest.NewRecorder()

		bc.handleBookmarkListView()(w, r)

		assertHXRedirectOnError(t, w, "/bookmarks")
	})

	t.Run("page number out of bounds, htmx request uses HX-Redirect", func(t *testing.T) {
		bc := newTestBookmarkController([]bookmark.Bookmark{testBookmarkEntry})
		r := newBookmarkListRequest(t, ctxUser, "page=99", true)
		w := httptest.NewRecorder()

		bc.handleBookmarkListView()(w, r)

		assertHXRedirectOnError(t, w, "/bookmarks")
	})

	t.Run("failed to retrieve bookmarks, htmx request uses HX-Redirect", func(t *testing.T) {
		// The owning user isn't present in the fixture, so OwnerGetByUUID
		// fails and the handler falls back to its generic retrieval error.
		queryingRepo := &bookmarkquerying.FakeRepository{}
		bc := bookmarkController{
			queryingService:  bookmarkquerying.NewService(queryingRepo),
			bookmarkListView: view.New("bookmark/bookmark_list.gohtml"),
		}
		r := newBookmarkListRequest(t, ctxUser, "", true)
		w := httptest.NewRecorder()

		bc.handleBookmarkListView()(w, r)

		assertHXRedirectOnError(t, w, "/bookmarks")
	})
}
