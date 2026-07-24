// Copyright VirtualTam 2022, 2026
// SPDX-License-Identifier: MIT

package httpsafe

import (
	"context"
	"errors"
	"fmt"
	"net"
)

// ErrIPBlocked indicates that a URL resolves to an address that must
// not be reached from this server.
//
// This verification helps prevent Server-Side Request Forgery (SSRF) attacks.
var ErrIPBlocked = errors.New("httpsafe: destination address is blocked")

// lookupIPs looks up host for the IP network using the local resolver.
//
// It returns a slice of that host's IP addresses, or ErrIPBlocked
// if any of them belongs to a blocked address range.
func lookupIPs(ctx context.Context, host string) ([]net.IP, error) {
	ips, err := net.DefaultResolver.LookupIP(ctx, "ip", host)
	if err != nil {
		return nil, fmt.Errorf("httpsafe: failed to resolve host %q: %w", host, err)
	}

	for _, ip := range ips {
		if isBlockedIP(ip) {
			return nil, fmt.Errorf("%w: %s resolves to %s", ErrIPBlocked, host, ip)
		}
	}

	return ips, nil
}

// isBlockedIP returns whether the IP address belongs to a range of blocked addresses.
func isBlockedIP(ip net.IP) bool {
	return ip.IsLoopback() ||
		ip.IsLinkLocalUnicast() ||
		ip.IsPrivate() ||
		ip.IsUnspecified() ||
		ip.IsMulticast() ||
		cgnatIPNet.Contains(ip)
}

// cgnatIPNet is the RFC 6598 Carrier-Grade NAT / shared address space (100.64.0.0/10).
//
// See:
// - https://en.wikipedia.org/wiki/Carrier-grade_NAT
// - https://blog.cloudflare.com/detecting-cgn-to-reduce-collateral-damage/
// - https://www.rfc-editor.org/info/rfc6598/
var cgnatIPNet = mustParseCIDR("100.64.0.0/10")

func mustParseCIDR(s string) *net.IPNet {
	_, cidr, err := net.ParseCIDR(s)
	if err != nil {
		panic(err)
	}

	return cidr
}
