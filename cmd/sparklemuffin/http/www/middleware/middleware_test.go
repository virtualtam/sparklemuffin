package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/virtualtam/sparklemuffin/cmd/sparklemuffin/http/www/httpcontext"
	"github.com/virtualtam/sparklemuffin/pkg/user"
)

func TestAdminUser(t *testing.T) {
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

	for _, tc := range cases {
		t.Run(tc.tname, func(t *testing.T) {
			var gotCalls int

			handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				gotCalls++
			})

			handler = AdminUser(handler)

			w := httptest.NewRecorder()

			r := httptest.NewRequest(http.MethodGet, "/", nil)
			if tc.user != nil {
				ctx := httpcontext.WithUser(r.Context(), *tc.user)
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

func TestAuthenticatedUser(t *testing.T) {
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

	for _, tc := range cases {
		t.Run(tc.tname, func(t *testing.T) {
			var gotCalls int

			handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				gotCalls++
			})

			handler = AuthenticatedUser(handler)

			w := httptest.NewRecorder()

			r := httptest.NewRequest(http.MethodGet, "/", nil)
			if tc.user != nil {
				ctx := httpcontext.WithUser(r.Context(), *tc.user)
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

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})
	handler = StaticCacheControl(handler)

	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/", nil)

	handler(w, r)
	got := w.Header().Get("Cache-Control")

	if got != want {
		t.Errorf("want Cache-Control %q, got %q", want, got)
	}
}
