// Copyright (c) VirtualTam
// SPDX-License-Identifier: MIT

package textkit_test

import (
	"testing"

	"github.com/virtualtam/sparklemuffin/internal/textkit"
)

func TestSummarize(t *testing.T) {
	const (
		summaryKeepIfUnder   = 20
		summaryTruncateAfter = 50
	)

	cases := []struct {
		tname string
		text  string
		want  string
	}{
		{
			tname: "empty string",
		},
		{
			tname: "one short paragraph",
			text:  "Hello",
			want:  "Hello",
		},
		{
			tname: "one medium paragraph",
			text:  "Hello world! This is a medium string.",
			want:  "Hello world! This is a medium string.",
		},
		{
			tname: "one long paragraph, truncated on a whitespace",
			text:  "Hello world! This is a long string that should be summarized.",
			want:  "Hello world! This is a long string that should be…",
		},
		{
			tname: "one long paragraph, truncated on a word",
			text:  "Hello world! This is a long string that should rather be summarized.",
			want:  "Hello world! This is a long string that should rat…",
		},
		{
			tname: "one medium paragraph with Unicode characters",
			text:  "Hello! This is a medium Unicode string: 🧡💛💚💙💜",
			want:  "Hello! This is a medium Unicode string: 🧡💛💚💙💜",
		},
		{
			tname: "one long paragraph with Unicode characters",
			text:  "Hello, world! This is a longer Unicode string: 🧡💛💚💙💜",
			want:  "Hello, world! This is a longer Unicode string: 🧡💛💚…",
		},
		{
			tname: "two paragraphs",
			text:  "First paragraph.\n\nThis is a second paragraph.",
			want:  "First paragraph.\n\nThis is a second paragraph.",
		},
		{
			tname: "three paragraphs",
			text:  "First paragraph.\n\nThis is a second paragraph.\n\nAnd a third paragraph.",
			want:  "First paragraph.\n\nThis is a second paragraph.",
		},
	}

	for _, tc := range cases {
		got := textkit.Summarize(tc.text, summaryKeepIfUnder, summaryTruncateAfter)
		if got != tc.want {
			t.Errorf("want %q, got %q", tc.want, got)
		}
	}
}
