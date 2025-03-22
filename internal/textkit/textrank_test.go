// Copyright (c) VirtualTam
// SPDX-License-Identifier: MIT

package textkit_test

import (
	"os"
	"testing"

	"github.com/virtualtam/sparklemuffin/internal/textkit"
)

func TestTextRanker(t *testing.T) {
	const topN = 20

	wikipediaDowntownSeattleTransitTunnel, err := os.ReadFile("testdata/wikipedia/Downtown_Seattle_Transit_Tunnel.txt")
	if err != nil {
		t.Fatalf("failed to read test input: %q", err)
	}

	wikipediaHistoryOfAluminium, err := os.ReadFile("testdata/wikipedia/History_of_aluminium.txt")
	if err != nil {
		t.Fatalf("failed to read test input: %q", err)
	}

	cases := []struct {
		tname           string
		document        string
		wantTopNPhrases []string
		wantTopNWords   []string
	}{
		{
			tname: "empty document",
		},
		{
			tname:    "Wikipedia: Downtown Seattle Transit Tunnel",
			document: string(wikipediaDowntownSeattleTransitTunnel),
			wantTopNPhrases: []string{
				// High-ranking phrases (core topic)
				"light rail",
				"avenue 3rd",
				"transit sound",
				"tunnel bus",
				"king county",
				"street pine",
				"metro county",
				"tunnel transit",
				"seattle downtown",
				"dual mode",

				// Lower-ranking phrases (context)
				"tunnel stations",
				"metro council",
				"place convention",
				"transit rapid",
				"seattle transit",
				"project tunnel",
				"service rail",
				"station place",
				"city seattle",
				"light link",
			},
			wantTopNWords: []string{
				// High-ranking words (core topic)
				"tunnel",
				"transit",
				"metro",
				"bus",
				"station",
				"seattle",
				"avenue",
				"street",
				"rail",
				"buses",

				// Lower-ranking words (context)
				"light",
				"million",
				"service",
				"stations",
				"3rd",
				"sound",
				"downtown",
				"south",
				"king",
				"project",
			},
		},
		{
			tname:    "Wikipedia: History of Aluminium",
			document: string(wikipediaHistoryOfAluminium),
			wantTopNPhrases: []string{
				// High-ranking phrases (core topic)
				"production aluminium",
				"tons metric",
				"states united",
				"aluminium produced",
				"german chemist",
				"metric 000",
				"chemist french",
				"aluminium recycling",
				"production industrial",
				"heroult hall",

				// Lower-ranking phrases (context)
				"production world",
				"price real",
				"aluminium used",
				"aluminium primary",
				"aluminium bronze",
				"acid sulfuric",
				"states dollars",
				"combined share",
				"non ferrous",
				"metric ton",
			},
			wantTopNWords: []string{
				// High-ranking words (core topic)
				"aluminium",
				"production",
				"metal",
				"alum",
				"chemist",
				"alumina",
				"deville",
				"used",
				"metric",
				"earth",

				// Lower-ranking words (context)
				"world",
				"hall",
				"tons",
				"united",
				"states",
				"price",
				"recycling",
				"produced",
				"heroult",
				"time",
			},
		},
	}

	textRanker := textkit.NewTextRanker()

	t.Run("RankTopNPhrases", func(t *testing.T) {
		for _, tc := range cases {
			t.Run(tc.tname, func(t *testing.T) {
				phrases := textRanker.RankTopNPhrases(tc.document, topN)

				assertStringSlicesEqual(t, phrases, tc.wantTopNPhrases)
			})
		}
	})

	t.Run("RankTopNWords", func(t *testing.T) {
		for _, tc := range cases {
			t.Run(tc.tname, func(t *testing.T) {
				words := textRanker.RankTopNWords(tc.document, topN)

				assertStringSlicesEqual(t, words, tc.wantTopNWords)
			})
		}
	})
}

func assertStringSlicesEqual(t *testing.T, got, want []string) {
	t.Helper()

	if len(got) != len(want) {
		t.Errorf("want %#v, got %#v", want, got)
		t.Fatalf("want %d items, got %d", len(want), len(got))
	}

	for i, wantItem := range want {
		if got[i] != wantItem {
			t.Errorf("want item %d %q, got %q", i, wantItem, got[i])
		}
	}
}
