// Copyright (c) VirtualTam
// SPDX-License-Identifier: MIT

// Package monitoring provides utilities to set up a HTTP server to expose monitoring information
// such as application metrics and metadata.
package monitoring

import (
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"github.com/virtualtam/sparklemuffin/internal/version"
)

const (
	webroot = `<html>
<head><title>SparkleMuffin Monitoring</title></head>
<body>
  <h1>Monitoring</h1>
  <p><a href="/metrics">Metrics</a></p>
</body>
</html>`
)

// NewServer initializes a Prometheus metrics registry, registers metrics collectors
// and returns an HTTP server to expose them.
func NewServer(metricsPrefix string, metricsListenAddr string, versionDetails *version.Details) (*http.Server, *prometheus.Registry) {
	metricsRegistry := prometheus.NewRegistry()
	metricsRegistry.MustRegister(
		collectors.NewGoCollector(),
		newVersionCollector(metricsPrefix, versionDetails),
	)

	opts := promhttp.HandlerOpts{}

	router := http.NewServeMux()

	router.Handle("/metrics", promhttp.HandlerFor(metricsRegistry, opts))
	router.HandleFunc("/", func(w http.ResponseWriter, _ *http.Request) {
		_, err := w.Write([]byte(webroot))
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	})

	metricsServer := &http.Server{
		Addr:         metricsListenAddr,
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
	}

	return metricsServer, metricsRegistry
}
