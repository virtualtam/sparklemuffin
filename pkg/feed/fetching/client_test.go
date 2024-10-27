package fetching_test

import (
	"io"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/feeds"
	"github.com/virtualtam/sparklemuffin/pkg/feed/fetching"
)

func TestClientFetch(t *testing.T) {
	userAgent := "sparklemuffin/test"

	t.Run("User-Agent", func(t *testing.T) {
		transport := &feedRoundTripper{}
		httpClient := &http.Client{
			Transport: transport,
		}

		client := fetching.NewClient(httpClient, userAgent)

		_, err := client.Fetch("")
		if err != nil {
			t.Fatalf("want no error, got %q", err)
		}

		if len(transport.requests) != 1 {
			t.Fatalf("want 1 request, got %d", len(transport.requests))
		}

		gotUserAgent := transport.requests[0].Header.Get("User-Agent")

		if gotUserAgent != userAgent {
			t.Errorf("want User-Agent %q, got %q", userAgent, gotUserAgent)
		}
	})
}

type feedRoundTripper struct {
	requests []*http.Request
}

func (rt *feedRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	rt.requests = append(rt.requests, req)

	now := time.Now().UTC()
	yesterday := now.Add(-24 * time.Hour)

	feed := &feeds.Feed{
		Title:   "Local Test",
		Updated: now,
		Items: []*feeds.Item{
			{
				Id:    "http://test.local/first-post",
				Title: "First post!",
				Link: &feeds.Link{
					Href: "http://test.local/first-post",
				},
				Created: now,
				Updated: now,
			},
			{
				Id:    "http://test.local/hello-world",
				Title: "Hello World",
				Link: &feeds.Link{
					Href: "http://test.local/hello-world",
				},
				Created: yesterday,
				Updated: yesterday,
			},
		},
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
