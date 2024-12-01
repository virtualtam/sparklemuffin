# Continuous Integration
## Github Actions Workflows
### CI Workflow
This workflow runs when:

- new commits are pushed to the `main` Git branch;
- new Git tags are pushed;
- Pull Requests are created or updated.

It runs all continuous integration tasks:

- Documentation build;
- SQL linter (static code analysis);
- Go linters (static code analysis);
- Go unitary and integration tests;
- Go build.

### Docker Workflow
This workflow runs when:

- new commits are pushed to the `main` Git branch;
- new Git tags are pushed.

It builds and tags the SparkleMuffin production Docker images, and pushes them to
the Github Container Registry (GHCR) at
[ghcr.io/virtualtam/sparklemuffin](https://github.com/virtualtam/sparklemuffin/pkgs/container/sparklemuffin).

### Documentation workflow
This workflow runs when:

- new commits are pushed to the `main` Git branch;
- new Git tags are pushed;
- Pull Requests are created or updated.

It generates the HTML documentation with `mdBook`.

When new commits are pushed to the `main` Git branch and the `CI` workflow is successful,
the documentation is uploaded to Github Pages and can be accessed here: [SparkleMuffin Documentation](https://virtualtam.github.io/sparklemuffin/).
