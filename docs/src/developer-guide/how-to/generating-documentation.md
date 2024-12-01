# Generating the HTML documentation
## Prerequisites
- Install [mdBook](https://rust-lang.github.io/mdBook/);
- Install [mdbook-linkcheck](https://github.com/Michael-F-Bryan/mdbook-linkcheck).

## HTML Documentation
Build the HTML documentation with:

```shell
$ make docs

mdbook build docs
2023-11-05 16:19:04 [INFO] (mdbook::book): Book building has started
2023-11-05 16:19:04 [INFO] (mdbook::book): Running the html backend
```

The generated website will be located under `docs/book/html`.


## Live-reload server
Start `mdBook`'s live-reload server with:

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

- The generated website will be located under `docs/book/html`;
- The live server can be accessed by opening [http://localhost:3000](http://localhost:3000) in a Web browser.

## Reference
- [SparkleMuffin Documentation Structure](../reference/documentation-structure.md)
- [mdbook build](https://rust-lang.github.io/mdBook/cli/build.html) command
- [mdbook serve](https://rust-lang.github.io/mdBook/cli/serve.html) command
- [SUMMARY.md](https://rust-lang.github.io/mdBook/format/summary.html)
- [mdBook Configuration](https://rust-lang.github.io/mdBook/format/configuration/index.html)
- [Markdown](https://rust-lang.github.io/mdBook/format/markdown.html)
