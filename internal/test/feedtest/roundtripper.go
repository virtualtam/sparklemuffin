// Copyright (c) VirtualTam
// SPDX-License-Identifier: MIT

package feedtest

import (
	"io"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/feeds"
)

var (
	_ http.RoundTripper = &RoundTripper{}
)

// RoundTripper records incoming HTTP requests and responds with Atom feed data.
//
// It supports HTTP conditional requests using the following HTTP headers:
// - ETag (response) / If-None-Match (request);
// - Last-Modified (response) / If-Modified-Since (request).
type RoundTripper struct {
	content         string
	eTag            string
	lastModified    time.Time
	lastModifiedStr string

	Requests []*http.Request
}

// NewRoundTripper initializes and returns a RoundTripper.
func NewRoundTripper(t *testing.T, feed feeds.Feed) *RoundTripper {
	t.Helper()

	feedStr, err := feed.ToAtom()
	if err != nil {
		t.Fatalf("failed to encode feed to Atom: %q", err)
	}

	return &RoundTripper{
		content:         feedStr,
		eTag:            HashETag(feedStr),
		lastModified:    feed.Updated,
		lastModifiedStr: formatLastModified(feed.Updated),
	}
}

func (rt *RoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	rt.Requests = append(rt.Requests, req)

	resp := &http.Response{
		Header: http.Header{},
	}
	resp.Header.Set(headerEntityTag, rt.eTag)
	resp.Header.Set(headerLastModified, rt.lastModifiedStr)

	ifNoneMatch := req.Header.Get(headerIfNoneMatch)
	ifModifiedSince := parseLastModified(req.Header.Get(headerIfModifiedSince))

	if ifNoneMatch == rt.eTag && ifModifiedSince.Equal(rt.lastModified) {
		// Simulate the behaviour of a remote server responding to a conditional request.
		//
		// Note that actual servers may behave differently!
		//
		// See:
		// - https://stackoverflow.com/questions/824152/what-takes-precedence-the-etag-or-last-modified-http-header
		resp.StatusCode = http.StatusNotModified
		return resp, nil
	}

	resp.StatusCode = http.StatusOK
	resp.Header.Set("Content-Disposition", "attachment; filename=test.atom")
	resp.Header.Set("Content-Type", "application/atom+xml")
	resp.Body = io.NopCloser(strings.NewReader(rt.content))

	return resp, nil
}
