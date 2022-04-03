package session

import (
	"errors"
	"testing"
)

func TestServiceAdd(t *testing.T) {
	cases := []struct {
		tname   string
		session Session
		wantErr error
	}{
		{
			tname:   "empty session",
			wantErr: ErrUserUUIDRequired,
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
			s := NewService(r, "hmac-key")

			err := s.Add(tc.session)

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
					UserUUID:          "bf4d9fe9-25e0-4a36-b992-69c5cb611f0b",
					RememberTokenHash: "W3o3hteHwgT5EGSxhpyotYHNtBhEYlzfkVxViAglBuk=",
				},
			},
			token: "tdk_BrK5adfbUapWUIeQO1VPMkGCtaQFjvF4A0KHy2g=",
			want: Session{
				UserUUID:          "bf4d9fe9-25e0-4a36-b992-69c5cb611f0b",
				RememberTokenHash: "W3o3hteHwgT5EGSxhpyotYHNtBhEYlzfkVxViAglBuk=",
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.tname, func(t *testing.T) {
			r := &FakeRepository{
				Sessions: tc.repositorySessions,
			}
			s := NewService(r, "ugotcookies")

			got, err := s.ByRememberToken(tc.token)

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
