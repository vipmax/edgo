package main

import (
	"testing"
)

func TestSearch(t *testing.T) {
	text := [][]rune{
		[]rune("This is the first line."),
		[]rune("This is the second line."),
	}
	text2 := [][]rune{
		[]rune("This is the first line."),
		[]rune("This is the second line."),
		[]rune("This is the third line."),
	}

	tests := []struct {
		name   string
		text   [][]rune
		pattern string
		startLine int
		wantLine int
		wantPos int
	}{
		{
			name: "Test 1 - Single match",
			text: text, pattern: "second", startLine: 0, wantLine: 1, wantPos: 12,
		},
		{
			name: "Test 2 - No match",
			text: text, pattern: "third", startLine: 0, wantLine: -1, wantPos: -1,
		},
		{
			name: "Test 3 - Multiple matches",
			text: text, pattern: "the", startLine: 0, wantLine: 0, wantPos: 8,
		},
		{
			name: "Test 1 - Pattern found after start line",
			text: text2, pattern: "third", startLine: 1, wantLine: 2, wantPos: 12,
		},
		{
			name: "Test 2 - Pattern not found after start line",
			text: text2, pattern: "first", startLine: 1, wantLine: -1, wantPos: -1,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			gotLine, gotPos := searchDown(tc.text, tc.pattern, tc.startLine)
			if gotLine != tc.wantLine || gotPos != tc.wantPos {
				t.Errorf("search() got %v, %v; want %v, %v", gotLine, gotPos, tc.wantLine, tc.wantPos)
			}
		})
	}
}
