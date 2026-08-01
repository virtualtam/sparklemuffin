# Continuous Integration
## GitHub Actions Workflows
Each Action pin uses a full commit SHA. A trailing comment shows the
released version, for example `actions/checkout@<sha> # v7.0.1`. Someone
can move a tag to point at different code. A commit SHA cannot change.
SHA-pinning protects workflows against a compromised or mistakenly
re-tagged Action release.

### CI Workflow
This workflow runs when:

- new commits are pushed to the `main` Git branch;
- new Git tags are pushed;
- Pull Requests are created or updated.

It runs three jobs. The Build job compiles SparkleMuffin. The Lint job
checks Go and SQL sources. The Test job runs the Go unit and integration
tests.

See [Compiling](../how-to/compiling.md),
[Running Static Analysis](../how-to/running-static-analysis.md), and
[Running Tests](../how-to/running-tests.md).

### Copywrite Workflow
This workflow runs when:

- new commits are pushed to the `main` Git branch;
- new Git tags are pushed;
- Pull Requests are created or updated.

It checks that Go source files have a valid license header. It uses
[copywrite](https://github.com/hashicorp/copywrite).

See [Running Static Analysis](../how-to/running-static-analysis.md).

### Vulnerabilities Workflow
This workflow runs when:

- new commits are pushed to the `main` Git branch;
- new Git tags are pushed;
- Pull Requests are created or updated.

It checks Go sources and `go.mod` for known vulnerabilities. It uses
[govulncheck](https://go.dev/blog/vuln).

See [Running Static Analysis](../how-to/running-static-analysis.md).

### Documentation Workflow
This workflow runs when:

- new commits are pushed to the `main` Git branch;
- new Git tags are pushed;
- Pull Requests are created or updated.

It generates the HTML documentation with `mdBook`. Then it checks the
documentation for broken links with [lychee](https://lychee.cli.rs/).

When new commits reach the `main` Git branch and the `CI` workflow succeeds,
this workflow uploads the documentation to GitHub Pages. Find it at
[SparkleMuffin Documentation](https://virtualtam.github.io/sparklemuffin/).

See [Generating Documentation](../how-to/generating-documentation.md).

### Docker Workflow
This workflow runs when:

- new commits are pushed to the `main` Git branch;
- new Git tags are pushed.

It builds and tags the SparkleMuffin production Docker images. It pushes them
to the GitHub Container Registry (GHCR) at
[ghcr.io/virtualtam/sparklemuffin](https://github.com/virtualtam/sparklemuffin/pkgs/container/sparklemuffin).

## Dependabot
Dependabot opens a Pull Request once a month to update:

- GitHub Actions, grouped into a single Pull Request;
- Go modules;
- Node.js packages, for the frontend asset pipeline;
- SQLFluff, pinned as a `uv` dependency in
  [`internal/repository/pyproject.toml`](https://github.com/virtualtam/sparklemuffin/blob/main/internal/repository/pyproject.toml).
