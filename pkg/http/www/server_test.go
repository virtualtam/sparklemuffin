package www

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/virtualtam/yawbe/pkg/session"
	"github.com/virtualtam/yawbe/pkg/user"
)

func TestServerAuthenticatedUser(t *testing.T) {
	cases := []struct {
		tname      string
		user       *user.User
		wantCalls  int
		wantStatus uint
	}{
		{
			tname:      "anonymous",
			wantCalls:  0,
			wantStatus: http.StatusNotFound,
		},
		{
			tname:      "authenticated, not admin",
			user:       &user.User{},
			wantCalls:  1,
			wantStatus: http.StatusOK,
		},
		{
			tname:      "authenticated, admin",
			user:       &user.User{IsAdmin: true},
			wantCalls:  1,
			wantStatus: http.StatusOK,
		},
	}

	s := NewServer(nil, nil, nil, nil, nil)

	for _, tc := range cases {
		t.Run(tc.tname, func(t *testing.T) {
			var gotCalls int

			handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				gotCalls++
			})

			handler = s.authenticatedUser(handler)

			w := httptest.NewRecorder()

			r := httptest.NewRequest(http.MethodGet, "/", nil)
			if tc.user != nil {
				ctx := withUser(r.Context(), *tc.user)
				r = r.WithContext(ctx)
			}

			handler(w, r)

			if uint(w.Code) != tc.wantStatus {
				t.Errorf("want response status %d, got %d", tc.wantStatus, w.Code)
			}
			if gotCalls != tc.wantCalls {
				t.Errorf("want %d handler calls, got %d", tc.wantCalls, gotCalls)
			}
		})
	}
}

func TestServerAdminUser(t *testing.T) {
	cases := []struct {
		tname      string
		user       *user.User
		wantCalls  int
		wantStatus uint
	}{
		{
			tname:      "anonymous",
			wantCalls:  0,
			wantStatus: http.StatusNotFound,
		},
		{
			tname:      "authenticated, not admin",
			user:       &user.User{},
			wantCalls:  0,
			wantStatus: http.StatusUnauthorized,
		},
		{
			tname:      "authenticated, admin",
			user:       &user.User{IsAdmin: true},
			wantCalls:  1,
			wantStatus: http.StatusOK,
		},
	}

	s := NewServer(nil, nil, nil, nil, nil)

	for _, tc := range cases {
		t.Run(tc.tname, func(t *testing.T) {
			var gotCalls int

			handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				gotCalls++
			})

			handler = s.adminUser(handler)

			w := httptest.NewRecorder()

			r := httptest.NewRequest(http.MethodGet, "/", nil)
			if tc.user != nil {
				ctx := withUser(r.Context(), *tc.user)
				r = r.WithContext(ctx)
			}

			handler(w, r)

			if uint(w.Code) != tc.wantStatus {
				t.Errorf("want response status %d, got %d", tc.wantStatus, w.Code)
			}
			if gotCalls != tc.wantCalls {
				t.Errorf("want %d handler calls, got %d", tc.wantCalls, gotCalls)
			}
		})
	}
}

func TestServerStaticCacheControl(t *testing.T) {
	want := "max-age=2592000"

	s := NewServer(nil, nil, nil, nil, nil)
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})
	handler = s.staticCacheControl(handler)

	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/", nil)

	handler(w, r)
	got := w.Header().Get("Cache-Control")

	if got != want {
		t.Errorf("want Cache-Control %q, got %q", want, got)
	}
}

func TestServerRememberUser(t *testing.T) {
	cases := []struct {
		tname               string
		repositorySessions  []session.Session
		repositoryUsers     []user.User
		rememberTokenCookie *http.Cookie
		wantUser            *user.User
	}{
		{
			tname: "no remember token cookie",
		},
		{
			tname: "remember token cookie set, no corresponding user",
			rememberTokenCookie: &http.Cookie{
				Name:     UserRememberTokenCookieName,
				Value:    "tdk_BrK5adfbUapWUIeQO1VPMkGCtaQFjvF4A0KHy2g=",
				HttpOnly: true,
			},
		},
		{
			tname: "remember token cookie set, corresponding user found",
			repositorySessions: []session.Session{
				{
					UserUUID:          "9c9903c3-d583-4d42-9687-dccdfc77fc3a",
					RememberTokenHash: "W3o3hteHwgT5EGSxhpyotYHNtBhEYlzfkVxViAglBuk=",
				},
			},
			repositoryUsers: []user.User{
				{
					UUID:  "9c9903c3-d583-4d42-9687-dccdfc77fc3a",
					Email: "cookie@domain.tld",
				},
			},
			rememberTokenCookie: &http.Cookie{
				Name:     UserRememberTokenCookieName,
				Value:    "tdk_BrK5adfbUapWUIeQO1VPMkGCtaQFjvF4A0KHy2g=",
				HttpOnly: true,
			},
			wantUser: &user.User{
				Email: "cookie@domain.tld",
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.tname, func(t *testing.T) {
			sessionRepository := &session.FakeRepository{
				Sessions: tc.repositorySessions,
			}
			sessionService := session.NewService(sessionRepository, "ugotcookies")

			userRepository := &user.FakeRepository{
				Users: tc.repositoryUsers,
			}
			userService := user.NewService(userRepository)

			s := NewServer(nil, nil, nil, sessionService, userService)

			var gotContext context.Context
			handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				gotContext = r.Context()
			})
			handler = s.rememberUser(handler)

			w := httptest.NewRecorder()
			r := httptest.NewRequest(http.MethodGet, "/", nil)
			if tc.rememberTokenCookie != nil {
				r.AddCookie(tc.rememberTokenCookie)
			}

			handler(w, r)

			if tc.wantUser == nil {
				return
			}

			got := userValue(gotContext)

			if got.Email != tc.wantUser.Email {
				t.Errorf("want user email %q, got %q", tc.wantUser.Email, got.Email)
			}
		})
	}
}
