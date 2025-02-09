# Running tests
## Dependencies
- [GNU Make](https://www.gnu.org/software/make/)
- [Docker](https://docs.docker.com/) for integration tests with [Testcontainers](https://testcontainers.com/)


## Run tests
Run unitary and integration tests:

```shell
$ make test
```

Run unitary and integration tests with race detection enabled:

```shell
$ make race
```

## Code coverage reports
Run unitary and integration tests with code coverage enabled:

```shell
$ make cover
```

Generate an HTML report and open it in your Web browser:

```shell
$ make coverhtml
```
