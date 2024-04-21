// Copyright (c) VirtualTam
// SPDX-License-Identifier: MIT

package paginate

import (
	"net/url"
	"testing"
)

func TestGetPageNumber(t *testing.T) {
	t.Run("empty (param not set)", func(t *testing.T) {
		got, _, err := GetPageNumber(url.Values{})
		if err != nil {
			t.Fatalf("expected no error, got %q", err)
		}

		if got != 1 {
			t.Errorf("want page 1, got %d", got)
		}
	})

	t.Run("positive integer", func(t *testing.T) {
		queryParams := url.Values{
			"page": []string{"12"},
		}

		got, _, err := GetPageNumber(queryParams)
		if err != nil {
			t.Fatalf("expected no error, got %q", err)
		}

		if got != 12 {
			t.Errorf("want page 12, got %d", got)
		}
	})

	errorCases := []struct {
		tname string
		input string
	}{
		{
			tname: "negative integer",
			input: "-264",
		},
		{
			tname: "random chars",
			input: "IEJUd7RAOW",
		},
	}

	for _, tc := range errorCases {
		t.Run(tc.tname, func(t *testing.T) {
			queryParams := url.Values{
				"page": []string{tc.input},
			}

			_, _, err := GetPageNumber(queryParams)
			if err == nil {
				t.Fatal("expected an error, got none")
			}
		})
	}
}
