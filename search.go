package main

import (
	"strings"
	"time"
)

/*  strings.Index() function in Go's standard library is implemented
    in a highly optimized way to ensure efficient and fast string searching.
    Under the hood, it uses some sophisticated algorithms like the
    Boyer-Moore string search algorithm or its variations, which have much better
    time complexity (nearly linear time in many cases) than the simple linear search.
*/
func searchDown(text [][]rune, pattern string, startLine int) (int, int) {
	start := time.Now()
	defer logger.info("search up end, elapsed:", time.Since(start).String())

	if len(pattern) == 0 { return -1, -1 }
	if startLine < 0 || startLine >= len(text) { return -1, -1 }

	for i := startLine; i < len(text); i++ {
		line := string(text[i])
		pos := strings.Index(line, pattern)
		if pos != -1 { return i, pos }
	}
	return -1, -1
}

func searchUp(text [][]rune, pattern string, startLine int) (int, int) {
	start := time.Now()
	defer logger.info("search up end, elapsed:", time.Since(start).String())

	if len(pattern) == 0 { return -1, -1 }
	if startLine < 0 || startLine >= len(text) { return -1, -1 }

	for i := startLine; i >= 0; i-- {
		line := string(text[i])
		pos := strings.Index(line, pattern)
		if pos != -1 { return i, pos }
	}
	return -1, -1
}
