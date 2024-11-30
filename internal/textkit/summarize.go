// Copyright (c) VirtualTam
// SPDX-License-Identifier: MIT

package textkit

import (
	"regexp"
	"strings"
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
		paragraphLength := len(paragraphs[0])
		if paragraphLength < truncateAfter {
			return text[:paragraphLength-1] + "…"
		}

		return text[:truncateAfter-1] + "…"
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
