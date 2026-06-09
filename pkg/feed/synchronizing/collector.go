// Copyright VirtualTam 2022, 2026
// SPDX-License-Identifier: MIT

package synchronizing

import (
	"github.com/prometheus/client_golang/prometheus"
)

const (
	labelFeedNotModified = "not-modified"
	labelFeedHashesMatch = "hashes-match"

	labelErrorTypeList           = "list"
	labelErrorTypeFetch          = "fetch"
	labelErrorTypeUpdateMetadata = "update-metadata"
	labelErrorTypeUpdateEntries  = "update-entries"
)

var _ prometheus.Collector = &Collector{}

// Collector tracks feed synchronization metrics.
type Collector struct {
	tasksTotal    prometheus.Counter
	durationTotal prometheus.Counter
	updatedFeeds  prometheus.Counter
	skippedFeeds  *prometheus.CounterVec
	bytesTotal    prometheus.Counter
	entriesTotal  prometheus.Counter
	errorsTotal   *prometheus.CounterVec
}

// NewCollector initializes and returns a new Collector for feed synchronization metrics.
func NewCollector(metricsPrefix string) *Collector {
	const (
		subsystem = "feed_sync"
	)

	tasksTotal := prometheus.NewCounter(
		prometheus.CounterOpts{
			Namespace: metricsPrefix,
			Subsystem: subsystem,
			Name:      "tasks_total",
			Help:      "Number of scheduled feed synchronization tasks.",
		},
	)

	durationTotal := prometheus.NewCounter(
		prometheus.CounterOpts{
			Namespace: metricsPrefix,
			Subsystem: subsystem,
			Name:      "duration_milliseconds_total",
			Help:      "Total task duration, in milliseconds.",
		},
	)

	updatedFeeds := prometheus.NewCounter(
		prometheus.CounterOpts{
			Namespace: metricsPrefix,
			Subsystem: subsystem,
			Name:      "feeds_updated_total",
			Help:      "Number of updated feeds.",
		},
	)
	skippedFeeds := prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: metricsPrefix,
			Subsystem: subsystem,
			Name:      "feeds_skipped_total",
			Help:      "Number of skipped feeds.",
		},
		[]string{"reason"},
	)
	skippedFeeds.WithLabelValues(labelFeedNotModified)
	skippedFeeds.WithLabelValues(labelFeedHashesMatch)

	bytesTotal := prometheus.NewCounter(
		prometheus.CounterOpts{
			Namespace: metricsPrefix,
			Subsystem: subsystem,
			Name:      "bytes_total",
			Help:      "Total size of feed data fetched from remote servers, in bytes.",
		},
	)

	entriesTotal := prometheus.NewCounter(
		prometheus.CounterOpts{
			Namespace: metricsPrefix,
			Subsystem: subsystem,
			Name:      "entries_total",
			Help:      "Number of synchronized feed entries.",
		},
	)

	errorsTotal := prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: metricsPrefix,
			Subsystem: subsystem,
			Name:      "errors_total",
			Help:      "Total number of feed synchronization errors, by pipeline stage.",
		},
		[]string{"type"},
	)
	errorsTotal.WithLabelValues(labelErrorTypeList)
	errorsTotal.WithLabelValues(labelErrorTypeFetch)
	errorsTotal.WithLabelValues(labelErrorTypeUpdateMetadata)
	errorsTotal.WithLabelValues(labelErrorTypeUpdateEntries)

	return &Collector{
		tasksTotal:    tasksTotal,
		durationTotal: durationTotal,
		updatedFeeds:  updatedFeeds,
		skippedFeeds:  skippedFeeds,
		bytesTotal:    bytesTotal,
		entriesTotal:  entriesTotal,
		errorsTotal:   errorsTotal,
	}
}

func (c *Collector) Collect(ch chan<- prometheus.Metric) {
	c.tasksTotal.Collect(ch)
	c.durationTotal.Collect(ch)
	c.updatedFeeds.Collect(ch)
	c.skippedFeeds.Collect(ch)
	c.bytesTotal.Collect(ch)
	c.entriesTotal.Collect(ch)
	c.errorsTotal.Collect(ch)
}

func (c *Collector) Describe(ch chan<- *prometheus.Desc) {
	prometheus.DescribeByCollect(c, ch)
}
