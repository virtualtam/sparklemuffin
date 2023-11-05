# Documentation
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

## Building the documentation
Build the documentation with:

```shell
$ make docs

mdbook build docs
2023-11-05 16:19:04 [INFO] (mdbook::book): Book building has started
2023-11-05 16:19:04 [INFO] (mdbook::book): Running the html backend
```

The generated website will be located under `docs/book`.


## Building and serving the documentation (live-reload)
Build ans serve the documentation with:

```shell
$ make live-docs

mdbook serve docs
2023-11-05 16:19:25 [INFO] (mdbook::book): Book building has started
2023-11-05 16:19:25 [INFO] (mdbook::book): Running the html backend
2023-11-05 16:19:25 [INFO] (mdbook::cmd::serve): Serving on: http://localhost:3000
2023-11-05 16:19:25 [INFO] (mdbook::cmd::watch): Listening for changes...
2023-11-05 16:19:25 [INFO] (warp::server): Server::run; addr=[::1]:3000
2023-11-05 16:19:25 [INFO] (warp::server): listening on http://[::1]:3000
```

The generated website will be located under `docs/book`, and the live server can be
accessed by opening [http://localhost:3000](http://localhost:3000) in a Web browser.


## Reference
- [mdbook build](https://rust-lang.github.io/mdBook/cli/build.html) command
- [mdbook serve](https://rust-lang.github.io/mdBook/cli/serve.html) command
- [SUMMARY.md](https://rust-lang.github.io/mdBook/format/summary.html)
- [mdBook Configuration](https://rust-lang.github.io/mdBook/format/configuration/index.html)
- [Markdown](https://rust-lang.github.io/mdBook/format/markdown.html)
