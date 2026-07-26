// Copyright VirtualTam 2022, 2026
// SPDX-License-Identifier: MIT

package controller

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/jaswdr/faker/v2"

	"github.com/virtualtam/sparklemuffin/internal/http/www/httpcontext"
	"github.com/virtualtam/sparklemuffin/internal/http/www/view"
	"github.com/virtualtam/sparklemuffin/pkg/session"
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

func TestHandleUserAddView(t *testing.T) {
	ac := adminController{
		adminUserAddView: view.New("admin/user_add.gohtml"),
	}

	r := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/admin/users/add", nil)
	w := httptest.NewRecorder()

	ac.handleUserAddView()(w, r)

	if w.Code != http.StatusOK {
		t.Fatalf("want status 200, got %d, body:\n%s", w.Code, w.Body.String())
	}

	body := w.Body.String()
	if !strings.Contains(body, `id="password" name="password" minlength="8"`) {
		t.Errorf("want the password field to enforce the minimum password length, got:\n%s", body)
	}
	if !strings.Contains(body, "Must be at least 8 characters long.") {
		t.Errorf("want a hint about the minimum password length, got:\n%s", body)
	}
}

func TestHandleUserEditView(t *testing.T) {
	fake := faker.New()
	targetUser := user.User{
		UUID:  fake.UUID().V4(),
		Email: fake.Internet().Email(),
	}
	ac := adminController{
		userService:       user.NewService(&user.FakeRepository{Users: []user.User{targetUser}}),
		adminUserEditView: view.New("admin/user_edit.gohtml"),
	}

	r := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/admin/users/"+targetUser.UUID, nil)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("uuid", targetUser.UUID)
	r = r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, rctx))
	w := httptest.NewRecorder()

	ac.handleUserEditView()(w, r)

	if w.Code != http.StatusOK {
		t.Fatalf("want status 200, got %d, body:\n%s", w.Code, w.Body.String())
	}

	body := w.Body.String()
	if !strings.Contains(body, `id="password" name="password" minlength="8"`) {
		t.Errorf("want the password field to enforce the minimum password length, got:\n%s", body)
	}
	if !strings.Contains(body, "Must be at least 8 characters long.") {
		t.Errorf("want a hint about the minimum password length, got:\n%s", body)
	}
}

func TestHandleUserAdd(t *testing.T) {
	fake := faker.New()
	ctxUser := user.User{UUID: fake.UUID().V4(), IsAdmin: true}

	t.Run("duplicate email flashes a user-friendly message", func(t *testing.T) {
		userRepo := &user.FakeRepository{
			Users: []user.User{
				{Email: "existing@domain.tld"},
			},
		}
		ac := adminController{
			userService: user.NewService(userRepo),
		}

		form := url.Values{}
		form.Set("email", "existing@domain.tld")
		form.Set("nick_name", "newuser")
		form.Set("display_name", "New User")
		form.Set("password", "new-password1234")

		r := httptest.NewRequestWithContext(t.Context(), http.MethodPost, "/admin/users", strings.NewReader(form.Encode()))
		r.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		r = r.WithContext(httpcontext.WithUser(r.Context(), ctxUser))
		w := httptest.NewRecorder()

		ac.handleUserAdd()(w, r)

		if w.Code != http.StatusSeeOther {
			t.Fatalf("want status %d, got %d, body:\n%s", http.StatusSeeOther, w.Code, w.Body.String())
		}
		if got := decodedFlashMessage(t, w); !strings.Contains(got, "This email address is already registered.") {
			t.Errorf("want a user-friendly duplicate-email message, got %q", got)
		}
		if strings.Contains(decodedFlashMessage(t, w), "user:") {
			t.Errorf("want no raw domain error leaked, got %q", decodedFlashMessage(t, w))
		}
	})
}

