// Copyright VirtualTam 2022, 2026
// SPDX-License-Identifier: MIT

package httpsafe

import (
	"context"
	"fmt"
	"net/url"
)

// ValidateURL resolves rawURL's host and returns ErrIPBlocked if
// any of the addresses it resolves to belongs to a blocked class of network.
// It performs no HTTP request.
//
// This is a best-effort, point-in-time check meant to give fast feedback
// (e.g. when a user submits a URL); the actual, authoritative enforcement
// happens at dial time, see NewSafeTransport.
func ValidateURL(ctx context.Context, rawURL string) error {
	parsedURL, err := url.Parse(rawURL)
	if err != nil {
		return fmt.Errorf("httpsafe: failed to parse URL: %w", err)
	}

	_, err = lookupIPs(ctx, parsedURL.Hostname())
	return err
}
