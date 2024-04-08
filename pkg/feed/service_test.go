// Copyright (c) VirtualTam
// SPDX-License-Identifier: MIT

package feed

import (
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/gorilla/feeds"
)

type testRoundTripper struct{}

func (testRoundTripper) RoundTrip(*http.Request) (*http.Response, error) {
	feed := &feeds.Feed{
		Title: "Local Test",
	}

	feedStr, err := feed.ToAtom()
	if err != nil {
		panic(err)
	}

	resp := &http.Response{
		StatusCode: 200,
		Header: map[string][]string{
			"Content-Disposition": {"attachment; filename=test.atom"},
			"Content-Type":        {"application/atom+xml"},
		},
		Body: io.NopCloser(strings.NewReader(feedStr)),
	}

	return resp, nil
}

func TestServiceGetOrCreateFeed(t *testing.T) {
	testClient := &http.Client{
		Transport: testRoundTripper{},
	}

	cases := []struct {
		tname           string
		feedURL         string
		repositoryFeeds []Feed
		want            Feed
		wantErr         error
	}{
		// nominal cases
		{
			tname:   "new feed (resolve metadata)",
			feedURL: "http://test.local",
			want: Feed{
				URL:   "http://test.local",
				Title: "Local Test",
				Slug:  "local-test",
			},
		},
		{
			tname:   "existing feed (leave unchanged, do not resolve metadata)",
			feedURL: "http://test.local",
			repositoryFeeds: []Feed{
				{
					URL:   "http://test.local",
					Title: "Existing Test",
					Slug:  "existing-test",
				},
			},
			want: Feed{
				URL:   "http://test.local",
				Title: "Existing Test",
				Slug:  "existing-test",
			},
		},

		// error cases
		{
			tname:   "empty URL",
			wantErr: ErrFeedURLInvalid,
		},
		{
			tname:   "empty URL (whitespace)",
			feedURL: "     ",
			wantErr: ErrFeedURLInvalid,
		},
		{
			tname:   "invalid URL (no host)",
			feedURL: "http://",
			wantErr: ErrFeedURLNoHost,
		},
		{
			tname:   "invalid URL (no scheme)",
			feedURL: "domain.tld",
			wantErr: ErrFeedURLNoScheme,
		},
		{
			tname:   "invalid URL (unsupported scheme)",
			feedURL: "ftp://domain.tld",
			wantErr: ErrFeedURLUnsupportedScheme,
		},
	}

	for _, tc := range cases {
		t.Run(tc.tname, func(t *testing.T) {
			r := &fakeRepository{
				Feeds: tc.repositoryFeeds,
			}
			s := NewService(r, testClient)

			got, err := s.getOrCreateFeed(tc.feedURL)

			if tc.wantErr != nil {
				if errors.Is(err, tc.wantErr) {
					return
				}
				if err == nil {
					t.Fatalf("want error %q, got nil", tc.wantErr)
				}
				t.Fatalf("want error %q, got %q", tc.wantErr, err)
			}

			if err != nil {
				t.Fatalf("want no error, got %q", err)
			}

			assertFeedEquals(t, got, tc.want)
		})
	}
}
