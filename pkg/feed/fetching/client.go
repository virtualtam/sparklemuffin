// Copyright (c) VirtualTam
// SPDX-License-Identifier: MIT

package fetching

import (
	"context"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
	"unicode"

	"github.com/cespare/xxhash/v2"
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

// Fetch performs an HTTP request to get feed data and parses it.
//
// Adapted from gofeed.Parser.ParseURL with the following modifications:
// - User-Agent header;
// - Use the value of the ETag header to set the If-None-Match header;
// - Use the value of the Last-Modified header to set the If-Modified-Since header.
func (c *Client) Fetch(ctx context.Context, feedURL string, eTag string, lastModified time.Time) (FeedStatus, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, feedURL, nil)
	if err != nil {
		return FeedStatus{}, fmt.Errorf("feed: failed to create request: %w", err)
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
		return FeedStatus{}, fmt.Errorf("feed: failed to perform request: %w", err)
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

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return FeedStatus{}, fmt.Errorf("feed: failed to read response body: %w", err)
	}

	parsedFeed, err := c.parse(body)
	if err != nil {
		return FeedStatus{}, err
	}

	feedStatus.Feed = parsedFeed
	feedStatus.Hash = xxhash.Sum64(body)

	return feedStatus, nil
}

// parse wraps gofeed.Parser to handle minor parsing errors for feeds containing improperly formatted data
// or invalid Unicode characters.
func (c *Client) parse(body []byte) (*gofeed.Feed, error) {
	feedStr := string(body)
	parsedFeed, err := c.feedParser.ParseString(feedStr)

	var xmlSyntaxError *xml.SyntaxError
	if errors.As(err, &xmlSyntaxError) {
		// Atom and RSS feed parsing errors.

		if strings.HasPrefix(xmlSyntaxError.Msg, "illegal character code") {
			// The feed contains non-printable Unicode characters.
			//
			// Filter the input to remove non-printable Unicode characters, then attempt to parse feed data.
			//
			// See:
			// - https://pkg.go.dev/unicode#IsPrint
			// - https://pkg.go.dev/unicode#IsGraphic
			// - https://pkg.go.dev/unicode#RangeTable
			filtered := strings.Map(func(r rune) rune {
				if unicode.IsPrint(r) {
					return r
				}
				return -1
			}, feedStr)

			parsedFeed, err = c.feedParser.ParseString(filtered)
		}
	}

	if err != nil {
		return &gofeed.Feed{}, err
	}

	return parsedFeed, nil
}
