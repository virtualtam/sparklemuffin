// Copyright VirtualTam 2022, 2026
// SPDX-License-Identifier: MIT

package session

import (
	"errors"
	"testing"
	"time"

	"github.com/virtualtam/sparklemuffin/pkg/user"
)

func TestServiceAdd(t *testing.T) {
	cases := []struct {
		tname   string
		session Session
		wantErr error
	}{
		{
			tname:   "empty session",
			wantErr: user.ErrUUIDRequired,
		},
		{
			tname: "empty remember token",
			session: Session{
				UserUUID: "0695b57a-1ab9-401d-b2db-a4430b7059ec",
			},
			wantErr: ErrRememberTokenRequired,
		},
	}

	for _, tc := range cases {
		t.Run(tc.tname, func(t *testing.T) {
			r := &FakeRepository{}
			s, err := NewService(r, "hmac-key")
			if err != nil {
				t.Fatal(err)
			}

			err = s.Add(t.Context(), tc.session)

			if tc.wantErr != nil {
				if errors.Is(err, tc.wantErr) {
					return
				}
				if err == nil {
					t.Fatalf("want error %q, got nil", tc.wantErr)
				}
				t.Fatalf("want error %q, got %q", tc.wantErr, err)
			}

			if err != nil {
				t.Fatalf("want no error, got %q", err)
			}
		})
	}
}

func TestServiceByRememberToken(t *testing.T) {
	cases := []struct {
		tname              string
		repositorySessions []Session
		token              string
		want               Session
		wantErr            error
	}{
		{
			tname:   "empty token",
			wantErr: ErrRememberTokenRequired,
		},
		{
			tname:   "not found",
			token:   "tdk_BrK5adfbUapWUIeQO1VPMkGCtaQFjvF4A0KHy2g=",
			wantErr: ErrNotFound,
		},
		{
			tname: "found",
			repositorySessions: []Session{
				{
					UserUUID:               "bf4d9fe9-25e0-4a36-b992-69c5cb611f0b",
					RememberTokenHash:      "W3o3hteHwgT5EGSxhpyotYHNtBhEYlzfkVxViAglBuk=",
					RememberTokenExpiresAt: time.Now().UTC().Add(1 * time.Hour),
				},
			},
			token: "tdk_BrK5adfbUapWUIeQO1VPMkGCtaQFjvF4A0KHy2g=",
			want: Session{
				UserUUID:          "bf4d9fe9-25e0-4a36-b992-69c5cb611f0b",
				RememberTokenHash: "W3o3hteHwgT5EGSxhpyotYHNtBhEYlzfkVxViAglBuk=",
			},
		},
		{
			tname: "expired session is not found",
			repositorySessions: []Session{
				{
					UserUUID:               "bf4d9fe9-25e0-4a36-b992-69c5cb611f0b",
					RememberTokenHash:      "W3o3hteHwgT5EGSxhpyotYHNtBhEYlzfkVxViAglBuk=",
					RememberTokenExpiresAt: time.Now().UTC().Add(-1 * time.Hour),
				},
			},
			token:   "tdk_BrK5adfbUapWUIeQO1VPMkGCtaQFjvF4A0KHy2g=",
			wantErr: ErrNotFound,
		},
	}

	for _, tc := range cases {
		t.Run(tc.tname, func(t *testing.T) {
			r := &FakeRepository{
				Sessions: tc.repositorySessions,
			}
			s, err := NewService(r, "ugotcookies")
			if err != nil {
				t.Fatal(err)
			}

			got, err := s.ByRememberToken(t.Context(), tc.token)

			if tc.wantErr != nil {
				if errors.Is(err, tc.wantErr) {
					return
				}
				if err == nil {
					t.Fatalf("want error %q, got nil", tc.wantErr)
				}
				t.Fatalf("want error %q, got %q", tc.wantErr, err)
			}

			if err != nil {
				t.Fatalf("want no error, got %q", err)
			}

			if got.UserUUID != tc.want.UserUUID {
				t.Errorf("want user UUID %q, got %q", tc.want.UserUUID, got.UserUUID)
			}
		})
	}
}

