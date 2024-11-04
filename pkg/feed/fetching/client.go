// Copyright (c) VirtualTam
// SPDX-License-Identifier: MIT

package fetching

import (
	"context"
	"net/http"
	"time"

	"github.com/mmcdole/gofeed"
)

// A Client performs outgoing HTTP requests to get remote feed data.
//
// It leverages previously saved HTTP headers to perform HTTP conditional requests
// and benefit from remote server caching.
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
// - Use the value of the ETag header to set the If-None-Match header
// - Use the value of the Last-Modified header to set the If-Modified-Since header
func (c *Client) Fetch(feedURL string, eTag string, lastModified time.Time) (FeedStatus, error) {
	ctx := context.Background()

	req, err := http.NewRequestWithContext(ctx, "GET", feedURL, nil)
	if err != nil {
		return FeedStatus{}, err
	}

	req.Header.Set("User-Agent", c.userAgent)

	if eTag != "" {
		req.Header.Set(HeaderIfNoneMatch, eTag)
	}

	if !lastModified.IsZero() {
		req.Header.Set(HeaderIfModifiedSince, formatLastModified(lastModified))
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return FeedStatus{}, err
	}

	defer func() {
		ce := resp.Body.Close()
		if ce != nil {
			err = ce
		}
	}()

	respETag := resp.Header.Get(HeaderEntityTag)
	respLastModified := parseLastModified(resp.Header.Get(HeaderLastModified))

	feedStatus := FeedStatus{
		StatusCode:   resp.StatusCode,
		ETag:         respETag,
		LastModified: respLastModified,
	}

	if resp.StatusCode == http.StatusNotModified {
		return feedStatus, nil
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return FeedStatus{}, gofeed.HTTPError{
			StatusCode: resp.StatusCode,
			Status:     resp.Status,
		}
	}

	parsedFeed, err := c.feedParser.Parse(resp.Body)
	if err != nil {
		return FeedStatus{}, err
	}

	feedStatus.Feed = parsedFeed

	return feedStatus, nil
}
