# Documentation Structure
## Markdown sources
The documentation is a static Website generated from
[Markdown](https://rust-lang.github.io/mdBook/format/markdown.html) files
using [mdBook](https://rust-lang.github.io/mdBook/).

The documentation resources are part of the SparkleMuffin repository, and located
under the `docs/` directory:

```shell
docs/
├── book       # Generated Website (not tracked in Git)
├── book.toml  # mdBook configuration
└── src        # Markdown source files
```

## Structure and sections

The documentation follows the [Diátaxis](https://diataxis.org/) approach.

It is split in two main sections:

- a [User Guide](./user-guide/index.md) that showcases SparkleMuffin's features and how to use them;
- a [Developer Guide](./developer-guide/index.md) that provides information on how SparkleMuffin works,
  and how to contribute to the project.

Pages are then split into four categories:

- Tutorials: learning-oriented lessons that take you through a series of steps to use a feature;
- How-to Guides: practical step-by-step guides to help you achieve a specific goal;
- Reference Guides: details about how SparkleMuffin works;
- Concept Guides: thoughts and reflections about how why things work the way they do.

## Reference
- [mdbook build](https://rust-lang.github.io/mdBook/cli/build.html) command
- [mdbook serve](https://rust-lang.github.io/mdBook/cli/serve.html) command
- [SUMMARY.md](https://rust-lang.github.io/mdBook/format/summary.html)
- [mdBook Configuration](https://rust-lang.github.io/mdBook/format/configuration/index.html)
- [Markdown](https://rust-lang.github.io/mdBook/format/markdown.html)
