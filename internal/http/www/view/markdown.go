// Copyright (c) VirtualTam
// SPDX-License-Identifier: MIT

package view

import (
	"fmt"
	"html/template"
	"strings"

	"github.com/alecthomas/chroma/v2/formatters/html"
	"github.com/rs/zerolog/log"
	"github.com/yuin/goldmark"
	highlighting "github.com/yuin/goldmark-highlighting/v2"
	"github.com/yuin/goldmark/extension"
)

var (
	// Initializes a Markdown renderer.
	markdown = goldmark.New(
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
				highlighting.WithStyle("catppuccin-latte"),
				highlighting.WithFormatOptions(
					html.WithLineNumbers(true),
				),
			),
		),
	)
)

// MarkdownToHTML renders a Markdown string as HTML.
func MarkdownToHTML(str string) (string, error) {
	var buf strings.Builder
	if err := markdown.Convert([]byte(str), &buf); err != nil {
		return "", fmt.Errorf("failed to render Markdown as HTML: %w", err)
	}

	return buf.String(), nil
}

// MarkdownToHTMLFunc returns a function suitable for usage with html/template.
func MarkdownToHTMLFunc() func(str string) template.HTML {
	return func(str string) template.HTML {
		var buf strings.Builder
		if err := markdown.Convert([]byte(str), &buf); err != nil {
			log.Error().Err(err).Msg("failed to render Markdown as HTML")
			return template.HTML("")
		}

		return template.HTML(buf.String())
	}
}
