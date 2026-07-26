// Copyright VirtualTam 2022, 2026
// SPDX-License-Identifier: MIT

package controller

import (
	"errors"
	"fmt"

	"github.com/virtualtam/sparklemuffin/pkg/user"
)

var userFacingErrorMessages = map[error]string{
	user.ErrEmailRequired:                "Email is required.",
	user.ErrEmailAlreadyRegistered:       "This email address is already registered.",
	user.ErrNickNameRequired:             "Nickname is required.",
	user.ErrNickNameInvalid:              "This nickname is invalid.",
	user.ErrNickNameAlreadyRegistered:    "This nickname is already taken.",
	user.ErrDisplayNameRequired:          "Display name is required.",
	user.ErrPasswordRequired:             "Password is required.",
	user.ErrPasswordTooShort:             fmt.Sprintf("Password must be at least %d characters long.", user.MinPasswordLength),
	user.ErrPasswordIncorrect:            "Your current password is incorrect.",
	user.ErrPasswordConfirmationMismatch: "The new password and confirmation do not match.",
}

// userFacingError maps a domain error returned by the user package to a
// message safe to display to an end user, falling back to a generic message
// for anything not explicitly mapped (e.g. ErrNotFound, storage errors).
func userFacingError(err error) string {
	for domainErr, message := range userFacingErrorMessages {
		if errors.Is(err, domainErr) {
			return message
		}
	}

	return "Something went wrong. Please try again."
}
