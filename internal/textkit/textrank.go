// Copyright (c) VirtualTam
// SPDX-License-Identifier: MIT

package textkit

import (
	"fmt"
	"sort"
	"strings"

	textrank "github.com/DavidBelicza/TextRank/v2"
	"github.com/DavidBelicza/TextRank/v2/convert"
	"github.com/DavidBelicza/TextRank/v2/parse"
	"github.com/DavidBelicza/TextRank/v2/rank"
	anyascii "github.com/anyascii/go"
)

var (
	textRankReplacer = strings.NewReplacer(
		"/", " ",
		"-", " ",
		":", " ",
		"-", " ",
	)
)

// A TextRanker normalizes and processes text to extract the top ranking phrases and keywords.
type TextRanker struct {
	language  convert.Language
	algorithm rank.Algorithm
	rule      parse.Rule
}

// NewTextRanker initializes and returns a new TextRanker.
func NewTextRanker() *TextRanker {
	return &TextRanker{
		language:  textrank.NewDefaultLanguage(),
		algorithm: textrank.NewDefaultAlgorithm(),
		rule:      textrank.NewDefaultRule(),
	}
}

// RankTopNPhrases extracts the top N ranking phrases using TextRank.
func (t *TextRanker) RankTopNPhrases(text string, topN int) []string {
	phrases := t.rankPhrases(text)

	if len(phrases) > topN {
		phrases = phrases[:topN]
	}

	return phrases
}

func (t *TextRanker) RankTopNWords(text string, topN int) []string {
	words := t.rankWords(text)

	if len(words) > topN {
		words = words[:topN]
	}

	return words
}

// normalizeText normalizes and filters text to simplify word and phrase extraction.
func normalizeText(text string) string {
	text = anyascii.Transliterate(text)
	return textRankReplacer.Replace(text)
}

func (t *TextRanker) rankPhrases(text string) []string {
	text = normalizeText(text)

	tr := textrank.NewTextRank()
	tr.Populate(text, t.language, t.rule)
	tr.Ranking(t.algorithm)

	// Extract phrases
	rankedPhrases := textrank.FindPhrases(tr)

	// Stable sorting for phrases
	// This is required for consistent output and test reproducibility
	sort.Slice(rankedPhrases, func(i, j int) bool {
		if rankedPhrases[i].Weight > rankedPhrases[j].Weight {
			return true
		} else if rankedPhrases[i].Weight < rankedPhrases[j].Weight {
			return false
		}

		if rankedPhrases[i].Qty > rankedPhrases[j].Qty {
			return true
		} else if rankedPhrases[i].Qty < rankedPhrases[j].Qty {
			return false
		}

		if rankedPhrases[i].LeftID > rankedPhrases[j].LeftID {
			return true
		} else if rankedPhrases[i].LeftID < rankedPhrases[j].LeftID {
			return false
		}

		return rankedPhrases[i].RightID > rankedPhrases[j].RightID
	})

	phrases := make([]string, len(rankedPhrases))
	for i, rankedPhrase := range rankedPhrases {
		phrases[i] = fmt.Sprintf("%s %s", rankedPhrase.Left, rankedPhrase.Right)
	}

	return phrases
}

// rank processes text to extract the top ranking phrases and words.
func (t *TextRanker) rankWords(text string) []string {
	text = normalizeText(text)

	tr := textrank.NewTextRank()
	tr.Populate(text, t.language, t.rule)
	tr.Ranking(t.algorithm)

	// Extract single words
	rankedWords := textrank.FindSingleWords(tr)

	// Stable sorting for words
	// This is required for consistent output and test reproducibility
	sort.Slice(rankedWords, func(i, j int) bool {
		if rankedWords[i].Weight > rankedWords[j].Weight {
			return true
		} else if rankedWords[i].Weight < rankedWords[j].Weight {
			return false
		}

		if rankedWords[i].Qty > rankedWords[j].Qty {
			return true
		} else if rankedWords[i].Qty < rankedWords[j].Qty {
			return false
		}

		return rankedWords[i].ID > rankedWords[j].ID
	})

	words := make([]string, len(rankedWords))
	for i, rankedWord := range rankedWords {
		words[i] = rankedWord.Word
	}

	return words
}
