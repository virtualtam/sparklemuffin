// Copyright VirtualTam 2022, 2026
// SPDX-License-Identifier: MIT

package controller

import (
	"context"
	"encoding/base64"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"

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

// newTestBookmarkControllerForTagEdit wires a bookmarkController against a
// bookmark fake repository seeded with the given bookmarks, for exercising
// the tag rename handlers (which only need bookmarkService).
func newTestBookmarkControllerForTagEdit(bookmarks []bookmark.Bookmark) bookmarkController {
	repo := &bookmark.FakeRepository{Bookmarks: bookmarks}

	return bookmarkController{
		bookmarkService: bookmark.NewService(repo),
		tagListView:     view.New("bookmark/tag_list.gohtml"),
		tagEditView:     view.New("bookmark/tag_edit.gohtml"),
	}
}

// newTagEditViewRequest builds a GET request against
// /bookmarks/tags/{name}/edit, with the tag name chi URL param base64-encoded
// as the real routes expect.
func newTagEditViewRequest(t *testing.T, ctxUser user.User, tagName string, hxRequest bool) *http.Request {
	t.Helper()

	encodedName := base64.URLEncoding.EncodeToString([]byte(tagName))

	r := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/bookmarks/tags/"+encodedName+"/edit", nil)
	if hxRequest {
		r.Header.Set("HX-Request", "true")
	}

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("name", encodedName)

	ctx := httpcontext.WithUser(r.Context(), ctxUser)
	ctx = context.WithValue(ctx, chi.RouteCtxKey, rctx)

	return r.WithContext(ctx)
}

// newTagEditViewRequestRaw builds a GET request against
// /bookmarks/tags/{name}/edit using rawEncodedName verbatim as the chi URL
// param, without base64-encoding it first: used to exercise the invalid-tag
// error path.
func newTagEditViewRequestRaw(t *testing.T, ctxUser user.User, rawEncodedName string, hxRequest bool) *http.Request {
	t.Helper()

	r := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/bookmarks/tags/"+rawEncodedName+"/edit", nil)
	if hxRequest {
		r.Header.Set("HX-Request", "true")
	}

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("name", rawEncodedName)

	ctx := httpcontext.WithUser(r.Context(), ctxUser)
	ctx = context.WithValue(ctx, chi.RouteCtxKey, rctx)

	return r.WithContext(ctx)
}

// newTagEditPostRequest builds a POST request against
// /bookmarks/tags/{name}/edit.
func newTagEditPostRequest(t *testing.T, ctxUser user.User, form url.Values, hxRequest bool) *http.Request {
	t.Helper()

	encodedName := base64.URLEncoding.EncodeToString([]byte("example"))

	r := httptest.NewRequestWithContext(t.Context(), http.MethodPost, "/bookmarks/tags/"+encodedName+"/edit", strings.NewReader(form.Encode()))
	r.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	if hxRequest {
		r.Header.Set("HX-Request", "true")
	}

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("name", encodedName)

	ctx := httpcontext.WithUser(r.Context(), ctxUser)
	ctx = context.WithValue(ctx, chi.RouteCtxKey, rctx)

	return r.WithContext(ctx)
}

func TestHandleTagEditView(t *testing.T) {
	ctxUser := testBookmarkCtxUser

	t.Run("plain browser request renders the full page", func(t *testing.T) {
		bc := newTestBookmarkControllerForTagEdit(nil)
		r := newTagEditViewRequest(t, ctxUser, "example", false)
		w := httptest.NewRecorder()

		bc.handleTagEditView()(w, r)

		if w.Code != http.StatusOK {
			t.Fatalf("want status 200, got %d, body:\n%s", w.Code, w.Body.String())
		}

		body := w.Body.String()
		if !strings.Contains(body, "<!DOCTYPE html>") {
			t.Errorf("want a full page (with layout), got:\n%s", body)
		}
		if !strings.Contains(body, `value="example"`) {
			t.Errorf("want the tag name pre-filled, got:\n%s", body)
		}
		if strings.Contains(body, "hx-post") {
			t.Errorf("want a plain form with no htmx attributes on the full page, got:\n%s", body)
		}
	})

	t.Run("htmx request renders only the form fragment", func(t *testing.T) {
		bc := newTestBookmarkControllerForTagEdit(nil)
		r := newTagEditViewRequest(t, ctxUser, "example", true)
		w := httptest.NewRecorder()

		bc.handleTagEditView()(w, r)

		if w.Code != http.StatusOK {
			t.Fatalf("want status 200, got %d, body:\n%s", w.Code, w.Body.String())
		}

		body := w.Body.String()
		if strings.Contains(body, "<!DOCTYPE html>") {
			t.Errorf("want a fragment with no layout, got:\n%s", body)
		}
		if !strings.Contains(body, `value="example"`) {
			t.Errorf("want the tag name pre-filled, got:\n%s", body)
		}
		if !strings.Contains(body, `hx-post="/bookmarks/tags/`) {
			t.Errorf("want the modal fragment's form to be htmx-enhanced, got:\n%s", body)
		}
	})

	t.Run("invalid tag, htmx request uses HX-Redirect", func(t *testing.T) {
		bc := newTestBookmarkControllerForTagEdit(nil)
		r := newTagEditViewRequestRaw(t, ctxUser, "not-valid-base64!!", true)
		w := httptest.NewRecorder()

		bc.handleTagEditView()(w, r)

		assertHXRedirectOnError(t, w, "/bookmarks/tags/not-valid-base64!!/edit")
	})
}

func TestHandleTagEdit(t *testing.T) {
	ctxUser := testBookmarkCtxUser

	t.Run("plain browser request renames and redirects", func(t *testing.T) {
		bc := newTestBookmarkControllerForTagEdit([]bookmark.Bookmark{testBookmarkEntry})
		form := url.Values{"name": {"renamed"}}
		r := newTagEditPostRequest(t, ctxUser, form, false)
		w := httptest.NewRecorder()

		bc.handleTagEdit()(w, r)

		if w.Code != http.StatusSeeOther {
			t.Fatalf("want status 303, got %d, body:\n%s", w.Code, w.Body.String())
		}
		if got := w.Header().Get("Location"); got != "/bookmarks/tags" {
			t.Errorf("want redirect to /bookmarks/tags, got %q", got)
		}
	})

	t.Run("htmx request retargets the response into the tag's row and closes the modal", func(t *testing.T) {
		bc := newTestBookmarkControllerForTagEdit([]bookmark.Bookmark{testBookmarkEntry})
		form := url.Values{"name": {"renamed"}}
		r := newTagEditPostRequest(t, ctxUser, form, true)
		w := httptest.NewRecorder()

		bc.handleTagEdit()(w, r)

		if w.Code != http.StatusOK {
			t.Fatalf("want status 200, got %d, body:\n%s", w.Code, w.Body.String())
		}

		oldEncodedName := base64.URLEncoding.EncodeToString([]byte("example"))
		newEncodedName := base64.URLEncoding.EncodeToString([]byte("renamed"))

		// Must target the pre-rename row id: it's still what's in the DOM.
		if got := w.Header().Get("HX-Retarget"); got != fmt.Sprintf("[id='tag-row-%s']", oldEncodedName) {
			t.Errorf("want HX-Retarget to the tag's pre-rename row, got %q", got)
		}
		if got := w.Header().Get("HX-Reswap"); got != "outerHTML" {
			t.Errorf("want HX-Reswap outerHTML, got %q", got)
		}
		if got := w.Header().Get("HX-Trigger"); got != "modal:close" {
			t.Errorf("want HX-Trigger modal:close, got %q", got)
		}

		body := w.Body.String()
		if !strings.Contains(body, `id="tag-row-`+newEncodedName+`"`) {
			t.Errorf("want the renamed tag's row rendered, got:\n%s", body)
		}
		if !strings.Contains(body, "renamed") {
			t.Errorf("want the new tag name rendered, got:\n%s", body)
		}
	})

	t.Run("failed to rename tag, htmx request uses HX-Redirect", func(t *testing.T) {
		bc := newTestBookmarkControllerForTagEdit([]bookmark.Bookmark{testBookmarkEntry})
		form := url.Values{"name": {"new name"}} // whitespace is rejected by TagUpdateQuery validation
		r := newTagEditPostRequest(t, ctxUser, form, true)
		w := httptest.NewRecorder()

		bc.handleTagEdit()(w, r)

		assertHXRedirectOnError(t, w, "/bookmarks/tags/"+base64.URLEncoding.EncodeToString([]byte("example"))+"/edit")
	})

	t.Run("renaming to the same name is a no-op", func(t *testing.T) {
		bc := newTestBookmarkControllerForTagEdit([]bookmark.Bookmark{testBookmarkEntry})
		form := url.Values{"name": {"example"}}
		r := newTagEditPostRequest(t, ctxUser, form, true)
		w := httptest.NewRecorder()

		bc.handleTagEdit()(w, r)

		if w.Code != http.StatusOK {
			t.Fatalf("want status 200, got %d, body:\n%s", w.Code, w.Body.String())
		}
		if got := w.Header().Get("HX-Trigger"); got != "modal:close" {
			t.Errorf("want HX-Trigger modal:close, got %q", got)
		}
	})
}
