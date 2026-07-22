// Copyright VirtualTam 2022, 2026
// SPDX-License-Identifier: MIT

package htmx

import "errors"

var ErrMissingRequestHeader = errors.New("htmx: missing HX-Request header")
