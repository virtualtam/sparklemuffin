// Copyright VirtualTam 2022, 2026
// SPDX-License-Identifier: MIT

package user // Password returns a fake password for Internet
import (
	"strings"
	"testing"

	"github.com/jaswdr/faker/v2"
)

// GenerateFakeUser generates a new user for testing.
func FakeUser(tb testing.TB, fake *faker.Faker) User {
	tb.Helper()

	person := fake.Person()
	internet := fake.Internet()

	// Nicknames must match nickNameRegex
	nick := strings.ReplaceAll(internet.User(), ".", "")

	return User{
		UUID:        fake.UUID().V4(),
		Email:       person.Contact().Email,
		NickName:    nick,
		DisplayName: person.Name(),
		Password:    FakePassword(tb, fake),
	}
}

// FakePassword returns a generated password matching security requirements.
func FakePassword(tb testing.TB, fake *faker.Faker) string {
	tb.Helper()
	pattern := strings.Repeat("*", fake.IntBetween(MinPasswordLength, 2*MinPasswordLength))
	return fake.Asciify(pattern)
}
