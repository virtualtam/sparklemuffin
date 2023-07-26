package static

import "embed"

//go:embed *.css *.png *.webmanifest
var FS embed.FS
