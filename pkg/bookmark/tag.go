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
	return requireUserUUID(dq.UserUUID)
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

// TagUpdateQuery represents a tag name update for all bookmarks for an authenticated user.
type TagUpdateQuery struct {
	UserUUID    string
	CurrentName string
	NewName     string
}

func (uq *TagUpdateQuery) ensureCurrentNameHasNoWhitespace() error {
	if whitespaceRegexp.MatchString(uq.CurrentName) {
		return newValidationError("current", ErrTagNameContainsWhitespace)
	}
	return nil
}

func (uq *TagUpdateQuery) ensureNewNameHasNoWhitespace() error {
	if whitespaceRegexp.MatchString(uq.NewName) {
		return newValidationError("new", ErrTagNameContainsWhitespace)
	}
	return nil
}

func (uq *TagUpdateQuery) ensureNewNameIsNotEqualToCurrentName() error {
	if uq.CurrentName == uq.NewName {
		return ErrTagNewNameEqualsCurrentName
	}
	return nil
}

func (uq *TagUpdateQuery) normalize() {
	uq.CurrentName = strings.TrimSpace(uq.CurrentName)
	uq.NewName = strings.TrimSpace(uq.NewName)
}

func (uq *TagUpdateQuery) requireCurrentName() error {
	if uq.CurrentName == "" {
		return newValidationError("current", ErrTagNameRequired)
	}
	return nil
}

func (uq *TagUpdateQuery) requireNewName() error {
	if uq.NewName == "" {
		return newValidationError("new", ErrTagNameRequired)
	}
	return nil
}

func (uq *TagUpdateQuery) requireUserUUID() error {
	return requireUserUUID(uq.UserUUID)
}
