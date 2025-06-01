// Copyright (c) VirtualTam
// SPDX-License-Identifier: MIT

package static

import "embed"

//go:embed *.css *.js *.png *.ttf *.woff2 *.webmanifest
var FS embed.FS
