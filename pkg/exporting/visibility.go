package exporting

import "errors"

type Visibility string

const (
	VisibilityAll     Visibility = "all"
	VisibilityPrivate Visibility = "private"
	VisibilityPublic  Visibility = "public"
)

var (
	ErrVisibilityInvalid error = errors.New("invalid value for visibility")
)
