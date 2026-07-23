// Copyright VirtualTam 2022, 2026
// SPDX-License-Identifier: MIT

package httpsafe_test

import (
	"errors"
	"testing"

	"github.com/virtualtam/sparklemuffin/internal/http/httpsafe"
)

func TestCheckDestination(t *testing.T) {
	cases := []struct {
		tname   string
		rawURL  string
		wantErr error
	}{
		// blocked: loopback
		{tname: "IPv4 loopback", rawURL: "http://127.0.0.1/feed", wantErr: httpsafe.ErrIPBlocked},
		{tname: "IPv4 loopback (non-canonical)", rawURL: "http://127.1.2.3/feed", wantErr: httpsafe.ErrIPBlocked},
		{tname: "IPv6 loopback", rawURL: "http://[::1]/feed", wantErr: httpsafe.ErrIPBlocked},

		// blocked: link-local (includes the 169.254.169.254 cloud metadata endpoint)
		{tname: "IPv4 link-local", rawURL: "http://169.254.1.1/feed", wantErr: httpsafe.ErrIPBlocked},
		{tname: "IPv4 link-local (cloud metadata)", rawURL: "http://169.254.169.254/latest/meta-data/", wantErr: httpsafe.ErrIPBlocked},
		{tname: "IPv6 link-local", rawURL: "http://[fe80::1]/feed", wantErr: httpsafe.ErrIPBlocked},

		// blocked: private
		{tname: "IPv4 private (RFC1918 10/8)", rawURL: "http://10.0.0.5/feed", wantErr: httpsafe.ErrIPBlocked},
		{tname: "IPv4 private (RFC1918 172.16/12)", rawURL: "http://172.16.0.5/feed", wantErr: httpsafe.ErrIPBlocked},
		{tname: "IPv4 private (RFC1918 192.168/16)", rawURL: "http://192.168.1.5/feed", wantErr: httpsafe.ErrIPBlocked},
		{tname: "IPv6 unique local (RFC4193)", rawURL: "http://[fc00::1]/feed", wantErr: httpsafe.ErrIPBlocked},

		// blocked: carrier-grade NAT / shared address space (RFC6598) --
		// notably used by some cloud providers for instance metadata
		// endpoints, e.g. Alibaba Cloud's 100.100.100.200.
		{tname: "IPv4 shared address space (RFC6598, range start)", rawURL: "http://100.64.0.0/feed", wantErr: httpsafe.ErrIPBlocked},
		{tname: "IPv4 shared address space (RFC6598, cloud metadata)", rawURL: "http://100.100.100.200/feed", wantErr: httpsafe.ErrIPBlocked},
		{tname: "IPv4 shared address space (RFC6598, range end)", rawURL: "http://100.127.255.255/feed", wantErr: httpsafe.ErrIPBlocked},

		// blocked: unspecified
		{tname: "IPv4 unspecified", rawURL: "http://0.0.0.0/feed", wantErr: httpsafe.ErrIPBlocked},
		{tname: "IPv6 unspecified", rawURL: "http://[::]/feed", wantErr: httpsafe.ErrIPBlocked},

		// blocked: multicast
		{tname: "IPv4 multicast", rawURL: "http://224.0.0.1/feed", wantErr: httpsafe.ErrIPBlocked},
		{tname: "IPv6 multicast", rawURL: "http://[ff02::1]/feed", wantErr: httpsafe.ErrIPBlocked},

		// allowed: public addresses
		{tname: "IPv4 public", rawURL: "http://8.8.8.8/feed"},
		{tname: "IPv6 public", rawURL: "http://[2001:4860:4860::8888]/feed"},
		{tname: "IPv4 just outside shared address space (RFC6598)", rawURL: "http://100.128.0.1/feed"},
	}

	for _, tc := range cases {
		t.Run(tc.tname, func(t *testing.T) {
			err := httpsafe.ValidateURL(t.Context(), tc.rawURL)

			if tc.wantErr != nil {
				if !errors.Is(err, tc.wantErr) {
					t.Errorf("want error %q, got %q", tc.wantErr, err)
				}
				return
			}

			if err != nil {
				t.Errorf("want no error, got %q", err)
			}
		})
	}
}

func TestCheckDestination_InvalidURL(t *testing.T) {
	err := httpsafe.ValidateURL(t.Context(), "http://[::not-an-address/feed")
	if err == nil {
		t.Error("want a parse error, got nil")
	}
}

func TestNewSafeTransport_BlocksDial(t *testing.T) {
	transport, err := httpsafe.NewSafeTransport()
	if err != nil {
		t.Fatalf("want no error, got %q", err)
	}

	if transport.DialContext == nil {
		t.Fatal("want a non-nil DialContext")
	}

	conn, err := transport.DialContext(t.Context(), "tcp", "127.0.0.1:80")
	if conn != nil {
		conn.Close()
		t.Error("want no connection to be established")
	}

	if !errors.Is(err, httpsafe.ErrIPBlocked) {
		t.Errorf("want error %q, got %q", httpsafe.ErrIPBlocked, err)
	}
}

// TestNewSafeTransport_DisablesProxy guards against a Transport.DialContext
// override silently applying the address-blocklist to a configured forward
// proxy's address instead of the request's actual destination -- Go's
// Transport dials the proxy's address, not the origin's, when a proxy is
// set, so this transport must not honor one.
func TestNewSafeTransport_DisablesProxy(t *testing.T) {
	transport, err := httpsafe.NewSafeTransport()
	if err != nil {
		t.Fatalf("want no error, got %q", err)
	}

	if transport.Proxy != nil {
		t.Error("want Proxy to be disabled (nil)")
	}
}
