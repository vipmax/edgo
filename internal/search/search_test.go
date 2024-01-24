package search

import (
	"fmt"
	"reflect"
	"runtime"
	"testing"
	"time"
)

func TestSearch(t *testing.T) {
	text := [][]rune{
		[]rune("Hello, World!"),
		[]rune("This is a test"),
		[]rune("Another test"),
	}

	testCases := []struct {
		pattern  string
		expected []SearchResult
	}{
		{
			pattern: "test",
			expected: []SearchResult{
				{1, 10},
				{2, 8},
			},
		},
		{
			pattern:  "",
			expected: []SearchResult{},
		},
		{
			pattern:  "foo",
			expected: []SearchResult{},
		},
		{
			pattern: "Hello",
			expected: []SearchResult{
				{0, 0},
			},
		},
		{
			pattern: "is",
			expected: []SearchResult{
				{1, 2}, {1, 5},
			},
		},
		{
			pattern: "o",
			expected: []SearchResult{
				{0, 4}, {0, 8}, {2, 2},
			},
		},
		{
			pattern: "World",
			expected: []SearchResult{
				{0, 7},
			},
		},
	}

	for _, tc := range testCases {
		results := Search(text, tc.pattern)
		if !reflect.DeepEqual(results, tc.expected) {
			t.Errorf("Search did not return the expected results for pattern '%s'. Expected: %v, Got: %v",
				tc.pattern, tc.expected, results)
		}
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


func TestLineCount(t *testing.T) {
	start := time.Now()
	lc, emptylc := LineCountOnFile("search_test.go")
	elapsed := time.Since(start)

	fmt.Println("SearchOnFile done, elapsed", elapsed.String())
	fmt.Println("Found", lc, "lines")
	fmt.Println("Found empty", emptylc, "lines")
}


func TestLineCountOnDirParallel(t *testing.T) {
	fmt.Println("cpu", runtime.NumCPU())

	start := time.Now()
	path := "/Users/max/apps/go/edgo"
	results, totalFilesProcessed, totalRowsProcessed := LineCountOnDirParallel(path)
	elapsed := time.Since(start)

	fmt.Println("total files processed", totalFilesProcessed)
	fmt.Println("total rows processed", totalRowsProcessed)

	fmt.Println("done, elapsed", elapsed.String())
	//fmt.Println("Found", len(results), "results")
	//
	//for _, result := range results {
	//	fmt.Println(result)
	//}

	langCount := LangCount(results)
	for _, result := range langCount {
		fmt.Println(result.FilesCount, result.Lang, result.LinesCount)
	}
}