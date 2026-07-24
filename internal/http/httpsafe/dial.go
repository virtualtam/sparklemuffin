// Copyright VirtualTam 2022, 2026
// SPDX-License-Identifier: MIT

package httpsafe

import (
	"context"
	"net"
	"time"
)

// dialContext resolves the address being dialed, rejects it if any
// resolved IP belongs to a blocked class of network (see isBlockedIP), and
// dials the resolved IPs directly -- rather than letting net.Dialer resolve
// the original hostname a second time -- so that the address actually
// connected to is guaranteed to be one that was checked. This closes the
// DNS-rebinding gap that a "resolve, check, then dial the hostname again"
// approach would leave open.
//
// It tries each allowed IP in turn until one succeeds, so a host with
// several resolved addresses (dual-stack, round-robin DNS) isn't limited to
// its first, in case that particular address is unreachable.
func dialContext(ctx context.Context, network, addr string) (net.Conn, error) {
	const (
		timeout   = 30 * time.Second
		keepAlive = 30 * time.Second
	)

	dialer := &net.Dialer{
		Timeout:   timeout,
		KeepAlive: keepAlive,
	}

	host, port, err := net.SplitHostPort(addr)
	if err != nil {
		return nil, err
	}

	ips, err := lookupIPs(ctx, host)
	if err != nil {
		return nil, err
	}

	var dialErr error
	for _, ip := range ips {
		conn, err := dialer.DialContext(ctx, network, net.JoinHostPort(ip.String(), port))
		if err == nil {
			return conn, nil
		}
		dialErr = err
	}

	return nil, dialErr
}
