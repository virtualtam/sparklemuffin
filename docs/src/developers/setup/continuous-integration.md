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


## Local development
See:

- [Development Tools](./development-tools.md)
- [Running Static Analysis](./running-static-analysis.md)
- [Running Tests](./running-tests.md)
- [Compiling](./compiling.md)
- [Live Development Server](./live-development-server.md)
- [Documentation](./documentation.md)
