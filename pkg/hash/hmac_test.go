package hash

import (
	"crypto/sha256"
	"encoding/base64"
	"testing"
)

func TestHMAC(t *testing.T) {
	secretKey := "SecretChiefs3"
	hmac := NewHMAC(secretKey)

	hashed, err := hmac.Hash("Masada Book III")

	if err != nil {
		t.Errorf("got an error but didn't expect one: %q", err)
	}

	data, err := base64.URLEncoding.DecodeString(hashed)

	if err != nil {
		t.Errorf("failed to decode a base64-url-encoded hash: %q", err)
	}

	if len(data) != sha256.Size {
		t.Errorf("expected a decoded hash of length %d, got hash of length %d", sha256.Size, len(data))
	}
}
