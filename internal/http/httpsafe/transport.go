// Copyright VirtualTam 2022, 2026
// SPDX-License-Identifier: MIT

package httpsafe

import (
	"fmt"
	"net/http"
)

// NewSafeTransport returns an http.Transport that refuses to connect to
// blocked destination addresses (see isBlockedIP), guarding every request
// made through it against Server-Side Request Forgery.
//
// Proxy support is deliberately disabled: DialContext only ever sees the
// address actually being connected to, which -- when a proxy is configured
// -- is the proxy's address, not the request's destination. Checking the
// proxy's address against the blocklist would be both wrong (a trusted
// proxy commonly lives on a private address) and pointless (it does nothing
// to constrain where the proxy then relays the request), so this transport
// always dials the origin server directly.
func NewSafeTransport() (*http.Transport, error) {
	baseTransport, ok := http.DefaultTransport.(*http.Transport)
	if !ok {
		return nil, fmt.Errorf("httpsafe: http.DefaultTransport is not *http.Transport (got %T)", http.DefaultTransport)
	}

	transport := baseTransport.Clone()
	transport.Proxy = nil
	transport.DialContext = dialContext

	return transport, nil
}
