package www

import (
	"html/template"
	"strings"

	"github.com/alecthomas/chroma/formatters/html"
	"github.com/yuin/goldmark"
	highlighting "github.com/yuin/goldmark-highlighting"
	"github.com/yuin/goldmark/extension"
)

// markdownToHTML initializes a Markdown renderer and returns a function
// suitable for usage with html/template.
func markdownToHTML() func(str string) template.HTML {
	markdown := goldmark.New(
		goldmark.WithExtensions(
			// https://github.com/yuin/goldmark#linkify-extension
			extension.NewLinkify(
				extension.WithLinkifyAllowedProtocols(
					[][]byte{
						[]byte("http:"),
						[]byte("https:"),
					},
				),
			),

			// https://github.com/yuin/goldmark-highlighting
			// https://github.com/alecthomas/chroma
			// https://github.com/alecthomas/chroma/tree/master/styles
			highlighting.NewHighlighting(
				highlighting.WithStyle("nord"),
				highlighting.WithFormatOptions(
					html.WithLineNumbers(true),
				),
			),
		),
	)

	return func(str string) template.HTML {
		var buf strings.Builder
		if err := markdown.Convert([]byte(str), &buf); err != nil {
			return template.HTML("")
		}

		return template.HTML(buf.String())
	}
}
