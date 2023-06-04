package bookmark

import (
	"regexp"
	"strings"
)

var (
	whitespaceRegexp = regexp.MustCompile(`\s`)
)

// TagDeleteQuery represents a tag deletion for all bookmarks of an authenticated user.
type TagDeleteQuery struct {
	UserUUID string
	Name     string
}

func (dq *TagDeleteQuery) normalize() {
	dq.Name = strings.TrimSpace(dq.Name)
}

func (dq *TagDeleteQuery) requireUserUUID() error {
	if dq.UserUUID == "" {
		return ErrUserUUIDRequired
	}
	return nil
}

func (dq *TagDeleteQuery) ensureNameHasNoWhitespace() error {
	if whitespaceRegexp.MatchString(dq.Name) {
		return ErrTagNameContainsWhitespace
	}
	return nil
}

func (dq *TagDeleteQuery) requireName() error {
	if dq.Name == "" {
		return ErrTagNameRequired
	}
	return nil
}

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