func TestServiceRequireRememberTokenHash(t *testing.T) {
	s := &Service{}

	cases := []struct {
		tname   string
		session Session
		wantErr error
	}{
		{
			tname:   "empty hash, non-empty token",
			session: Session{RememberToken: "some-token"},
			wantErr: ErrRememberTokenHashRequired,
		},
		{
			tname:   "non-empty hash",
			session: Session{RememberToken: "some-token", RememberTokenHash: "some-hash"},
		},
	}

	for _, tc := range cases {
		t.Run(tc.tname, func(t *testing.T) {
			err := s.requireRememberTokenHash(&tc.session)

			if tc.wantErr != nil {
				if !errors.Is(err, tc.wantErr) {
					t.Fatalf("want error %q, got %q", tc.wantErr, err)
				}
				return
			}

			if err != nil {
				t.Fatalf("want no error, got %q", err)
			}
		})
	}
}

func TestServiceDeleteByRememberToken(t *testing.T) {
	cases := []struct {
		tname              string
		repositorySessions []Session
		token              string
		wantErr            error
		wantRemaining      int
	}{
		{
			tname:   "empty token",
			wantErr: ErrRememberTokenRequired,
		},
		{
			tname: "deletes the matching session",
			repositorySessions: []Session{
				{
					UserUUID:          "bf4d9fe9-25e0-4a36-b992-69c5cb611f0b",
					RememberTokenHash: "W3o3hteHwgT5EGSxhpyotYHNtBhEYlzfkVxViAglBuk=",
				},
				{
					UserUUID:          "0695b57a-1ab9-401d-b2db-a4430b7059ec",
					RememberTokenHash: "other-hash",
				},
			},
			token:         "tdk_BrK5adfbUapWUIeQO1VPMkGCtaQFjvF4A0KHy2g=",
			wantRemaining: 1,
		},
	}

	for _, tc := range cases {
		t.Run(tc.tname, func(t *testing.T) {
			r := &FakeRepository{
				Sessions: tc.repositorySessions,
			}
			s, err := NewService(r, "ugotcookies")
			if err != nil {
				t.Fatal(err)
			}

			err = s.DeleteByRememberToken(t.Context(), tc.token)

			if tc.wantErr != nil {
				if errors.Is(err, tc.wantErr) {
					return
				}
				if err == nil {
					t.Fatalf("want error %q, got nil", tc.wantErr)
				}
				t.Fatalf("want error %q, got %q", tc.wantErr, err)
			}

			if err != nil {
				t.Fatalf("want no error, got %q", err)
			}

			if len(r.Sessions) != tc.wantRemaining {
				t.Errorf("want %d remaining sessions, got %d", tc.wantRemaining, len(r.Sessions))
			}
		})
	}
}

func TestServiceDeleteByUserUUID(t *testing.T) {
	cases := []struct {
		tname              string
		repositorySessions []Session
		userUUID           string
		wantErr            error
		wantRemaining      int
	}{
		{
			tname:   "empty user UUID",
			wantErr: user.ErrUUIDRequired,
		},
		{
			tname: "deletes all sessions for the user",
			repositorySessions: []Session{
				{UserUUID: "bf4d9fe9-25e0-4a36-b992-69c5cb611f0b", RememberTokenHash: "hash-1"},
				{UserUUID: "bf4d9fe9-25e0-4a36-b992-69c5cb611f0b", RememberTokenHash: "hash-2"},
				{UserUUID: "0695b57a-1ab9-401d-b2db-a4430b7059ec", RememberTokenHash: "hash-3"},
			},
			userUUID:      "bf4d9fe9-25e0-4a36-b992-69c5cb611f0b",
			wantRemaining: 1,
		},
	}

	for _, tc := range cases {
		t.Run(tc.tname, func(t *testing.T) {
			r := &FakeRepository{
				Sessions: tc.repositorySessions,
			}
			s, err := NewService(r, "ugotcookies")
			if err != nil {
				t.Fatal(err)
			}

			err = s.DeleteByUserUUID(t.Context(), tc.userUUID)

			if tc.wantErr != nil {
				if errors.Is(err, tc.wantErr) {
					return
				}
				if err == nil {
					t.Fatalf("want error %q, got nil", tc.wantErr)
				}
				t.Fatalf("want error %q, got %q", tc.wantErr, err)
			}

			if err != nil {
				t.Fatalf("want no error, got %q", err)
			}

			if len(r.Sessions) != tc.wantRemaining {
				t.Errorf("want %d remaining sessions, got %d", tc.wantRemaining, len(r.Sessions))
			}
		})
	}
}
