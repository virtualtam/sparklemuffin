package www

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/virtualtam/yawbe/pkg/user"
)

func TestServerStaticCacheControl(t *testing.T) {
	want := "max-age=2592000"

	s := NewServer(nil)
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
			repositoryUsers: []user.User{
				{
					Email:             "cookie@domain.tld",
					RememberTokenHash: "W3o3hteHwgT5EGSxhpyotYHNtBhEYlzfkVxViAglBuk=",
				},
			},
			rememberTokenCookie: &http.Cookie{
				Name:     UserRememberTokenCookieName,
				Value:    "tdk_BrK5adfbUapWUIeQO1VPMkGCtaQFjvF4A0KHy2g=",
				HttpOnly: true,
			},
			wantUser: &user.User{
				Email:             "cookie@domain.tld",
				RememberTokenHash: "W3o3hteHwgT5EGSxhpyotYHNtBhEYlzfkVxViAglBuk=",
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.tname, func(t *testing.T) {
			userRepository := &user.FakeRepository{
				Users: tc.repositoryUsers,
			}
			userService := user.NewService(userRepository, "ugotcookies")

			s := NewServer(userService)

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
