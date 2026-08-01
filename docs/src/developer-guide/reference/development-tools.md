# Development Tools

## Git
SparkleMuffin uses [Git](https://git-scm.com/) to track source code changes.
The code is available on GitHub at
[github.com/virtualtam/sparklemuffin](https://github.com/virtualtam/sparklemuffin).

To start using Git and GitHub:

- [Get started with GitHub](https://docs.github.com/en/get-started)
- [First-Time Git Setup](https://git-scm.com/book/en/v2/Getting-Started-First-Time-Git-Setup)


## Go
SparkleMuffin is written mainly in the [Go programming language](https://go.dev/).

See [`go.mod`](https://github.com/virtualtam/sparklemuffin/blob/main/go.mod) for the
minimum Go version that SparkleMuffin needs.

### Linux
Install Go with your Linux distribution's package manager.

### macOS
Install Go with the [Homebrew](https://brew.sh/) community packages:

```shell
$ brew install go
```

### Windows
Install Go with [winget](https://github.com/microsoft/winget-cli):

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

Install the current Long-Term Support (LTS) version of Node.js.

### Linux
Install Node.js with your Linux distribution's package manager.

### macOS
Install Node.js with the [Homebrew](https://brew.sh/) community packages:

```shell
$ brew install node
```

### Windows
Install Node.js with [winget](https://github.com/microsoft/winget-cli):

```shell
$ winget install --id=OpenJS.NodeJS
```

## Docker
SparkleMuffin uses [Docker](https://docs.docker.com/) to:

- Package the application as easy-to-run Docker images;
- Run database integration tests with [Testcontainers](https://testcontainers.com/);
- Spin a local development environment with [Docker Compose](https://docs.docker.com/compose/)


Building Docker images locally needs a recent version of Docker. Local builds use:

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
Install uv with the
[standalone installer](https://docs.astral.sh/uv/getting-started/installation/):

```shell
$ curl -LsSf https://astral.sh/uv/install.sh | sh
```

### Windows
Install uv with [winget](https://github.com/microsoft/winget-cli):

```shell
$ winget install --id=astral-sh.uv
```

## GNU Make
A [Makefile](https://www.gnu.org/software/make/) provides targets to run tests
and linters, generate documentation, and start local development environments.

## lychee
SparkleMuffin uses [lychee](https://lychee.cli.rs/) to check the generated HTML
documentation for broken links.

## mdBook
SparkleMuffin uses [mdBook](https://rust-lang.github.io/mdBook/) to generate a
static HTML documentation from [Markdown](https://rust-lang.github.io/mdBook/format/markdown.html) files.

## Watchexec
SparkleMuffin uses [watchexec](https://github.com/watchexec/watchexec) to
live-reload the development server when source files change on disk.
