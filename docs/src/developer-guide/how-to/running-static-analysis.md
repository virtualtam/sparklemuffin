# Static Analysis
## Dependencies
- [GNU Make](https://www.gnu.org/software/make/)
- [copywrite](https://github.com/hashicorp/copywrite)
- [golangci-lint](https://github.com/golangci/golangci-lint)
- [govulncheck](https://go.dev/blog/vuln)
- [uv](https://docs.astral.sh/uv/)
- [SQLFluff](https://docs.sqlfluff.com/en/stable/index.html)

## Install development utilities
Install the Go linter, the vulnerability scanner, and the license checker:

```shell
$ make dev-install-tools
```

Install SQLFluff:

```shell
$ make dev-install-sqlfluff
```

This uses `uv` to create a virtual environment under
`internal/repository/.venv/`. The files `internal/repository/pyproject.toml`
and `internal/repository/uv.lock` set the exact package versions.

## Run linters
### Go
Check Go sources with golangci-lint:

```shell
$ make lint
```

Check Go source headers with copywrite:

```shell
$ make copywrite
```

Check Go sources and `go.mod` for vulnerabilities:

```shell
$ make vulncheck
```


### SQL Migrations
Check SQL files with SQLFluff:

```shell
$ make lint-sql
```

Format SQL files with SQLFluff:

```shell
$ make format-sql
```

Applied migrations must keep their statements unchanged. A newer
SQLFluff release can add a rule that an applied migration no longer
passes. In this case, add a `-- noqa: <RULE>` comment on the failing
line. Do not change the statement. New migrations must pass the
current rule set.