func TestHandleUserEdit(t *testing.T) {
	fake := faker.New()
	ctxUser := user.User{UUID: fake.UUID().V4(), IsAdmin: true}

	t.Run("editing a user revokes their existing sessions", func(t *testing.T) {
		userRepo := &user.FakeRepository{}
		userService := user.NewService(userRepo)

		targetUser := user.User{
			UUID:        fake.UUID().V4(),
			Email:       "target@example.com",
			NickName:    "target1",
			DisplayName: "Target One",
			Password:    "original-password1234",
		}
		if err := userService.Add(t.Context(), targetUser); err != nil {
			t.Fatalf("failed to seed user: %q", err)
		}
		seededUser, err := userService.ByNickName(t.Context(), targetUser.NickName)
		if err != nil {
			t.Fatalf("failed to retrieve seeded user: %q", err)
		}

		sessionRepo := &session.FakeRepository{
			Sessions: []session.Session{
				{UserUUID: seededUser.UUID, RememberTokenHash: "hash-1"},
				{UserUUID: "another-user", RememberTokenHash: "hash-2"},
			},
		}
		sessionService, err := session.NewService(sessionRepo, "hmac-key")
		if err != nil {
			t.Fatal(err)
		}

		ac := adminController{
			userService:    userService,
			sessionService: sessionService,
		}

		form := url.Values{}
		form.Set("email", seededUser.Email)
		form.Set("nick_name", seededUser.NickName)
		form.Set("display_name", seededUser.DisplayName)
		form.Set("password", "admin-reset-password1234")

		r := httptest.NewRequestWithContext(t.Context(), http.MethodPost, "/admin/users/"+seededUser.UUID, strings.NewReader(form.Encode()))
		r.Header.Set("Content-Type", "application/x-www-form-urlencoded")

		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("uuid", seededUser.UUID)
		ctx := context.WithValue(r.Context(), chi.RouteCtxKey, rctx)
		ctx = httpcontext.WithUser(ctx, ctxUser)
		r = r.WithContext(ctx)
		w := httptest.NewRecorder()

		ac.handleUserEdit()(w, r)

		if w.Code != http.StatusSeeOther {
			t.Fatalf("want status %d, got %d, body:\n%s", http.StatusSeeOther, w.Code, w.Body.String())
		}

		if len(sessionRepo.Sessions) != 1 {
			t.Fatalf("want 1 remaining session, got %d", len(sessionRepo.Sessions))
		}
		if sessionRepo.Sessions[0].UserUUID != "another-user" {
			t.Errorf("want the other user's session to remain, got session for %q", sessionRepo.Sessions[0].UserUUID)
		}

		if got := decodedFlashLevel(t, w); got != "success" {
			t.Errorf("want a success flash, got %q", got)
		}
	})

	t.Run("session revocation failure still redirects but flashes a warning instead of success", func(t *testing.T) {
		userRepo := &user.FakeRepository{}
		userService := user.NewService(userRepo)

		targetUser := user.User{
			UUID:        fake.UUID().V4(),
			Email:       "target2@example.com",
			NickName:    "target2",
			DisplayName: "Target Two",
			Password:    "original-password1234",
		}
		if err := userService.Add(t.Context(), targetUser); err != nil {
			t.Fatalf("failed to seed user: %q", err)
		}
		seededUser, err := userService.ByNickName(t.Context(), targetUser.NickName)
		if err != nil {
			t.Fatalf("failed to retrieve seeded user: %q", err)
		}

		sessionRepo := &session.FakeRepository{
			SessionDeleteByUserUUIDErr: errors.New("connection reset"),
		}
		sessionService, err := session.NewService(sessionRepo, "hmac-key")
		if err != nil {
			t.Fatal(err)
		}

		ac := adminController{
			userService:    userService,
			sessionService: sessionService,
		}

		form := url.Values{}
		form.Set("email", seededUser.Email)
		form.Set("nick_name", seededUser.NickName)
		form.Set("display_name", seededUser.DisplayName)
		form.Set("password", "admin-reset-password1234")

		r := httptest.NewRequestWithContext(t.Context(), http.MethodPost, "/admin/users/"+seededUser.UUID, strings.NewReader(form.Encode()))
		r.Header.Set("Content-Type", "application/x-www-form-urlencoded")

		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("uuid", seededUser.UUID)
		ctx := context.WithValue(r.Context(), chi.RouteCtxKey, rctx)
		ctx = httpcontext.WithUser(ctx, ctxUser)
		r = r.WithContext(ctx)
		w := httptest.NewRecorder()

		ac.handleUserEdit()(w, r)

		if w.Code != http.StatusSeeOther {
			t.Fatalf("want status %d, got %d, body:\n%s", http.StatusSeeOther, w.Code, w.Body.String())
		}

		if got := decodedFlashLevel(t, w); got != "warning" {
			t.Errorf("want a warning flash when session revocation fails, got %q", got)
		}
	})

	t.Run("duplicate email flashes a user-friendly message", func(t *testing.T) {
		targetUser := user.User{UUID: fake.UUID().V4(), Email: "target3@example.com"}
		userRepo := &user.FakeRepository{
			Users: []user.User{
				targetUser,
				{UUID: fake.UUID().V4(), Email: "other@example.com"},
			},
		}
		ac := adminController{
			userService: user.NewService(userRepo),
		}

		form := url.Values{}
		form.Set("email", "other@example.com")
		form.Set("nick_name", "target3")
		form.Set("display_name", "Target Three")
		form.Set("password", "admin-reset-password1234")

		r := httptest.NewRequestWithContext(t.Context(), http.MethodPost, "/admin/users/"+targetUser.UUID, strings.NewReader(form.Encode()))
		r.Header.Set("Content-Type", "application/x-www-form-urlencoded")

		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("uuid", targetUser.UUID)
		ctx := context.WithValue(r.Context(), chi.RouteCtxKey, rctx)
		ctx = httpcontext.WithUser(ctx, ctxUser)
		r = r.WithContext(ctx)
		w := httptest.NewRecorder()

		ac.handleUserEdit()(w, r)

		if w.Code != http.StatusSeeOther {
			t.Fatalf("want status %d, got %d, body:\n%s", http.StatusSeeOther, w.Code, w.Body.String())
		}
		if got := decodedFlashMessage(t, w); !strings.Contains(got, "This email address is already registered.") {
			t.Errorf("want a user-friendly duplicate-email message, got %q", got)
		}
	})
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
