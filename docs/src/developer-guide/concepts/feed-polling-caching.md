# Feed polling and caching

SparkleMuffin periodically makes HTTP requests to update Atom and RSS feeds.
This creates two requirements:

- do not put unnecessary load on the remote servers;
- do not perform unnecessary database updates when the remote content has not changed.

SparkleMuffin uses HTTP caching features from the HTTP specification. It also
performs extra checks on the feed content.

## HTTP Conditional Requests

When responding to an HTTP request, a remote server may set the following headers:

- `ETag`: the current entity tag for the selected representation (usually a hash of the feed data);
- `Last-Modified`: the date and time the origin server last modified the selected representation.

When present, SparkleMuffin stores these values in the database. It uses them
to set the following headers in later requests:

- `If-None-Match`: the value of the `ETag` header from the previous response;
- `If-Modified-Since`: the value of the `Last-Modified` header from the previous response.

The remote server responds in one of two ways:

- `200 OK`: the content changed. SparkleMuffin updates the feed and its entries;
- `304 Not Modified`: nothing changed. SparkleMuffin only updates the feed's `ETag` and `Last-Modified` headers.

## Feed content hash
A remote server can send a different `ETag` or `Last-Modified` value even
when the feed content has not changed. It can also send neither header. To
handle this, SparkleMuffin:

- computes and stores a hash of the feed data, using the [xxHash](https://xxhash.com/) non-cryptographic hash function;
- compares the hash of the feed data with the value stored in the database;
- returns early if the hashes match, to avoid unnecessary database updates.


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
