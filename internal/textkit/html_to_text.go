// Copyright (c) VirtualTam
// SPDX-License-Identifier: MIT

package textkit

import (
	"regexp"
	"strings"

	"github.com/k3a/html2text"
)

var (
	removeHTMLAnchorRegexp = regexp.MustCompile(` <[^>]+>`)
)

func NormalizeHTMLToText(htmlDocument string) string {
	text := html2text.HTML2TextWithOptions(
		htmlDocument,
		html2text.WithUnixLineBreaks(),
		html2text.WithListSupport(),
		html2text.WithLinksInnerText(),
	)

	text = removeHTMLAnchorRegexp.ReplaceAllString(text, "")
	return strings.TrimSpace(text)
}
