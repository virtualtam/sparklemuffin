// Copyright VirtualTam 2022, 2026
// SPDX-License-Identifier: MIT

package user

import (
	"errors"
	"testing"

	"github.com/jaswdr/faker/v2"
	"golang.org/x/crypto/bcrypt"
)

func TestServiceAdd(t *testing.T) {
	fake := faker.New()

	cases := []struct {
		tname           string
		repositoryUsers []User
		user            User
		wantErr         error
	}{
		// Nominal cases.
		{
			tname: "valid user",
			user: User{
				UUID:        fake.UUID().V4(),
				Email:       "new@domain.tld",
				NickName:    "dat-new-pal3",
				DisplayName: "The New Pal",
				Password:    FakePassword(t, &fake),
			},
		},
		{
			tname: "valid admin user",
			user: User{
				UUID:        fake.UUID().V4(),
				Email:       "newadmin@domain.tld",
				NickName:    "newadmin",
				DisplayName: "PID One",
				Password:    FakePassword(t, &fake),
				IsAdmin:     true,
			},
		},

		// Error cases.
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
			tname: "email already registered",
			repositoryUsers: []User{
				{Email: "registered@domain.tld"},
			},
			user:    User{Email: "registered@domain.tld"},
			wantErr: ErrEmailAlreadyRegistered,
		},
		{
			tname: "empty nick (whitespace)",
			user: User{
				Email:    "nickless@domain.tld",
				NickName: "   ",
			},
			wantErr: ErrNickNameRequired,
		},
		{
			tname: "invalid nick (whitespace)",
			user: User{
				Email:    "spacenick@domain.tld",
				NickName: "s p a c e",
			},
			wantErr: ErrNickNameInvalid,
		},
		{
			tname: "invalid nick (slash)",
			user: User{
				Email:    "saul@huds.on",
				NickName: "s/lash",
			},
			wantErr: ErrNickNameInvalid,
		},
		{
			tname: "invalid nick (question mark)",
			user: User{
				Email:    "invader@domain.tld",
				NickName: "s?pace",
			},
			wantErr: ErrNickNameInvalid,
		},
		{
			tname: "nick already registered",
			repositoryUsers: []User{
				{
					Email:    "registered@domain.tld",
					NickName: "regis33",
				},
			},
			user: User{
				Email:    "regis33@domain.tld",
				NickName: "regis33",
			},
			wantErr: ErrNickNameAlreadyRegistered,
		},
		{
			tname: "empty display name",
			user: User{
				Email:    "noname@domain.tld",
				NickName: "noname",
			},
			wantErr: ErrDisplayNameRequired,
		},
		{
			tname: "empty password",
			user: User{
				Email:       "nopass@domain.tld",
				NickName:    "nopass",
				DisplayName: "No Pass",
			},
			wantErr: ErrPasswordRequired,
		},
		{
			tname: "password too short",
			user: User{
				Email:       "shortpass@domain.tld",
				NickName:    "shortpass",
				DisplayName: "Short Pass",
				Password:    fake.Lorem().Text(MinPasswordLength - 1),
			},
			wantErr: ErrPasswordTooShort,
		},
	}

	for _, tc := range cases {
		t.Run(tc.tname, func(t *testing.T) {
			r := &FakeRepository{
				Users: tc.repositoryUsers,
			}
			s := NewService(r)

			err := s.Add(t.Context(), tc.user)

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

			got, err := r.UserGetByEmail(t.Context(), tc.user.Email)
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

func TestServiceAll(t *testing.T) {
	cases := []struct {
		tname           string
		repositoryUsers []User
		wantLen         int
	}{
		{
			tname:   "no users",
			wantLen: 0,
		},
		{
			tname: "2 users",
			repositoryUsers: []User{
				{Email: "one@domain.tld"},
				{Email: "two@domain.tld"},
			},
			wantLen: 2,
		},
	}

	for _, tc := range cases {
		t.Run(tc.tname, func(t *testing.T) {
			r := &FakeRepository{Users: tc.repositoryUsers}
			s := NewService(r)

			got, err := s.All(t.Context())

			if err != nil {
				t.Fatalf("want no error, got %q", err)
			}
			if len(got) != tc.wantLen {
				t.Fatalf("want %d users, got %d", tc.wantLen, len(got))
			}
		})
	}
}

func TestServiceByNickName(t *testing.T) {
	cases := []struct {
		tname           string
		repositoryUsers []User
		nick            string
		want            User
		wantErr         error
	}{
		// error cases
		{
			tname:   "empty nick",
			wantErr: ErrNickNameRequired,
		},
		{
			tname:   "empty nick (whitespace)",
			nick:    "   ",
			wantErr: ErrNickNameRequired,
		},
		{
			tname:   "invalid nick (space)",
			nick:    "in valid",
			wantErr: ErrNickNameInvalid,
		},
		{
			tname:   "not found",
			nick:    "ghost",
			wantErr: ErrNotFound,
		},

		// nominal cases
		{
			tname: "found",
			repositoryUsers: []User{
				{NickName: "testuser", Email: "test@domain.tld"},
			},
			nick: "testuser",
			want: User{NickName: "testuser", Email: "test@domain.tld"},
		},
		{
			tname: "found (nick normalized to lowercase)",
			repositoryUsers: []User{
				{NickName: "testuser", Email: "test@domain.tld"},
			},
			nick: "TestUser",
			want: User{NickName: "testuser", Email: "test@domain.tld"},
		},
	}

	for _, tc := range cases {
		t.Run(tc.tname, func(t *testing.T) {
			r := &FakeRepository{Users: tc.repositoryUsers}
			s := NewService(r)

			got, err := s.ByNickName(t.Context(), tc.nick)

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

func TestServiceByUUID(t *testing.T) {
	cases := []struct {
		tname           string
		repositoryUsers []User
		userUUID        string
		want            User
		wantErr         error
	}{
		// error cases
		{
			tname:   "empty UUID",
			wantErr: ErrUUIDRequired,
		},
		{
			tname:    "not found",
			userUUID: "b52cd2d5-89f7-4489-b023-722896ca3f98",
			wantErr:  ErrNotFound,
		},

		// nominal cases
		{
			tname: "found",
			repositoryUsers: []User{
				{UUID: "b52cd2d5-89f7-4489-b023-722896ca3f98", Email: "found@domain.tld"},
			},
			userUUID: "b52cd2d5-89f7-4489-b023-722896ca3f98",
			want:     User{UUID: "b52cd2d5-89f7-4489-b023-722896ca3f98", Email: "found@domain.tld"},
		},
	}

	for _, tc := range cases {
		t.Run(tc.tname, func(t *testing.T) {
			r := &FakeRepository{Users: tc.repositoryUsers}
			s := NewService(r)

			got, err := s.ByUUID(t.Context(), tc.userUUID)

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
			tname: "wrong password",
			repositoryUsers: []User{
				{
					Email:        "found@domain.tld",
					PasswordHash: "$2b$10$J0z6wKdvrPMmbUgg.uhhROv0Zp4bFQ19GnTshpsazLpK2l5fOnEmy",
				},
			},
			email:    "found@domain.tld",
			password: "nottest",
			wantErr:  ErrPasswordIncorrect,
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
			s := NewService(r)

			got, err := s.Authenticate(t.Context(), tc.email, tc.password)

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
		},
	}

	for _, tc := range cases {
		t.Run(tc.tname, func(t *testing.T) {
			r := &FakeRepository{
				Users: tc.repositoryUsers,
			}
			s := NewService(r)

			err := s.DeleteByUUID(t.Context(), tc.userUUID)

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
	fake := faker.New()
	existingUser := FakeUser(t, &fake)

	newPassword := FakePassword(t, &fake)
	newPasswordHashBytes, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		t.Fatalf("failed to generate new password hash: %v", err)
	}

	newPasswordHash := string(newPasswordHashBytes)

	shortPassword := FakePassword(t, &fake)[:MinPasswordLength-1]

	cases := []struct {
		tname           string
		repositoryUsers []User
		user            User
		wantErr         error
	}{
		// Nominal cases.
		{
			tname: "update an existing user",
			repositoryUsers: []User{
				existingUser,
			},
			user: User{
				UUID:         existingUser.UUID,
				Email:        existingUser.Email,
				NickName:     existingUser.NickName,
				DisplayName:  existingUser.DisplayName,
				Password:     newPassword,
				PasswordHash: newPasswordHash,
			},
		},

		// Error cases.
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
			tname: "empty (whitespace) nick",
			user: User{
				UUID:     "a6548986-5ae4-4ad3-b208-c2cf3fab4e08",
				Email:    "nonick@domain.tld",
				NickName: "   ",
			},
			wantErr: ErrNickNameRequired,
		},
		{
			tname: "invalid nick (whitespace)",
			user: User{
				UUID:     "a6548986-5ae4-4ad3-b208-c2cf3fab4e08",
				Email:    "spacenick@domain.tld",
				NickName: "s p a c e",
			},
			wantErr: ErrNickNameInvalid,
		},
		{
			tname: "invalid nick (slash)",
			user: User{
				UUID:     "a6548986-5ae4-4ad3-b208-c2cf3fab4e08",
				Email:    "saul@huds.on",
				NickName: "s/lash",
			},
			wantErr: ErrNickNameInvalid,
		},
		{
			tname: "invalid nick (question mark)",
			user: User{
				UUID:     "a6548986-5ae4-4ad3-b208-c2cf3fab4e08",
				Email:    "invader@domain.tld",
				NickName: "s?pace",
			},
			wantErr: ErrNickNameInvalid,
		},
		{
			tname: "empty (whitespace) display name",
			user: User{
				UUID:        "a6548986-5ae4-4ad3-b208-c2cf3fab4e08",
				Email:       "noname@domain.tld",
				NickName:    "noname",
				DisplayName: "   ",
			},
			wantErr: ErrDisplayNameRequired,
		},
		{
			tname: "empty password",
			user: User{
				UUID:        "a6548986-5ae4-4ad3-b208-c2cf3fab4e08",
				NickName:    "nopass",
				DisplayName: "No Pass",
				Email:       "nopass@domain.tld",
			},
			wantErr: ErrPasswordRequired,
		},
		{
			tname: "password too short",
			user: User{
				UUID:        "a6548986-5ae4-4ad3-b208-c2cf3fab4e08",
				NickName:    "nopass",
				DisplayName: "No Pass",
				Email:       "nopass@domain.tld",
				Password:    shortPassword,
			},
			wantErr: ErrPasswordTooShort,
		},
		{
			tname:   "not found",
			user:    FakeUser(t, &fake),
			wantErr: ErrNotFound,
		},
	}

	for _, tc := range cases {
		t.Run(tc.tname, func(t *testing.T) {
			r := &FakeRepository{
				Users: tc.repositoryUsers,
			}
			s := NewService(r)

			err := s.Update(t.Context(), tc.user)

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

func TestServiceUpdateInfo(t *testing.T) {
	cases := []struct {
		tname           string
		repositoryUsers []User
		info            InfoUpdate
		wantErr         error
	}{
		{
			tname:   "empty update",
			wantErr: ErrUUIDRequired,
		},
		{
			tname: "empty email",
			info: InfoUpdate{
				UserUUID: "2a16ed9e-fdb0-4d8e-a196-3fe4d24d1c34",
			},
			wantErr: ErrEmailRequired,
		},
		{
			tname: "email already registered",
			repositoryUsers: []User{
				{
					UUID:  "5a347515-e178-4aeb-bf3e-cf1a56b50c02",
					Email: "mimic@domain.tld",
				},
				{
					UUID:  "2a16ed9e-fdb0-4d8e-a196-3fe4d24d1c34",
					Email: "sleuth@domain.tld",
				},
			},
			info: InfoUpdate{
				UserUUID: "2a16ed9e-fdb0-4d8e-a196-3fe4d24d1c34",
				Email:    "mimic@domain.tld",
			},
			wantErr: ErrEmailAlreadyRegistered,
		},
		{
			tname: "empty nick (whitespace)",
			repositoryUsers: []User{
				{
					UUID:  "2a16ed9e-fdb0-4d8e-a196-3fe4d24d1c34",
					Email: "mimic@domain.tld",
				},
			},
			info: InfoUpdate{
				UserUUID: "2a16ed9e-fdb0-4d8e-a196-3fe4d24d1c34",
				Email:    "mimic@domain.tld",
				NickName: "    ",
			},
			wantErr: ErrNickNameRequired,
		},
		{
			tname: "invalid nick (whitespace)",
			repositoryUsers: []User{
				{
					UUID:  "2a16ed9e-fdb0-4d8e-a196-3fe4d24d1c34",
					Email: "mimic@domain.tld",
				},
			},
			info: InfoUpdate{
				UserUUID: "2a16ed9e-fdb0-4d8e-a196-3fe4d24d1c34",
				Email:    "mimic@domain.tld",
				NickName: "s p a c e",
			},
			wantErr: ErrNickNameInvalid,
		},
		{
			tname: "invalid nick (slash)",
			repositoryUsers: []User{
				{
					UUID:  "2a16ed9e-fdb0-4d8e-a196-3fe4d24d1c34",
					Email: "mimic@domain.tld",
				},
			},
			info: InfoUpdate{
				UserUUID: "2a16ed9e-fdb0-4d8e-a196-3fe4d24d1c34",
				Email:    "mimic@domain.tld",
				NickName: "s/lash",
			},
			wantErr: ErrNickNameInvalid,
		},
		{
			tname: "invalid nick (question mark)",
			repositoryUsers: []User{
				{
					UUID:  "2a16ed9e-fdb0-4d8e-a196-3fe4d24d1c34",
					Email: "mimic@domain.tld",
				},
			},
			info: InfoUpdate{
				UserUUID: "2a16ed9e-fdb0-4d8e-a196-3fe4d24d1c34",
				Email:    "mimic@domain.tld",
				NickName: "s?lash",
			},
			wantErr: ErrNickNameInvalid,
		},
		{
			tname: "empty display name (whitespace)",
			repositoryUsers: []User{
				{
					UUID:  "2a16ed9e-fdb0-4d8e-a196-3fe4d24d1c34",
					Email: "empty@domain.tld",
				},
			},
			info: InfoUpdate{
				UserUUID:    "2a16ed9e-fdb0-4d8e-a196-3fe4d24d1c34",
				Email:       "empty@domain.tld",
				NickName:    "empty",
				DisplayName: "   ",
			},
			wantErr: ErrDisplayNameRequired,
		},
		{
			tname: "not found",
			info: InfoUpdate{
				UserUUID:    "2a16ed9e-fdb0-4d8e-a196-3fe4d24d1c34",
				Email:       "ghost@domain.tld",
				NickName:    "ghost",
				DisplayName: "Busted Ghost",
			},
			wantErr: ErrNotFound,
		},
		{
			tname: "update with no change",
			repositoryUsers: []User{
				{
					UUID:        "2a16ed9e-fdb0-4d8e-a196-3fe4d24d1c34",
					Email:       "mimic@domain.tld",
					NickName:    "mimic",
					DisplayName: "Mimic",
				},
			},
			info: InfoUpdate{
				UserUUID:    "2a16ed9e-fdb0-4d8e-a196-3fe4d24d1c34",
				Email:       "mimic@domain.tld",
				NickName:    "Mimic",
				DisplayName: "Mimic",
			},
		},
		{
			tname: "update with new information",
			repositoryUsers: []User{
				{
					UUID:        "2a16ed9e-fdb0-4d8e-a196-3fe4d24d1c34",
					Email:       "mimic@domain.tld",
					NickName:    "mimic",
					DisplayName: "Mimic",
				},
			},
			info: InfoUpdate{
				UserUUID:    "2a16ed9e-fdb0-4d8e-a196-3fe4d24d1c34",
				Email:       "chest@domain.tld",
				NickName:    "chester",
				DisplayName: "Chester",
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.tname, func(t *testing.T) {
			r := &FakeRepository{
				Users: tc.repositoryUsers,
			}
			s := NewService(r)

			err := s.UpdateInfo(t.Context(), tc.info)

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

func TestServiceUpdatePassword(t *testing.T) {
	fake := faker.New()

	existingUser := FakeUser(t, &fake)
	currentPassword := existingUser.Password
	if err := existingUser.hashPassword(); err != nil {
		t.Fatal(err)
	}

	newPassword := FakePassword(t, &fake)
	shortPassword := FakePassword(t, &fake)[:MinPasswordLength-1]

	cases := []struct {
		tname           string
		repositoryUsers []User
		passwordUpdate  PasswordUpdate
		wantErr         error
	}{
		// Nominal cases.
		{
			tname: "password update",
			repositoryUsers: []User{
				existingUser,
			},
			passwordUpdate: PasswordUpdate{
				UserUUID:                existingUser.UUID,
				CurrentPassword:         currentPassword,
				NewPassword:             newPassword,
				NewPasswordConfirmation: newPassword,
			},
		},

		// Error cases.
		{
			tname:   "empty update",
			wantErr: ErrUUIDRequired,
		},
		{
			tname: "empty password",
			passwordUpdate: PasswordUpdate{
				UserUUID: "546e3bff-5dbb-4269-ab01-c35a90c382dc",
			},
			wantErr: ErrPasswordRequired,
		},
		{
			tname: "user not found",
			passwordUpdate: PasswordUpdate{
				UserUUID:        "546e3bff-5dbb-4269-ab01-c35a90c382dc",
				CurrentPassword: "test",
			},
			wantErr: ErrNotFound,
		},
		{
			tname: "invalid current password",
			repositoryUsers: []User{
				{
					UUID: "546e3bff-5dbb-4269-ab01-c35a90c382dc",
					// Password: "test"
					PasswordHash: "$2b$10$AIUHvtnoIppMHkhpoTFdROVwedB9YC.iJvGaHpnIXEUesD6VHTLLK",
				},
			},
			passwordUpdate: PasswordUpdate{
				UserUUID:        "546e3bff-5dbb-4269-ab01-c35a90c382dc",
				CurrentPassword: "isitnottest?",
			},
			wantErr: ErrPasswordIncorrect,
		},
		{
			tname: "new password and confirmation mismatch",
			repositoryUsers: []User{
				{
					UUID: "546e3bff-5dbb-4269-ab01-c35a90c382dc",
					// Password: "test"
					PasswordHash: "$2b$10$AIUHvtnoIppMHkhpoTFdROVwedB9YC.iJvGaHpnIXEUesD6VHTLLK",
				},
			},
			passwordUpdate: PasswordUpdate{
				UserUUID:                "546e3bff-5dbb-4269-ab01-c35a90c382dc",
				CurrentPassword:         "test",
				NewPassword:             "asdf",
				NewPasswordConfirmation: "qsdf",
			},
			wantErr: ErrPasswordConfirmationMismatch,
		},
		{
			tname: "password too short",
			repositoryUsers: []User{
				existingUser,
			},
			passwordUpdate: PasswordUpdate{
				UserUUID:                existingUser.UUID,
				CurrentPassword:         currentPassword,
				NewPassword:             shortPassword,
				NewPasswordConfirmation: shortPassword,
			},
			wantErr: ErrPasswordTooShort,
		},
	}

	for _, tc := range cases {
		t.Run(tc.tname, func(t *testing.T) {
			r := &FakeRepository{
				Users: tc.repositoryUsers,
			}
			s := NewService(r)

			err := s.UpdatePassword(t.Context(), tc.passwordUpdate)

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
	if got.IsAdmin != want.IsAdmin {
		t.Errorf("want admin %t, got %t", want.IsAdmin, got.IsAdmin)
	}
	if got.PasswordHash != want.PasswordHash {
		t.Errorf("want password hash %q, got %q", want.PasswordHash, got.PasswordHash)
	}
}
