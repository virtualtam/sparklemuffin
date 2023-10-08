package static

import "embed"

//go:embed *.css *.png */*.ttf *.webmanifest
var FS embed.FS
