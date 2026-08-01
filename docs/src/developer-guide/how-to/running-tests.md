# Running tests
## Dependencies
- [GNU Make](https://www.gnu.org/software/make/)
- [Docker](https://docs.docker.com/) for integration tests with [Testcontainers](https://testcontainers.com/)


## Run tests
Run unit and integration tests:

```shell
$ make test
```

Run unit and integration tests with race detection enabled:

```shell
$ make race
```

## Code coverage reports
Run unit and integration tests with code coverage enabled:

```shell
$ make cover
```

Generate an HTML report. Open it in your Web browser:

```shell
$ make coverhtml
```
