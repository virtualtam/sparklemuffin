package importing

import "errors"

type OnConflictStrategy string

const (
	OnConflictOverwrite    OnConflictStrategy = "overwrite"
	OnConflictKeepExisting OnConflictStrategy = "keep"
)

var (
	ErrOnConflictStrategyInvalid error = errors.New("invalid value for on-conflict strategy")
)
