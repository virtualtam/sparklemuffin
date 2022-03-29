package user

import (
	"errors"
	"testing"
)

func TestServiceAuthenticate(t *testing.T) {
	cases := []struct {
		tname           string
		repositoryUsers []User

		email    string
		password string

		want    User
		wantErr error
	}{
		{
			tname:   "empty email",
			wantErr: ErrEmailRequired,
		},
		{
			tname:   "empty (whitespace) email",
			email:   "   ",
			wantErr: ErrEmailRequired,
		},
		{
			tname:   "not found",
			email:   "ghost@domain.tld",
			wantErr: ErrNotFound,
		},
		{
			tname: "found",
			repositoryUsers: []User{
				{
					Email:        "found@domain.tld",
					PasswordHash: "$2b$10$J0z6wKdvrPMmbUgg.uhhROv0Zp4bFQ19GnTshpsazLpK2l5fOnEmy",
				},
			},
			email:    "found@domain.tld",
			password: "test",
			want: User{
				Email:        "found@domain.tld",
				PasswordHash: "$2b$10$J0z6wKdvrPMmbUgg.uhhROv0Zp4bFQ19GnTshpsazLpK2l5fOnEmy",
			},
		},
		{
			tname: "found (email contains whitespace)",
			repositoryUsers: []User{
				{
					Email:        "found@domain.tld",
					PasswordHash: "$2b$10$J0z6wKdvrPMmbUgg.uhhROv0Zp4bFQ19GnTshpsazLpK2l5fOnEmy",
				},
			},
			email:    "   found@domain.tld  ",
			password: "test",
			want: User{
				Email:        "found@domain.tld",
				PasswordHash: "$2b$10$J0z6wKdvrPMmbUgg.uhhROv0Zp4bFQ19GnTshpsazLpK2l5fOnEmy",
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.tname, func(t *testing.T) {
			r := &FakeRepository{
				Users: tc.repositoryUsers,
			}
			s := NewService(r, "hmac-key")

			got, err := s.Authenticate(tc.email, tc.password)

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

			assertUsersEqual(t, got, tc.want)
		})
	}
}

func TestServiceUpdate(t *testing.T) {
	cases := []struct {
		tname           string
		repositoryUsers []User
		user            User
		wantErr         error
	}{
		{
			tname:   "empty user",
			wantErr: ErrEmailRequired,
		},
		{
			tname:   "empty (whitespace) email",
			user:    User{Email: "   "},
			wantErr: ErrEmailRequired,
		},
		{
			tname: "empty password hash",
			user: User{
				Email: "unhashed@domain.tld",
			},
			wantErr: ErrPasswordHashRequired,
		},
		{
			tname: "not found",
			user: User{
				Email:        "ghost@domain.tld",
				PasswordHash: "$2b$10$LSH.kwYeRt8msI5.5YJv8eqle6SPcevq848BK2vZ2M5FjXTvU1r.e",
			},
			wantErr: ErrNotFound,
		},
		{
			tname: "update user",
			user: User{
				Email:        "ghost@domain.tld",
				PasswordHash: "$2b$10$LSH.kwYeRt8msI5.5YJv8eqle6SPcevq848BK2vZ2M5FjXTvU1r.e",
			},
			wantErr: ErrNotFound,
		},
	}

	for _, tc := range cases {
		t.Run(tc.tname, func(t *testing.T) {
			r := &FakeRepository{
				Users: tc.repositoryUsers,
			}
			s := NewService(r, "hmac-key")

			err := s.Update(tc.user)

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

func assertUsersEqual(t *testing.T, got, want User) {
	t.Helper()

	if got.Email != want.Email {
		t.Errorf("want email %q, got %q", want.Email, got.Email)
	}
	if got.PasswordHash != want.PasswordHash {
		t.Errorf("want password hash %q, got %q", want.PasswordHash, got.PasswordHash)
	}
	if got.RememberTokenHash != want.RememberTokenHash {
		t.Errorf("want remember token %q, got %q", want.RememberToken, got.RememberToken)
	}
}
