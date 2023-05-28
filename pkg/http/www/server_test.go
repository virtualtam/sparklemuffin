package www

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/virtualtam/sparklemuffin/pkg/session"
	"github.com/virtualtam/sparklemuffin/pkg/user"
)

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

			s := NewServer(
				WithSessionService(sessionService),
				WithUserService(userService),
			)

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
