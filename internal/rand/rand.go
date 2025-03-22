// Copyright (c) VirtualTam
// SPDX-License-Identifier: MIT

package rand

import (
	"crypto/rand"
	"encoding/base64"
	"errors"
)

var (
	ErrRandUnexpectedLength = errors.New("rand: unexpected length")
	ErrRandNegativeLength   = errors.New("rand: negative length")
)

// RandomBytes generates a cryptographically secure random byte array.
func RandomBytes(length int) ([]byte, error) {
	if length < 0 {
		return []byte{}, ErrRandNegativeLength
	}

	b := make([]byte, length)

	n, err := rand.Read(b)
	if err != nil {
		return []byte{}, err
	}
	if n != length {
		return []byte{}, ErrRandUnexpectedLength
	}

	return b, nil
}

// RandomBase64URLString generates a cryptographically secure, URL-safe,
// Base64-encoded string.
func RandomBase64URLString(length int) (string, error) {
	b, err := RandomBytes(length)
	if err != nil {
		return "", err
	}

	s := base64.URLEncoding.EncodeToString(b)

	return s, nil
}
