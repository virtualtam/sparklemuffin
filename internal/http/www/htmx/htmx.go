// Copyright VirtualTam 2022, 2026
// SPDX-License-Identifier: MIT

// Package htmx provides helpers to integrate the HTMX Javascript library.
package htmx

const (
	// HTMX response headers
	// See https://htmx.org/reference/#response_headers

	// HeaderRefresh indicates to the client-side whether to do a full refresh of the page.
	HeaderRefresh = "HX-Refresh"

	// HeaderRedirect instructs the client-side to do a full client-side
	// redirect (window.location) rather than swapping the response into the
	// requesting element's target.
	HeaderRedirect = "HX-Redirect"

	// HTMX request headers
	// See https://htmx.org/reference/#request_headers

	// HeaderRequest is set to "true" by htmx on every request it issues,
	// letting the server tell an htmx-initiated request apart from a plain
	// browser request.
	HeaderRequest = "HX-Request"
)
