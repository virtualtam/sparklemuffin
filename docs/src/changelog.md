# Change Log

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/) and this
project adheres to [Semantic Versioning](https://semver.org/).

## [v0.1.0](https://github.com/virtualtam/sparklemuffin/releases/tag/v0.1.0) - 2024-01-26
_Initial release_

### Added
#### Main Features
- Create and manage users of the application
- Create and manage bookmarks to Web pages (links)
- Display bookmarks and bookmark tags
- Import existing bookmarks
- Search bookmarks by keywords (full-text search)

#### Command-line & configuration
- Add a `sparklemuffin` root command to handle common program configuration
- Add a `createadmin` subcommand to create users with administrator privileges
- Add a `migrate` subcommand to manage database migrations
- Add a `run` subcommand to start all services
- Add a `version` subcommand to display the running version (featuring Git version information)
- Allow configuring services via:
    - application defaults
    - configuration file
    - command-line flags
    - environment variables

#### Observability
- Setup structured logging (formats: console, JSON)
- Expose Prometheus metrics:
    - Go runtime
    - HTTP requests
    - Application build and version information

#### Packaging & automation
- Package the application as a Docker container
- Provide Docker Compose configuration for:
    - local development
    - example usage
- Add mdBook documentation
- Add Make targets to:
    - run static analysis tools (linters)
    - run unitary tests
    - run integration tests
    - generate coverage reports
    - run live-reload development servers
    - build HTML documentation
- Add Github Actions workflows:
    - CI: build the application, run linters, run tests
    - Copywrite: ensure license headers are present in source files
    - Docker: build and publish Docker image to the Github Container Registry (GHCR)
