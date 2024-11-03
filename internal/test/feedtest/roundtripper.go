// Copyright (c) VirtualTam
// SPDX-License-Identifier: MIT

package feedtest

import (
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/gorilla/feeds"
)

type RoundTripper struct {
	content string
	eTag    string

	Requests []*http.Request
}

func NewRoundTripper(t *testing.T, feed feeds.Feed) *RoundTripper {
	t.Helper()

	feedStr, err := feed.ToAtom()
	if err != nil {
		t.Fatalf("failed to encode feed to Atom: %q", err)
	}

	return &RoundTripper{
		content: feedStr,
		eTag:    HashETag(feedStr),
	}
}

func (rt *RoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	rt.Requests = append(rt.Requests, req)

	ifNoneMatch := req.Header.Get(HeaderIfNoneMatch)
	if ifNoneMatch == rt.eTag {
		resp := &http.Response{
			StatusCode: http.StatusNotModified,
			Header: http.Header{
				"Etag": {rt.eTag},
			},
		}

		return resp, nil
	}

	resp := &http.Response{
		StatusCode: http.StatusOK,
		Header: http.Header{
			"Content-Disposition": {"attachment; filename=test.atom"},
			"Content-Type":        {"application/atom+xml"},
			"Etag":                {rt.eTag},
		},
		Body: io.NopCloser(strings.NewReader(rt.content)),
	}

	return resp, nil
}
