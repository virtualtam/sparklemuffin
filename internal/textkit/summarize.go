// Copyright (c) VirtualTam
// SPDX-License-Identifier: MIT

package textkit

import (
	"regexp"
	"strings"

	"golang.org/x/exp/utf8string"
)

var (
	paragraphSplitRegexp = regexp.MustCompile(`\n\s*\n`)
)

// Summarize returns a summary of the given text.
//
// If the text is shorter than keepIfUnder characters, it is returned as is.
// The summary is then built by keeping the first paragraphs, up to truncateAfter characters.
func Summarize(text string, keepIfUnder int, truncateAfter int) string {
	if text == "" {
		return ""
	}

	if len(text) <= keepIfUnder {
		return text
	}

	paragraphs := paragraphSplitRegexp.Split(text, -1)
	if len(paragraphs) == 1 {
		utf8paragraph := utf8string.NewString(paragraphs[0])

		if utf8paragraph.RuneCount() <= truncateAfter {
			return paragraphs[0]
		}

		return strings.TrimSpace(utf8paragraph.Slice(0, truncateAfter)) + "…"
	}

	return summarizeParagraphs(paragraphs, truncateAfter)
}

// summarizeParagraphs builds a summary from multiple paragraphs, keeping it under truncateAfter characters.
func summarizeParagraphs(paragraphs []string, truncateAfter int) string {
	var summary strings.Builder

	for _, paragraph := range paragraphs {
		paragraph = strings.TrimSpace(paragraph)
		if paragraph == "" {
			continue
		}

		if summary.Len()+len(paragraph) > truncateAfter {
			break
		}

		if summary.Len() > 0 {
			summary.WriteString("\n\n")
		}

		summary.WriteString(paragraph)
	}

	return summary.String()
}
