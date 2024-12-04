# Feed polling and caching

As SparkleMuffin periodically makes HTTP requests to update Atom and RSS feeds, we need to ensure:

- we do not put unnecessary load on the remote servers;
- we do not perform unnecessary database updates if the remote content has not changed.

To this effect, we leverage features from the HTTP specification to benefit from remote server caching,
and perform additional checks on the feed content.

## HTTP Conditional Requests

When responding to an HTTP request, a remote server may set the following headers:

- `ETag`: the current entity tag for the selected representation (usually a hash of the feed data));
- `Last-Modified`: a timestamp indicating the date and time at which the origin server believes
  the selected representation was last modified.

When present, we store these values in the database, and use them to set the following headers in subsequent
requests:

- `If-None-Match`: the value of the `ETag` header from the previous response;
- `If-Modified-Since`: the value of the `Last-Modified` header from the previous response.

Depending on whether the feed has changed since the last request, the remote server will then respond with:


- `200 OK`: the content has changed, we update the feed and its entries;
- `304 Not Modified`: there are no changes, we only update the feed's `ETag` and `Last-Modified` headers.

## Feed content hash
As a remote server may send a different `ETag` or `Last-Modified` value without the feed content being modified,
or not send any of these headers at all, we:

- compute and store a hash of the feed data using the [xxHash](https://xxhash.com/) non-cryptographic hash function;
- compare the hash of the feed data with what we already have in the database;
- return early if the hashes match, to avoid unnecessary database updates.


## Reference
### Feed caching
- [feed reader score project](https://rachelbythebay.com/fs/)
- [A sysadmin's rant about feed readers and crawlers](https://rachelbythebay.com/w/2022/03/07/get/)
- [Feeds, updates, 200s, 304s, and now 429s](https://rachelbythebay.com/w/2023/01/18/http/)
- [So many feed readers, so many bizarre behaviors](https://rachelbythebay.com/w/2024/05/27/feed/)
- [The feed reader score service is now online](https://rachelbythebay.com/w/2024/05/30/fs/)

### RFCs
- [RFC 7232 - Hypertext Transfer Protocol (HTTP/1.1) - Validators - Last-Modified](https://datatracker.ietf.org/doc/html/rfc7232#section-2.2)
- [RFC 7232 - Hypertext Transfer Protocol (HTTP/1.1):- Validators - ETag](https://datatracker.ietf.org/doc/html/rfc7232#section-2.3)
- [RFC 9110 - HTTP Semantics](https://www.rfc-editor.org/rfc/rfc9110)

### HTTP Conditional Requests
- [HTTP Conditional Requests Explained](https://http.dev/conditional-requests)
- [Bret Simmons - NetNewsWire and Conditional GET Issues](https://inessential.com/2024/08/03/netnewswire_and_conditional_get_issues.html)
- [John Brayton - Feed Polling for Unread Cloud](https://www.goldenhillsoftware.com/2024/08/feed-polling-for-unread-cloud/)
- [Jeff Kaufman - Looking at RSS User-Agents](https://www.jefftk.com/p/looking-at-rss-user-agents)
- [Chris Siebenmann - The case of the very old If-Modified-Since HTTP header](https://utcc.utoronto.ca/~cks/space/blog/web/VeryOldIfModifiedSince)
- [ETag and HTTP caching](https://rednafi.com/misc/etag_and_http_caching/)
- [Caching - What takes precedence: the ETag or Last-Modified HTTP header?](https://stackoverflow.com/questions/824152/what-takes-precedence-the-etag-or-last-modified-http-header)

### Non-cryptographic hash functions
- [xxHash](https://xxhash.com/), an extremely fast non-cryptographic hash algorithm
- [cespare/xxHash](https://github.com/cespare/xxhash) library for Go
