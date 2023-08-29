package version

import (
	"time"

	"github.com/carlmjohnson/versioninfo"
)

// Details provides detailed build information about the application.
type Details struct {
	Short       string     `json:"short"`
	Revision    string     `json:"revision"`
	CommittedAt *time.Time `json:"commited_at,omitempty"`
	DirtyBuild  bool       `json:"dirty_build"`
}

// NewDetails retrieves and returns build version Details.
func NewDetails() *Details {
	v := &Details{
		Short:      versioninfo.Short(),
		Revision:   versioninfo.Revision,
		DirtyBuild: versioninfo.DirtyBuild,
	}

	if !versioninfo.LastCommit.IsZero() {
		v.CommittedAt = &versioninfo.LastCommit
	}

	return v
}