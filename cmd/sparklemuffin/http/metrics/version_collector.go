package metrics

import (
	"strconv"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/virtualtam/sparklemuffin/cmd/sparklemuffin/version"
)

type versionCollector struct {
	commitedAtEpoch string
	isDirty         string
	revision        string
	version         string

	versionDesc *prometheus.Desc
}

func newVersionCollector(metricsPrefix string, versionDetails *version.Details) prometheus.Collector {
	var commitedAtEpoch string

	if versionDetails.CommittedAt != nil && !versionDetails.CommittedAt.IsZero() {
		commitedAtEpoch = strconv.FormatInt(versionDetails.CommittedAt.Unix(), 10)
	}

	return &versionCollector{
		commitedAtEpoch: commitedAtEpoch,
		isDirty:         strconv.FormatBool(versionDetails.DirtyBuild),
		revision:        versionDetails.Revision,
		version:         versionDetails.Short,

		versionDesc: prometheus.NewDesc(
			prometheus.BuildFQName(metricsPrefix, "", "version"),
			"Build Version",
			[]string{"commited_at_seconds", "is_dirty", "revision", "version"},
			nil,
		),
	}
}

// Describe publishes the description of each version metric to a metrics
// channel.
func (vc *versionCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- vc.versionDesc
}

// Collect returns version metrics.
func (vc *versionCollector) Collect(ch chan<- prometheus.Metric) {
	ch <- prometheus.MustNewConstMetric(
		vc.versionDesc,
		prometheus.UntypedValue,
		1,
		vc.commitedAtEpoch,
		vc.isDirty,
		vc.revision,
		vc.version,
	)
}