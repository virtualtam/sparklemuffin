// Copyright (c) VirtualTam
// SPDX-License-Identifier: MIT

package fetching_test

import (
	"net/http"
	"testing"
	"time"

	"github.com/cespare/xxhash/v2"
	"github.com/virtualtam/sparklemuffin/internal/test/feedtest"
	"github.com/virtualtam/sparklemuffin/pkg/feed/fetching"
)

func TestClientFetch(t *testing.T) {
	userAgent := "sparklemuffin/test"

	now, err := time.Parse(time.DateTime, "2024-10-30 20:54:16")
	if err != nil {
		t.Fatalf("failed to parse date: %q", err)
	}
	lastWeek := now.Add(-7 * 24 * time.Hour)
	nextWeek := now.Add(7 * 24 * time.Hour)

	feed := feedtest.GenerateDummyFeed(t, now)
	feedStr, err := feed.ToAtom()
	if err != nil {
		t.Fatalf("failed to encode feed to Atom: %q", err)
	}
	feedETag := feedtest.HashETag(feedStr)
	feedLastModified := feed.Updated
	feedHash := xxhash.Sum64String(feedStr)

	cases := []struct {
		tname          string
		eTag           string
		lastModified   time.Time
		wantStatusCode int
	}{
		{
			tname:          "first request",
			wantStatusCode: http.StatusOK,
		},
		{
			tname:          "If-None-Match = ETag, If-Modified-Since = Last-Modified",
			eTag:           feedETag,
			lastModified:   feedLastModified,
			wantStatusCode: http.StatusNotModified,
		},
		{
			tname:          "If-None-Match = ETag, If-Modified-Since < Last-Modified",
			eTag:           feedETag,
			lastModified:   lastWeek,
			wantStatusCode: http.StatusOK,
		},
		{
			tname:          "If-None-Match = ETag, If-Modified-Since > Last-Modified",
			eTag:           feedETag,
			lastModified:   nextWeek,
			wantStatusCode: http.StatusOK,
		},
		{
			tname:          "If-None-Match != ETag, If-Modified-Since = Last-Modified",
			eTag:           `W/"5d2e8871966e0dd7ff59684904b3d9fecf6ab62a09869e26163efe8b2e07539d"`,
			lastModified:   lastWeek,
			wantStatusCode: http.StatusOK,
		},
	}

	for _, tc := range cases {
		t.Run(tc.tname, func(t *testing.T) {
			transport := feedtest.NewRoundTripper(t, feed)
			httpClient := &http.Client{
				Transport: transport,
			}

			client := fetching.NewClient(httpClient, userAgent)
			feedStatus, err := client.Fetch("", tc.eTag, tc.lastModified)
			if err != nil {
				t.Fatalf("want no error, got %q", err)
			}

			if len(transport.Requests) != 1 {
				t.Fatalf("want 1 request, got %d", len(transport.Requests))
			}

			gotUserAgent := transport.Requests[0].Header.Get("User-Agent")

			if gotUserAgent != userAgent {
				t.Errorf("want User-Agent %q, got %q", userAgent, gotUserAgent)
			}

			gotIfNoneMatch := transport.Requests[0].Header.Get("If-None-Match")

			if gotIfNoneMatch != tc.eTag {
				t.Errorf("want If-None-Match %q, got %q", tc.eTag, gotIfNoneMatch)
			}

			if feedStatus.StatusCode != tc.wantStatusCode {
				t.Errorf("want StatusCode %d, got %d", tc.wantStatusCode, feedStatus.StatusCode)
			}

			if feedStatus.ETag != feedETag {
				t.Errorf("want ETag %q, got %q", feedETag, feedStatus.ETag)
			}

			if feedStatus.StatusCode == http.StatusOK {
				if feedStatus.Hash != feedHash {
					t.Errorf("want Hash %d, got %d", feedHash, feedStatus.Hash)
				}
			}
		})
	}
}
