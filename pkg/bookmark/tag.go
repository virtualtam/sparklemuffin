package bookmark

import (
	"regexp"
	"strings"
)

var (
	whitespaceRegexp = regexp.MustCompile(`\s`)
)

// TagNameUpdate represents a tag name update for all bookmarks for an authenticated user.
type TagNameUpdate struct {
	UserUUID    string
	CurrentName string
	NewName     string
}

func (u *TagNameUpdate) ensureCurrentNameHasNoWhitespace() error {
	if whitespaceRegexp.MatchString(u.CurrentName) {
		return ErrTagCurrentNameContainsWhitespace
	}
	return nil
}

func (u *TagNameUpdate) ensureNewNameHasNoWhitespace() error {
	if whitespaceRegexp.MatchString(u.NewName) {
		return ErrTagNewNameContainsWhitespace
	}
	return nil
}

func (u *TagNameUpdate) ensureNewNameIsNotEqualToCurrentName() error {
	if u.CurrentName == u.NewName {
		return ErrTagNewNameEqualsCurrentName
	}
	return nil
}

func (u *TagNameUpdate) normalize() {
	u.CurrentName = strings.TrimSpace(u.CurrentName)
	u.NewName = strings.TrimSpace(u.NewName)
}

func (u *TagNameUpdate) requireCurrentName() error {
	if u.CurrentName == "" {
		return ErrTagCurrentNameRequired
	}
	return nil
}

func (u *TagNameUpdate) requireNewName() error {
	if u.NewName == "" {
		return ErrTagNewNameRequired
	}
	return nil
}

func (u *TagNameUpdate) requireUserUUID() error {
	if u.UserUUID == "" {
		return ErrUserUUIDRequired
	}
	return nil
}
