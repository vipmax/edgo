package search

import (
	"bufio"
	. "edgo/internal/logger"
	"edgo/internal/utils"
	"path/filepath"
	"runtime"
	"sync"

	"os"
	"strings"
	"time"
)

/*  strings.Index() function in Go's standard library is implemented
    in a highly optimized way to ensure efficient and fast string searching.
    Under the hood, it uses some sophisticated algorithms like the
    Boyer-Moore string search algorithm or its variations, which have much better
    time complexity (nearly linear time in many cases) than the simple linear search.
*/
func SearchDown(text [][]rune, pattern string, startLine int) (int, int) {
	start := time.Now()
	defer Log.Info("search up end, elapsed:", time.Since(start).String())

	if len(pattern) == 0 { return -1, -1 }
	if startLine < 0 || startLine >= len(text) { return -1, -1 }

	for i := startLine; i < len(text); i++ {
		line := string(text[i])
		pos := strings.Index(line, pattern)
		if pos != -1 { return i, pos }
	}
	return -1, -1
}

func SearchUp(text [][]rune, pattern string, startLine int) (int, int) {
	start := time.Now()
	defer Log.Info("search up end, elapsed:", time.Since(start).String())

	if len(pattern) == 0 { return -1, -1 }
	if startLine < 0 || startLine >= len(text) { return -1, -1 }

	for i := startLine; i >= 0; i-- {
		line := string(text[i])
		pos := strings.Index(line, pattern)
		if pos != -1 { return i, pos }
	}
	return -1, -1
}

type SearchResult struct {
	Line     int
	Position int
}


func SearchOnFile(filename string, pattern string) ([]SearchResult, int) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, 0
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)

	results := []SearchResult{}

	var lineindex = 1
	for scanner.Scan() {
		var line = scanner.Text()

		pos := strings.Index(line, pattern)
		if pos != -1 {
			searchResult := SearchResult{lineindex, pos}
			results = append(results, searchResult)
		}
		lineindex++
	}

	return results, lineindex
}

type FileSearchResult struct {
	File    string
	Results []SearchResult
}

func SearchOnDir(dir string, pattern string) ([]FileSearchResult, int) {
	var files []string
	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil { return err }
		if !info.IsDir() { files = append(files, path) }
		return nil
	})

	if err != nil {
		return []FileSearchResult{}, 0
	}

	results := []FileSearchResult{}

	linesCount := 0
	for _, file := range files {
		fileResults, lc := SearchOnFile(file, pattern)
		linesCount += lc
		if len(fileResults) > 0 {
			results = append(results, FileSearchResult{file, fileResults})
		}
	}

	return results, linesCount
}


func SearchOnDirParallel(dir string, pattern string) ([]FileSearchResult, int, int) {
	var files []string

	var IgnoreDirs = []string{
		".git", ".idea", "node_modules", "dist", "target", "__pycache__", "build",
		".DS_Store", ".venv", "venv",
	}


	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil { return err }
		if info.IsDir() && utils.IsIgnored(info.Name(), IgnoreDirs) { return filepath.SkipDir }
		if !info.IsDir() { files = append(files, path) }
		return nil
	})

	if err != nil { return []FileSearchResult{}, 0, 0 }

	var wg sync.WaitGroup
	resultCh := make(chan FileSearchResult, len(files))
	rowsProcessedCh := make(chan int, len(files))
	sem := make(chan struct{}, runtime.NumCPU())

	for _, file := range files {
		wg.Add(1)
		sem <- struct{}{}
		go func(file string) {
			defer wg.Done()
			fileResults, rowsProcessed := SearchOnFile(file, pattern)
			resultCh <- FileSearchResult{file, fileResults}
			rowsProcessedCh <- rowsProcessed
			<-sem
		}(file)
	}

	// Start a go routine to close the resultCh after all other go routines are done
	go func() {
		wg.Wait()
		close(resultCh)
		close(rowsProcessedCh)
	}()

	results := []FileSearchResult{}
	filesProcessedCount := 0
	for result := range resultCh {
		filesProcessedCount++
		if len(result.Results) > 0 {
			results = append(results, result)
		}
	}

	totalRowsProcessed := 0
	for rows := range rowsProcessedCh {
		totalRowsProcessed += rows
	}

	return results, filesProcessedCount, totalRowsProcessed
}