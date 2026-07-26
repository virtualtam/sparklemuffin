// Copyright VirtualTam 2022, 2026
// SPDX-License-Identifier: MIT

package controller

import (
	"errors"
	"fmt"
	"testing"

	"github.com/virtualtam/sparklemuffin/pkg/user"
)

func TestUserFacingError(t *testing.T) {
	passwordTooShort := fmt.Sprintf("Password must be at least %d characters long.", user.MinPasswordLength)

	cases := []struct {
		tname string
		err   error
		want  string
	}{
		{"email required", user.ErrEmailRequired, "Email is required."},
		{"email already registered", user.ErrEmailAlreadyRegistered, "This email address is already registered."},
		{"nickname required", user.ErrNickNameRequired, "Nickname is required."},
		{"nickname invalid", user.ErrNickNameInvalid, "This nickname is invalid."},
		{"nickname already registered", user.ErrNickNameAlreadyRegistered, "This nickname is already taken."},
		{"display name required", user.ErrDisplayNameRequired, "Display name is required."},
		{"password required", user.ErrPasswordRequired, "Password is required."},
		{"password too short", user.ErrPasswordTooShort, passwordTooShort},
		{"password incorrect", user.ErrPasswordIncorrect, "Your current password is incorrect."},
		{"password confirmation mismatch", user.ErrPasswordConfirmationMismatch, "The new password and confirmation do not match."},
		{"wrapped sentinel is still recognized", fmt.Errorf("wrap: %w", user.ErrPasswordTooShort), passwordTooShort},
		{"unmapped error falls back to a generic message", errors.New("some internal detail"), "Something went wrong. Please try again."},
	}

	for _, tc := range cases {
		t.Run(tc.tname, func(t *testing.T) {
			if got := userFacingError(tc.err); got != tc.want {
				t.Errorf("userFacingError(%v) = %q, want %q", tc.err, got, tc.want)
			}
		})
	}
}
