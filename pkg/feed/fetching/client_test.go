// Copyright (c) VirtualTam
// SPDX-License-Identifier: MIT

package fetching_test

import (
	"net/http"
	"strconv"
	"strings"
	"testing"
	"text/template"
	"time"

	"github.com/cespare/xxhash/v2"

	"github.com/virtualtam/sparklemuffin/internal/test/feedtest"
	"github.com/virtualtam/sparklemuffin/pkg/feed/fetching"
)

const (
	userAgent = "sparklemuffin/test"
)

func TestClientFetch(t *testing.T) {

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
			transport := feedtest.NewRoundTripperFromFeed(t, feed)
			httpClient := &http.Client{
				Transport: transport,
			}

			client := fetching.NewClient(httpClient, userAgent)
			feedStatus, err := client.Fetch(t.Context(), "", tc.eTag, tc.lastModified)
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

func TestClientFetch_NonPrintableUnicode(t *testing.T) {
	now := time.Now()
	lastWeek := now.Add(-7 * 24 * time.Hour)

	contentTmpl := `<?xml version="1.0" encoding="utf-8"?>
<feed xmlns="http://www.w3.org/2005/Atom">
  <title>Unicode Test Feed</title>
  <entry>
    <title>Example Entry</title>
    <id>urn:uuid:abcdefab-cdef-abcd-efab-cdefabcdefab</id>
    <updated>2025-01-01T00:00:00Z</updated>
    <summary>Hello{{.}}World</summary>
  </entry>
</feed>
`

	tmpl, err := template.New("test").Parse(contentTmpl)
	if err != nil {
		t.Fatalf("failed to parse template: %s", err)
	}

	// Non-printable, non-whitespace control codes.
	//
	// - https://en.wikipedia.org/wiki/C0_and_C1_control_codes
	// - https://en.wikipedia.org/wiki/List_of_Unicode_characters#Control_codes
	controlCodeRunes := []rune{
		0x00, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07,
		0x0e, 0x0f, 0x10, 0x11, 0x12, 0x13, 0x14, 0x15,
		0x16, 0x17, 0x18, 0x19, 0x1a, 0x1b, 0x1c, 0x1d,
		0x1e, 0x1f,
	}

	for _, controlCodeRune := range controlCodeRunes {
		t.Run(strconv.QuoteRune(controlCodeRune), func(t *testing.T) {
			sb := &strings.Builder{}
			if err := tmpl.Execute(sb, string(controlCodeRune)); err != nil {
				t.Fatalf("failed to execute template: %s", err)
			}

			transport := feedtest.NewRoundTripperFromString(t, sb.String(), lastWeek)
			httpClient := &http.Client{
				Transport: transport,
			}

			client := fetching.NewClient(httpClient, userAgent)

			parsedFeed, err := client.Fetch(t.Context(), "", "", lastWeek)
			if err != nil {
				t.Fatalf("want no error, got %q", err)
			}

			if len(parsedFeed.Feed.Items) != 1 {
				t.Fatalf("want 1 item, got %d", len(parsedFeed.Feed.Items))
			}

			if parsedFeed.Feed.Items[0].Description != "HelloWorld" {
				t.Errorf("want description %q, got %q", "HelloWorld", parsedFeed.Feed.Items[0].Description)
			}
		})
	}
}
