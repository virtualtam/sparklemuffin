# Observability
## Structured Logs
SparkleMuffin logs information on the standard error (stderr) stream,
using a structured log message format (JSON or logfmt).

The log level can be specified via [configuration](./configuration.md).

TODO: add command-line flag to configure the log format (JSON or logfmt).

### Example logs: program startup
```shell
2024-01-20T17:03:52+01:00 INF configuration: no file found config_paths=["/etc","/home/dev/.config","."]
2024-01-20T17:03:53+01:00 INF database: successfully created connection pool database_addr=localhost:15432 database_driver=pgx database_name=sparklemuffin
2024-01-20T17:03:53+01:00 INF global: setting up services log_level=info version=devel
2024-01-20T17:03:53+01:00 INF metrics: listening for HTTP requests metrics_addr=127.0.0.1:8081
2024-01-20T17:03:53+01:00 INF sparklemuffin: listening for HTTP requests http_addr=0.0.0.0:8080
2024-01-20T17:04:44+01:00 INF handle request duration_ms=0.750875 host=localhost:8080 method=GET path=/ remote_addr=127.0.0.1:51440 request_id=localhost.local/96bRV2ceWt-000001 size=1187 status=200
2024-01-20T17:04:44+01:00 INF handle request duration_ms=4.369792 host=localhost:8080 method=GET path=/static/awesomplete.css remote_addr=127.0.0.1:51441 request_id=localhost.local/96bRV2ceWt-000004 size=167 status=200
2024-01-20T17:04:44+01:00 INF handle request duration_ms=5.682958 host=localhost:8080 method=GET path=/static/easymde.css remote_addr=127.0.0.1:51442 request_id=localhost.local/96bRV2ceWt-000003 size=931 status=200
2024-01-20T17:04:44+01:00 INF handle request duration_ms=7.072792 host=localhost:8080 method=GET path=/static/www.css remote_addr=127.0.0.1:51440 request_id=localhost.local/96bRV2ceWt-000002 size=5402 status=200
```

## Prometheus Metrics
SparkleMuffin exposes [Prometheus metrics](https://prometheus.io/docs/concepts/metric_types/),
providing useful information that can be used for monitoring and alerting.

These metrics are exposed by default on `http://0.0.0.0:8081/metrics`; the host and port can be
specified via [configuration](./configuration.md).

### Available Metrics
- Go runtime metrics exposed by [prometheus/client_golang/prometheus](https://github.com/prometheus/client_golang/tree/main/prometheus);
- Go HTTP metrics exposed by [prometheus/client_golang/prometheus/promhttp](https://github.com/prometheus/client_golang/tree/main/prometheus/promhttp).
- SparkleMuffin build and version information.

- TODO: expose business information
- TODO: example Grafana dashboard
- TODO: example observability stack
