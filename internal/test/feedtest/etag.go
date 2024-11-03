// Copyright (c) VirtualTam
// SPDX-License-Identifier: MIT

package feedtest

import (
	"crypto/sha256"
	"fmt"
)

func HashETag(content string) string {
	h := sha256.New()
	h.Write([]byte(content))
	sum := h.Sum(nil)

	return fmt.Sprintf("W/\"%x\"", sum)
}
