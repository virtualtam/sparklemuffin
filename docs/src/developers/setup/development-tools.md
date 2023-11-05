# Development Tools

## Git
The source code is tracked using the [Git](https://git-scm.com/) Source Code Management
tool, and available on Github at
[github.com/virtualtam/sparklemuffin](https://github.com/virtualtam/sparklemuffin).

To get started with using Git and Github:

- [Get started with Github](https://docs.github.com/en/get-started)
- [First-Time Git Setup](https://git-scm.com/book/en/v2/Getting-Started-First-Time-Git-Setup)


## Go
SparkleMuffin is mainly written with the [Go programming language](https://go.dev/).

See [go.mod](https://github.com/virtualtam/sparklemuffin/blob/main/go.mod) for the
minimum version of Go required by SparkleMuffin.

### Linux
The recommended way of installing Go is via your Linux distribution's package manager.

### macOS
The recommended way of installing Go is via the [Homebrew](https://brew.sh/)
community packages:

```shell
$ brew install go
```

### Windows
The recommended way of installing Go is via [winget](https://github.com/microsoft/winget-cli):

```shell
$ winget install --id=GoLang.Go
```

### Manual installation (advanced users)
To install a specific version of Go, see:

- [Download and Install](https://go.dev/doc/install) page;
- [Installing Go from sources](https://go.dev/doc/install/source)
- [Managing Go installations](https://go.dev/doc/manage-install)


## Docker
[Docker](https://docs.docker.com/) is used to:

- Package the application as easy-to-run Docker images;
- Run database integration tests with [Testcontainers](https://testcontainers.com/);
- Spin a local development environment with [Docker Compose](https://docs.docker.com/compose/)


A recent version of Docker is required to build Docker images locally, as we leverage:

- [Multi-stage builds](https://docs.docker.com/build/building/multi-stage/)
- [Local build cache volumes](https://docs.docker.com/build/cache/)
- The [buildx](https://docs.docker.com/engine/reference/commandline/buildx_build/)
  integration for [BuildKit](https://docs.docker.com/build/buildkit/)


## GNU Make
A [Makefile](https://www.gnu.org/software/make/) is provided for convenience to help
running tests, linters, generate documentation and spin local development environments.


## mdBook
[mdBook](https://rust-lang.github.io/mdBook/) is used to generate a static HTML documentation
from [Markdown](https://rust-lang.github.io/mdBook/format/markdown.html) files.


## Watchexec
[watchexec](https://github.com/watchexec/watchexec) is used to live-reload the development
server when source files have been changed on the disk.
