// Copyright VirtualTam 2022, 2026
// SPDX-License-Identifier: MIT

package static

import "embed"

//go:embed *.css *.js *.png *.ttf *.woff2 *.webmanifest
var FS embed.FS
