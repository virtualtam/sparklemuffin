// Copyright VirtualTam 2022, 2026
// SPDX-License-Identifier: MIT

package controller

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/jaswdr/faker/v2"

	"github.com/virtualtam/sparklemuffin/internal/http/www/httpcontext"
	"github.com/virtualtam/sparklemuffin/internal/http/www/view"
	"github.com/virtualtam/sparklemuffin/pkg/user"
)

// newTestAdminControllerForUserDelete wires an adminController against the
// given user, for exercising the user delete handlers.
func newTestAdminControllerForUserDelete(u user.User) adminController {
	repo := &user.FakeRepository{Users: []user.User{u}}

	return adminController{
		userService:         user.NewService(repo),
		adminUserDeleteView: view.New("admin/user_delete.gohtml"),
	}
}

// newUserDeleteViewRequest builds a GET request against
// /admin/users/{uuid}/delete.
func newUserDeleteViewRequest(t *testing.T, ctxUser user.User, userUUID string, hxRequest bool) *http.Request {
	t.Helper()

	r := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/admin/users/"+userUUID+"/delete", nil)
	if hxRequest {
		r.Header.Set("HX-Request", "true")
	}

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("uuid", userUUID)

	ctx := httpcontext.WithUser(r.Context(), ctxUser)
	ctx = context.WithValue(ctx, chi.RouteCtxKey, rctx)

	return r.WithContext(ctx)
}

// newUserDeletePostRequest builds a POST request against
// /admin/users/{uuid}/delete.
func newUserDeletePostRequest(t *testing.T, ctxUser user.User, userUUID string, hxRequest bool) *http.Request {
	t.Helper()

	r := httptest.NewRequestWithContext(t.Context(), http.MethodPost, "/admin/users/"+userUUID+"/delete", strings.NewReader(""))
	r.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	if hxRequest {
		r.Header.Set("HX-Request", "true")
	}

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("uuid", userUUID)

	ctx := httpcontext.WithUser(r.Context(), ctxUser)
	ctx = context.WithValue(ctx, chi.RouteCtxKey, rctx)

	return r.WithContext(ctx)
}

func TestHandleUserDeleteView(t *testing.T) {
	fake := faker.New()
	ctxUser := user.User{UUID: fake.UUID().V4(), IsAdmin: true}

	newFixture := func() user.User {
		return user.User{
			UUID:  fake.UUID().V4(),
			Email: fake.Internet().Email(),
		}
	}

	t.Run("plain browser request renders the full page", func(t *testing.T) {
		u := newFixture()
		ac := newTestAdminControllerForUserDelete(u)
		r := newUserDeleteViewRequest(t, ctxUser, u.UUID, false)
		w := httptest.NewRecorder()

		ac.handleUserDeleteView()(w, r)

		if w.Code != http.StatusOK {
			t.Fatalf("want status 200, got %d, body:\n%s", w.Code, w.Body.String())
		}

		body := w.Body.String()
		if !strings.Contains(body, "<!DOCTYPE html>") {
			t.Errorf("want a full page (with layout), got:\n%s", body)
		}
		if !strings.Contains(body, u.Email) {
			t.Errorf("want the user's email rendered, got:\n%s", body)
		}
		if strings.Contains(body, "hx-post") {
			t.Errorf("want a plain form with no htmx attributes on the full page, got:\n%s", body)
		}
	})

	t.Run("htmx request renders only the form fragment", func(t *testing.T) {
		u := newFixture()
		ac := newTestAdminControllerForUserDelete(u)
		r := newUserDeleteViewRequest(t, ctxUser, u.UUID, true)
		w := httptest.NewRecorder()

		ac.handleUserDeleteView()(w, r)

		if w.Code != http.StatusOK {
			t.Fatalf("want status 200, got %d, body:\n%s", w.Code, w.Body.String())
		}

		body := w.Body.String()
		if strings.Contains(body, "<!DOCTYPE html>") {
			t.Errorf("want a fragment with no layout, got:\n%s", body)
		}
		if !strings.Contains(body, u.Email) {
			t.Errorf("want the user's email rendered, got:\n%s", body)
		}
		if !strings.Contains(body, `hx-post="/admin/users/`) {
			t.Errorf("want the modal fragment's form to be htmx-enhanced, got:\n%s", body)
		}
	})

	t.Run("unknown user, htmx request uses HX-Redirect", func(t *testing.T) {
		u := newFixture()
		ac := newTestAdminControllerForUserDelete(u)
		unknownUUID := fake.UUID().V4()
		r := newUserDeleteViewRequest(t, ctxUser, unknownUUID, true)
		w := httptest.NewRecorder()

		ac.handleUserDeleteView()(w, r)

		assertHXRedirectOnError(t, w, "/admin/users/"+unknownUUID+"/delete")
	})
}

func TestHandleUserDelete(t *testing.T) {
	fake := faker.New()
	ctxUser := user.User{UUID: fake.UUID().V4(), IsAdmin: true}

	newFixture := func() user.User {
		return user.User{
			UUID:  fake.UUID().V4(),
			Email: fake.Internet().Email(),
		}
	}

	t.Run("plain browser request deletes and redirects", func(t *testing.T) {
		u := newFixture()
		ac := newTestAdminControllerForUserDelete(u)
		r := newUserDeletePostRequest(t, ctxUser, u.UUID, false)
		w := httptest.NewRecorder()

		ac.handleUserDelete()(w, r)

		if w.Code != http.StatusSeeOther {
			t.Fatalf("want status 303, got %d, body:\n%s", w.Code, w.Body.String())
		}
		if got := w.Header().Get("Location"); got != "/admin/users" {
			t.Errorf("want redirect to /admin/users, got %q", got)
		}
	})

	t.Run("htmx request retargets an empty response into the user's row and closes the modal", func(t *testing.T) {
		u := newFixture()
		ac := newTestAdminControllerForUserDelete(u)
		r := newUserDeletePostRequest(t, ctxUser, u.UUID, true)
		w := httptest.NewRecorder()

		ac.handleUserDelete()(w, r)

		if w.Code != http.StatusOK {
			t.Fatalf("want status 200, got %d, body:\n%s", w.Code, w.Body.String())
		}
		if got := w.Header().Get("HX-Retarget"); got != "#user-row-"+u.UUID {
			t.Errorf("want HX-Retarget to the user's row, got %q", got)
		}
		if got := w.Header().Get("HX-Reswap"); got != "outerHTML" {
			t.Errorf("want HX-Reswap outerHTML, got %q", got)
		}
		if got := w.Header().Get("HX-Trigger"); got != "modal:close" {
			t.Errorf("want HX-Trigger modal:close, got %q", got)
		}
		if w.Body.String() != "" {
			t.Errorf("want an empty response body to remove the row, got:\n%s", w.Body.String())
		}
	})

	t.Run("unknown user, htmx request uses HX-Redirect", func(t *testing.T) {
		u := newFixture()
		ac := newTestAdminControllerForUserDelete(u)
		unknownUUID := fake.UUID().V4()
		r := newUserDeletePostRequest(t, ctxUser, unknownUUID, true)
		w := httptest.NewRecorder()

		ac.handleUserDelete()(w, r)

		assertHXRedirectOnError(t, w, "/admin/users/"+unknownUUID+"/delete")
	})
}
