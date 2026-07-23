// Copyright VirtualTam 2022, 2026
// SPDX-License-Identifier: MIT

package httpsafe_test

import (
	"testing"

	"github.com/virtualtam/sparklemuffin/internal/http/httpsafe"
)

func TestProxyConfigured(t *testing.T) {
	cases := []struct {
		tname      string
		httpProxy  string
		httpsProxy string
		want       bool
	}{
		{tname: "no proxy configured"},
		{tname: "HTTP_PROXY set", httpProxy: "http://proxy.example.com:3128", want: true},
		{tname: "HTTPS_PROXY set", httpsProxy: "http://proxy.example.com:3128", want: true},
		{tname: "both set", httpProxy: "http://proxy.example.com:3128", httpsProxy: "http://proxy.example.com:3128", want: true},
	}

	for _, tc := range cases {
		t.Run(tc.tname, func(t *testing.T) {
			t.Setenv("HTTP_PROXY", tc.httpProxy)
			t.Setenv("HTTPS_PROXY", tc.httpsProxy)
			t.Setenv("http_proxy", "")
			t.Setenv("https_proxy", "")
			t.Setenv("NO_PROXY", "")
			t.Setenv("no_proxy", "")

			got := httpsafe.ProxyConfigured()

			if got != tc.want {
				t.Errorf("want %v, got %v", tc.want, got)
			}
		})
	}
}
