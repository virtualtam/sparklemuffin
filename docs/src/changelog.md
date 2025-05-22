# Change Log

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/) and this
project adheres to [Semantic Versioning](https://semver.org/).

## [v0.5.0](https://github.com/virtualtam/sparklemuffin/releases/tag/v0.5.0) - UNRELEASED
### Security
- Bump `golang.org/x/crypto` to v0.35.0:
    - [Vulnerability Report: GO-2025-3487](https://pkg.go.dev/vuln/GO-2025-3487)
    - [CVE-2025-22869](https://www.cve.org/CVERecord?id=CVE-2025-22869)

### Added
#### Bookmarks
- Export as a JSON document

### Changed
#### WWW
- Log pagination errors as warnings

#### Packaging and automation
- Build with Go 1.24
- Update CI workflow
- Update direct and transitive dependencies
- Run the `modernize` analyzer and apply fixes
- Bump golangci-lint to v2
- Setup golangci-lint to run additional linters

### Removed
#### Feed
- Remove unused `importing.Repository` type


## [v0.4.3](https://github.com/virtualtam/sparklemuffin/releases/tag/v0.4.3) - 2025-01-05
### Fixed
#### Feeds
- Ensure truncating entry descriptions does not result in invalid UTF-8 code points


## [v0.4.2](https://github.com/virtualtam/sparklemuffin/releases/tag/v0.4.2) - 2024-12-21
### Changed
#### Feeds
- When deleting a category or subscription, propagate the deletion to feeds with no remaining subscriptions

### Security
- Bump `golang/x/net` to v0.33.0:
    - [CVE-2024-45338](https://nvd.nist.gov/vuln/detail/CVE-2024-45338)
    - [Vulnerability in golang.org/x/net](https://groups.google.com/g/golang-announce/c/wSCRmFnNmPA/m/Lvcd0mRMAwAJ)


## [v0.4.1](https://github.com/virtualtam/sparklemuffin/releases/tag/v0.4.1) - 2024-12-14
### Security
- Bump `golang.org/x/crypto` to v0.31.0:
    - [CVE-2024-45337](https://nvd.nist.gov/vuln/detail/CVE-2024-45337)
    - [Vulnerability in golang.org/x/crypto](https://groups.google.com/g/golang-announce/c/-nPEi39gI4Q)


## [v0.4.0](https://github.com/virtualtam/sparklemuffin/releases/tag/v0.4.0) - 2024-12-10
### Added
#### Database
- Add PostgreSQL integration tests for feed operations

### Changed
#### Database
- Split PostgreSQL repository into dedicated domain repositories
- Update testcontainers configuration to use a tmpfs volume and disable WAL features to speed up integration tests

#### Feeds
- Update page title to display the subscription alias (if set) or the feed title
- Update listed entries to display the subscription alias (if set) or the feed title

#### WWW
- Relocate HTTP packages to `internal/http`
- Relocate version detection helpers to `internal/version`


## [v0.3.1](https://github.com/virtualtam/sparklemuffin/releases/tag/v0.3.1) - 2024-12-07
### Fixed
#### Feeds
- Fix HTML templates after renaming querying models


## [v0.3.0](https://github.com/virtualtam/sparklemuffin/releases/tag/v0.3.0) - 2024-12-07
### Added
#### Database
- Add dedicated tests for PostrgeSQL database migrations (up/down)

#### Documentation
- Add custom CSS to display wider content on large screens

#### Feeds
- For each entry in the list, display the title of the corresponding feed
- Save and display feed descriptions
- Extract keywords (significant terms) from entry content/description with TextRank
- Add full-text search based on feed and entry metadata
- Store and compare the hash (xxHash64) of the feed data to avoid unnecessary database upserts
- Document the feed polling and caching strategy
- Allow users to set an alias title for feed subscriptions

### Changed
### CI
- Lint and format SQL files with SQLFluff
- Publish HTML documentation to Github Pages

### Documentation
- Update documentation structure to follow the Diátaxis approach
- Disable mdBook file auto-creation
- Check for broken links with mdbook-linkcheck

#### www
- Update the home page
- Render HTTP 4xx errors as HTML views

### Fixed
#### Docker
- Install the `ca-certificates` package in the Docker image for TLS connections

#### Feeds
- Ensure entry publication and update dates are non-zero
- Ensure entry publication and update dates are not in the (far) future (limit: 2 days)
- In the subscription edit form, ensure the correct category is selected


## [v0.2.0](https://github.com/virtualtam/sparklemuffin/releases/tag/v0.2.0) - 2024-11-14
### Added
#### Feeds
- Subscribe to Atom and RSS feeds
- Categorize subscriptions
- Display subscriptions
- Periodically synchronize subscriptions
- Import existing subscriptions from OPML
- Export subscriptions

### Changed
#### Bookmarks
- Enforce CSRF validation for import and export forms

#### PostgreSQL
- Update repository helpers

#### Packaging & automation
- Build with Go 1.23
- Update direct and transitive dependencies


## [v0.1.1](https://github.com/virtualtam/sparklemuffin/releases/tag/v0.1.1) - 2024-01-26
_Initial release_

### Added
#### Bookmarks
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
