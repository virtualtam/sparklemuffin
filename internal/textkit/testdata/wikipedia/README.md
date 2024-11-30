# `textkit` - Wikipedia test data

The files used in unitary tests are plain-text extracts from Wikipedia articles:

- `Downtown_Seattle_Transit_Tunnel.txt`, extracted from [Downtown Seattle Transit Tunnel](https://en.wikipedia.org/wiki/Downtown_Seattle_Transit_Tunnel)
- `History_of_aluminium.txt`, extracted from [History of aluminium](https://en.wikipedia.org/wiki/History_of_aluminium)

These articles are released under the [Creative Commons Attribution-Share-Alike License 4.0](https://creativecommons.org/licenses/by-sa/4.0/).

They were chosen as test input from Wikipedia's [featured articles page](https://en.wikipedia.org/wiki/Wikipedia:Featured_articles)
at the time the tests were added.

## Downloading a plain-text extract of a Wikipedia article

Get a JSON document containing the article metadata and a plain-text extractfrom the Wikipedia API:

```shell
$ curl \
    'https://en.wikipedia.org/w/api.php?action=query&format=json&titles=Downtown_Seattle_Transit_Tunnel&prop=extracts&explaintext' \
    | jq -r '.query.pages.[].extract' \
    > Downtown_Seattle_Transit_Tunnel.txt
```

## Resources
- [MediaWiki - Extension:TextExtracts](https://www.mediawiki.org/wiki/Extension:TextExtracts)
- [MediaWiki - API:Parsing wikitext](https://www.mediawiki.org/wiki/API:Parsing_wikitext)
- [How to get plain text out of Wikipedia](https://stackoverflow.com/questions/4452102/how-to-get-plain-text-out-of-wikipedia)
- [Wikipedia text download](https://stackoverflow.com/questions/2683506/wikipedia-text-download)
