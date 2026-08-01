# Development Tools

## Git
The source code is tracked using the [Git](https://git-scm.com/) Source Code Management
tool, and available on GitHub at
[github.com/virtualtam/sparklemuffin](https://github.com/virtualtam/sparklemuffin).

To get started with using Git and GitHub:

- [Get started with GitHub](https://docs.github.com/en/get-started)
- [First-Time Git Setup](https://git-scm.com/book/en/v2/Getting-Started-First-Time-Git-Setup)


## Go
SparkleMuffin is mainly written with the [Go programming language](https://go.dev/).

See [`go.mod`](https://github.com/virtualtam/sparklemuffin/blob/main/go.mod) for the
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

## Node.js
SparkleMuffin uses the [Node.js runtime](https://nodejs.org/) to build its frontend assets.

We recommend installing the current Long-Term Support (LTS) version of Node.js.

### Linux
The recommended way of installing Node.js is via your Linux distribution's package manager.

### macOS
The recommended way of installing Node.js is via the [Homebrew](https://brew.sh/)
community packages:

```shell
$ brew install node
```

### Windows
The recommended way of installing Node.js is via [winget](https://github.com/microsoft/winget-cli):

```shell
$ winget install --id=OpenJS.NodeJS
```

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


## Python and uv
SparkleMuffin uses [SQLFluff](https://sqlfluff.com/) to lint and format SQL
migration files. SQLFluff is a Python tool, managed with
[uv](https://docs.astral.sh/uv/).

See [`internal/repository/pyproject.toml`](https://github.com/virtualtam/sparklemuffin/blob/main/internal/repository/pyproject.toml)
for the pinned SQLFluff version.

### Linux and macOS
The recommended way of installing uv is via the
[standalone installer](https://docs.astral.sh/uv/getting-started/installation/):

```shell
$ curl -LsSf https://astral.sh/uv/install.sh | sh
```

### Windows
The recommended way of installing uv is via [winget](https://github.com/microsoft/winget-cli):

```shell
$ winget install --id=astral-sh.uv
```

## GNU Make
A [Makefile](https://www.gnu.org/software/make/) is provided for convenience to help
running tests, linters, generate documentation and spin local development environments.

## lychee
[lychee](https://lychee.cli.rs/) is used to check the generated HTML documentation for broken links.

## mdBook
[mdBook](https://rust-lang.github.io/mdBook/) is used to generate a static HTML documentation from [Markdown](https://rust-lang.github.io/mdBook/format/markdown.html) files.

## Watchexec
[watchexec](https://github.com/watchexec/watchexec) is used to live-reload the development server when source files
have been changed on the disk.
