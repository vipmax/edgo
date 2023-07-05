package search

import (
	"fmt"
	"runtime"
	"testing"
	"time"
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
			gotLine, gotPos := SearchDown(tc.text, tc.pattern, tc.startLine)
			if gotLine != tc.wantLine || gotPos != tc.wantPos {
				t.Errorf("search() got %v, %v; want %v, %v", gotLine, gotPos, tc.wantLine, tc.wantPos)
			}
		})
	}
}

func TestSearchOnFile(t *testing.T) {
	start := time.Now()
	results, _ := SearchOnFile("search_test.go", "text")
	elapsed := time.Since(start)

	fmt.Println("SearchOnFile done, elapsed", elapsed.String())
	fmt.Println("Found", len(results), "results")

	for _, searchResult := range results { fmt.Println(searchResult) }
}

func TestSearchOnDir(t *testing.T) {
	start := time.Now()
	results, _ := SearchOnDir("/Users/max/apps/go/edgo/internal", "text")
	elapsed := time.Since(start)

	fmt.Println("SearchOnDir done, elapsed", elapsed.String())
	fmt.Println("Found", len(results), "results")

	for _, searchResult := range results { fmt.Println(searchResult) }
}

func TestSearchOnDirParallel(t *testing.T) {
	fmt.Println("CPU", runtime.NumCPU())

	start := time.Now()
	results, _, _ := SearchOnDirParallel("/Users/max/apps/go/edgo/internal", "text")
	elapsed := time.Since(start)

	fmt.Println("SearchOnDirParallel done, elapsed", elapsed.String())
	fmt.Println("Found", len(results), "results")

	for _, searchResult := range results { fmt.Println(searchResult) }
}

func TestSearchOnDirParallel2(t *testing.T) {
	fmt.Println("CPU", runtime.NumCPU())

	start := time.Now()
	results, totalRowsProcessed, _ := SearchOnDirParallel("/Users/max/Downloads/spark-master", "def main")
	elapsed := time.Since(start)

	fmt.Println("totalRowsProcessed", totalRowsProcessed)

	fmt.Println("SearchOnDirParallel done, elapsed", elapsed.String())
	fmt.Println("Found", len(results), "results")

	for _, searchResult := range results {
		fmt.Println(searchResult)
	}
}