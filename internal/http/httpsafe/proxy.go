// Copyright VirtualTam 2022, 2026
// SPDX-License-Identifier: MIT

package httpsafe

import "golang.org/x/net/http/httpproxy"

// ProxyConfigured reports whether HTTP or HTTPS traffic is configured to be
// routed through a forward proxy, per the standard HTTP_PROXY / HTTPS_PROXY
// environment variables.
//
// It uses the same environment-parsing logic net/http's own
// http.ProxyFromEnvironment relies on, so it agrees with what an
// unconfigured http.Client would actually do.
//
// NewSafeTransport's guard is incompatible with a forward proxy (see its
// doc comment); callers should use ProxyConfigured to decide whether it's
// safe to use.
func ProxyConfigured() bool {
	cfg := httpproxy.FromEnvironment()
	return cfg.HTTPProxy != "" || cfg.HTTPSProxy != ""
}
