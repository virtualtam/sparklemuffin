// Copyright (c) VirtualTam
// SPDX-License-Identifier: MIT

// Package fetching provides an HTTP Client to fetch syndication feeds from remote servers.
//
// The Client performs HTTP conditional requests, and leverages the following HTTP headers;
// - ETag (response) / If-None-Match (request)
// - Last-Modified (response) / If-Modified-Since (request)
//
// See:
// - https://www.rfc-editor.org/rfc/rfc9110 - HTTP Semantics
// - https://http.dev/conditional-requests
// - https://rednafi.com/misc/etag_and_http_caching/
// - https://rachelbythebay.com/w/2022/03/07/get/
// - https://rachelbythebay.com/w/2024/05/27/feed/
// - https://stackoverflow.com/questions/824152/what-takes-precedence-the-etag-or-last-modified-http-header
package fetching
