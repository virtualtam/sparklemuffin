# Static Analysis
## Dependencies
- [GNU Make](https://www.gnu.org/software/make/)
- [copywrite](https://github.com/hashicorp/copywrite)
- [golangci-lint](https://github.com/golangci/golangci-lint)
- [govulncheck](https://go.dev/blog/vuln)
- [SQLFluff](https://docs.sqlfluff.com/en/stable/index.html)

## Install development utilities
Install Go linters, vulnerability detection and license check tools:

```shell
$ make dev-install-tools
```

Install SQLFluff:

```shell
$ make dev-install-sqlfluff
```

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
