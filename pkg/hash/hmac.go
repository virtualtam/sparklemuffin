// Copyright (c) VirtualTam
// SPDX-License-Identifier: MIT

// Package hash provides hashing helpers to generate application tokens that can
// be used in a web application.
package hash

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"hash"
	"sync"
)

// HMAC provides helpers to generate strong application tokens.
type HMAC struct {
	hmac hash.Hash
	l    sync.Locker
}

// Hash returns the Base64 URL-encoded HMAC hash of the provided input.
func (h *HMAC) Hash(input string) (string, error) {
	h.l.Lock()
	defer h.l.Unlock()

	h.hmac.Reset()

	_, err := h.hmac.Write([]byte(input))
	if err != nil {
		return "", err
	}

	b := h.hmac.Sum(nil)
	return base64.URLEncoding.EncodeToString(b), nil
}

// NewHMAC returns an HMAC initialized from a secret key.
func NewHMAC(key string) *HMAC {
	h := hmac.New(sha256.New, []byte(key))
	return &HMAC{
		hmac: h,
		l:    &sync.Mutex{},
	}
}
