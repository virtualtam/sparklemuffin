# Netscape Bookmark Parser
The [Netscape Bookmark File Format](https://learn.microsoft.com/en-us/previous-versions/windows/internet-explorer/ie-developer/platform-apis/aa753582(v=vs.85))
is a format commonly used by Web browsers and Web bookmarking applications to export
and import bookmarks.

It has a very loose specification (i.e. no
[DTD](https://en.wikipedia.org/wiki/Document_type_definition)
nor [XSL Stylesheet](https://en.wikipedia.org/wiki/XSL)),
and may be assimilated to XML "with some quirks":

- some elements have an opening tag, but no closing tag:
    - `<DT>` items;
    - `<DD>` item descriptions;
- some elements have an opening and closing tag:
    - `<A>...</A>` bookmarks;
    - `<H1>...</H1>` export title;
    - `<H3>...</H3>` folder name;
- some elements have *surprising* opening and closing tags:
    - `<DL><p>...</DL><p>` item lists;
- depending on the implementation:
    - elements may (or may not) be capitalized;
    - some elements may (or may not) be nested;
    - some attributes may (or may not) be present.


## Example Netscape Bookmark Document
```xml
<!DOCTYPE NETSCAPE-Bookmark-file-1>
<!-- This is an automatically generated file.
     It will be read and overwritten.
     DO NOT EDIT! -->
<TITLE>Bookmarks</TITLE>
<H1>Bookmarks</H1>
<DL><p>
    <DT><H3 ADD_DATE="1622567473" LAST_MODIFIED="1627855786">Favorites</H3>
    <DD>Add bookmarks here
    <DL><p>
        <DT><A HREF="https://domain.tld" ADD_DATE="1641057073" PRIVATE="1">Test Domain</A>
        <DT><A HREF="https://test.domain.tld" ADD_DATE="1641057073" LAST_MODIFIED="1646172586" PRIVATE="1">Test Domain II</A>
        <DD>Second test
    </DL><p>
</DL><p>
```


## Go Parser
SparkleMuffin uses [virtualtam/netscape-go](https://github.com/virtualtam/netscape-go)
to parse (unmarshal) and export (marshal) bookmarks using the Netscape Bookmark File Format.

This allows users to import or synchronize their existing bookmarks to SparkleMuffin,
and to export them for usage with another bookmarking service.

[virtualtam/netscape-go](https://github.com/virtualtam/netscape-go)
is provided as a standalone library in the hope other users may find it useful.


It leverages:

- the streaming parser abilities of Go's [encoding/xml](https://pkg.go.dev/encoding/xml) package for most of the heavy lifing;
- the HTML character escaping andunescaping of Go's [html](https://pkg.go.dev/html) package;
- previous work on [Shaarli](https://github.com/shaarli/Shaarli)'s
  [netscape-bookmark-parser](https://github.com/shaarli/netscape-bookmark-parser),
  especially its [test fixtures](https://github.com/shaarli/netscape-bookmark-parser/tree/master/tests/Fixtures/Encoder/input).

