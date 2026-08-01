# Documentation Structure
## Markdown sources
[mdBook](https://rust-lang.github.io/mdBook/) generates the documentation as
a static Website from
[Markdown](https://rust-lang.github.io/mdBook/format/markdown.html) files.

The SparkleMuffin repository stores the documentation resources under the
`docs/` directory:

```shell
docs/
├── book       # Generated Website (not tracked in Git)
├── book.toml  # mdBook configuration
└── src        # Markdown source files
```

## Sections and page categories

The documentation has two main sections:

- a [User Guide](../../user-guide/index.md) that shows SparkleMuffin's features and how to use them;
- a [Developer Guide](../../developer-guide/index.md) that explains how SparkleMuffin works,
  and how to contribute to the project.

The [Diátaxis](https://diataxis.fr/) approach organizes pages into four categories:

- **Tutorials**: learning-oriented lessons that take you through a series of steps to use a feature;
- **How-to Guides**: practical step-by-step guides to help you achieve a specific goal;
- **Reference Guides**: details about how SparkleMuffin works;
- **Concept Guides**: thoughts and reflections about why things work the way they do.

## Reference
- [mdbook build](https://rust-lang.github.io/mdBook/cli/build.html) command
- [mdbook serve](https://rust-lang.github.io/mdBook/cli/serve.html) command
- [SUMMARY.md](https://rust-lang.github.io/mdBook/format/summary.html)
- [mdBook Configuration](https://rust-lang.github.io/mdBook/format/configuration/index.html)
- [Markdown](https://rust-lang.github.io/mdBook/format/markdown.html)
