package user

import (
	"errors"
	"testing"
)

func TestServiceAdd(t *testing.T) {
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
			tname:   "empty email (whitespace)",
			user:    User{Email: "    "},
			wantErr: ErrEmailRequired,
		},
		{
			tname: "already registered",
			repositoryUsers: []User{
				{Email: "registered@domain.tld"},
			},
			user:    User{Email: "registered@domain.tld"},
			wantErr: ErrEmailAlreadyRegistered,
		},
		{
			tname:   "empty password",
			user:    User{Email: "nopass@domain.tld"},
			wantErr: ErrPasswordRequired,
		},
		{
			tname: "valid user",
			user: User{
				Email:    "new@domain.tld",
				Password: "ImN3w!",
			},
		},
		{
			tname: "valid adminuser",
			user: User{
				Email:    "new@domain.tld",
				Password: "ImN3w!",
				IsAdmin:  true,
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.tname, func(t *testing.T) {
			r := &FakeRepository{
				Users: tc.repositoryUsers,
			}
			s := NewService(r, "hmac-key")

			err := s.Add(tc.user)

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

			got, err := r.GetUserByEmail(tc.user.Email)
			if err != nil {
				t.Fatalf("want no error, got %q", err)
			}

			if got.Password != "" {
				t.Errorf("password must be empty, got %q", got.Password)
			}
			if got.PasswordHash == "" {
				t.Error("password hash must be set")
			}
			if got.UUID == "" {
				t.Error("UUID must be set")
			}
		})
	}
}

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

func TestServiceDeleteByUUID(t *testing.T) {
	cases := []struct {
		tname           string
		repositoryUsers []User
		userUUID        string
		wantErr         error
	}{
		{
			tname:   "empty UUID",
			wantErr: ErrUUIDRequired,
		},
		{
			tname:    "unknown UUID",
			userUUID: "b52cd2d5-89f7-4489-b023-722896ca3f98",
			wantErr:  ErrNotFound,
		},
		{
			tname: "delete by UUID",
			repositoryUsers: []User{
				{UUID: "ebd1bec1-e15f-4502-ae97-a631f7d7df91"},
			},
			userUUID: "ebd1bec1-e15f-4502-ae97-a631f7d7df91",
			wantErr:  ErrNotFound,
		},
	}

	for _, tc := range cases {
		t.Run(tc.tname, func(t *testing.T) {
			r := &FakeRepository{
				Users: tc.repositoryUsers,
			}
			s := NewService(r, "hmac-key")

			err := s.DeleteByUUID(tc.userUUID)

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

func TestServiceUpdate(t *testing.T) {
	cases := []struct {
		tname           string
		repositoryUsers []User
		user            User
		wantErr         error
	}{
		{
			tname:   "empty user",
			wantErr: ErrUUIDRequired,
		},
		{
			tname: "empty (whitespace) email",
			user: User{
				UUID:  "a6548986-5ae4-4ad3-b208-c2cf3fab4e08",
				Email: "   ",
			},
			wantErr: ErrEmailRequired,
		},
		{
			tname: "empty password",
			user: User{
				UUID:  "a6548986-5ae4-4ad3-b208-c2cf3fab4e08",
				Email: "nopass@domain.tld",
			},
			wantErr: ErrPasswordRequired,
		},
		{
			tname: "not found",
			user: User{
				UUID:         "a6548986-5ae4-4ad3-b208-c2cf3fab4e08",
				Email:        "ghost@domain.tld",
				Password:     "test",
				PasswordHash: "$2b$10$LSH.kwYeRt8msI5.5YJv8eqle6SPcevq848BK2vZ2M5FjXTvU1r.e",
			},
			wantErr: ErrNotFound,
		},
		{
			tname: "update user",
			user: User{
				UUID:         "a6548986-5ae4-4ad3-b208-c2cf3fab4e08",
				Email:        "valid@domain.tld",
				Password:     "test",
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
