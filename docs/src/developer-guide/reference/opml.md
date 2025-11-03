# OPML Feed Subscription Parser
[Outline Processor Markup Language (OPML)](https://opml.org/spec2.opml) is a format
commonly used by feed aggregators and feed readers to export and import subscriptions
to [Atom](https://en.wikipedia.org/wiki/Atom_(web_standard)) and
[RSS](https://en.wikipedia.org/wiki/RSS) feeds.

It has a permissive specification, and each feed aggregator or reader may:

- specify extra attributes;
- use non-standard attributes or attribute formats (e.g. to format dates and time);
- use a nested structure to represent subscriptions, categories and directories.

## Specifications

- [OPML 2.0 Specification](https://opml.org/spec2.opml)
- [OPML on Wikipedia](https://en.wikipedia.org/wiki/OPML)
- [OPML 2.0 Format Description by the Library of Congress](https://loc.gov/preservation/digital/formats/fdd/fdd000554.shtml)
- [scripting/opml.org - Issue 3 - Questions about grey-areas in the specification](https://github.com/scripting/opml.org/issues/3)
- [Mozilla - How to Subscribe to News Feeds and Blogs](https://support.mozilla.org/en-US/kb/how-subscribe-news-feeds-and-blogs)

## Example OPML Document
```xml
<?xml version="1.0" encoding="UTF-8"?>

<opml version="1.0">
    <head>
        <title>My subscriptions in feedly Cloud</title>
    </head>
    <body>
        <outline text="Programming" title="Programming">
            <outline type="rss" text="Elixir Lang" title="Elixir Lang" xmlUrl="https://feeds.feedburner.com/ElixirLang" htmlUrl="https://elixir-lang.org"/>
            <outline type="rss" text="Python Insider" title="Python Insider" xmlUrl="https://feeds.feedburner.com/PythonInsider" htmlUrl="https://pythoninsider.blogspot.com/"/>
        </outline>
        <outline text="Games" title="Games">
            <outline type="rss" text="Vintage Story" title="Vintage Story" xmlUrl="https://www.vintagestory.at/blog.html/?rss=1" htmlUrl="https://www.vintagestory.at/blog.html/"/>
        </outline>
    </body>
</opml>
```

## Go Parser
SparkleMuffin uses [virtualtam/opml-go](https://github.com/virtualtam/opml-go/)
to parse (unmarshal) and export (marshal) feed subscriptions using the OPML file format.

This allows users to import or synchronize their existing subscriptions to SparkleMuffin,
and to export them for usage with another feed aggregator or feed reader.

[virtualtam/opml-go](https://github.com/virtualtam/opml-go/)
is provided as a standalone library in the hope other users may find it useful.
