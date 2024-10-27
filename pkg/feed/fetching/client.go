// Copyright (c) VirtualTam
// SPDX-License-Identifier: MIT

package fetching

import (
	"context"
	"net/http"

	"github.com/mmcdole/gofeed"
)

// A Client performs outgoing HTTP requests to get remote feed data.
type Client struct {
	httpClient *http.Client
	userAgent  string

	feedParser *gofeed.Parser
}

// NewClient initializes and returns a Client.
func NewClient(httpClient *http.Client, userAgent string) *Client {
	return &Client{
		httpClient: httpClient,
		userAgent:  userAgent,
		feedParser: gofeed.NewParser(),
	}
}

// Fetch performs a HTTP request to get feed data and parses it.
//
// Adapted from gofeed.Parser.ParseURL with the following modifications:
// - User-Agent header
// - TODO: ETag header
// - TODO: If-Modified-Since header
func (c *Client) Fetch(feedURL string) (*gofeed.Feed, error) {
	ctx := context.Background()

	req, err := http.NewRequestWithContext(ctx, "GET", feedURL, nil)
	if err != nil {
		return &gofeed.Feed{}, err
	}

	req.Header.Set("User-Agent", c.userAgent)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return &gofeed.Feed{}, err
	}

	if resp != nil {
		defer func() {
			ce := resp.Body.Close()
			if ce != nil {
				err = ce
			}
		}()
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return &gofeed.Feed{}, gofeed.HTTPError{
			StatusCode: resp.StatusCode,
			Status:     resp.Status,
		}
	}

	return c.feedParser.Parse(resp.Body)
}
