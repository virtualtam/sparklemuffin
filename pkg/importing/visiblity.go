package importing

import "errors"

type Visibility string

const (
	VisibilityDefault Visibility = "default"
	VisibilityPrivate Visibility = "private"
	VisibilityPublic  Visibility = "public"
)

var (
	ErrVisibilityInvalid error = errors.New("invalid value for visibility")
)
