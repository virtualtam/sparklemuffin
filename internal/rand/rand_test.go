// Copyright (c) VirtualTam
// SPDX-License-Identifier: MIT

package rand

import (
	"encoding/base64"
	"errors"
	"testing"
)

func TestRandomBytes(t *testing.T) {
	cases := []struct {
		tname   string
		length  int
		wantErr error
	}{
		// Nominal cases
		{
			tname:  "zero-length",
			length: 0,
		},
		{
			tname:  "positive length",
			length: 10,
		},

		// Error cases
		{
			tname:   "negative length",
			length:  -1,
			wantErr: ErrRandNegativeLength,
		},
	}

	for _, tc := range cases {
		randomBytes, err := RandomBytes(tc.length)

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

		if len(randomBytes) != tc.length {
			t.Errorf("expected result of length %d, got %d", tc.length, len(randomBytes))
		}
	}
}

func TestRandomBase64URLString(t *testing.T) {
	cases := []struct {
		tname   string
		length  int
		wantErr error
	}{
		// Nominal cases
		{
			tname:  "zero-length",
			length: 0,
		},
		{
			tname:  "positive length",
			length: 10,
		},

		// Error cases
		{
			tname:   "negative length",
			length:  -1,
			wantErr: ErrRandNegativeLength,
		},
	}

	for _, tc := range cases {
		randomString, err := RandomBase64URLString(tc.length)

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

		if len(randomString)%4 != 0 {
			t.Error("expected length to be a multiple of 4")
		}

		decodedBytes, err := base64.URLEncoding.DecodeString(randomString)
		if err != nil {
			t.Error("failed to decode Base64URL string")
		}

		if len(decodedBytes) != tc.length {
			t.Errorf("expected length of decoded []byte %d, got %d", tc.length, len(decodedBytes))
		}
	}
}
