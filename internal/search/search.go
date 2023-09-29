package search

import (
	"bufio"
	. "edgo/internal/logger"
	"edgo/internal/utils"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"
)

/*  strings.Index() function in Go's standard library is implemented
    in a highly optimized way to ensure efficient and fast string searching.
    Under the hood, it uses some sophisticated algorithms like the
    Boyer-Moore string search algorithm or its variations, which have much better
    time complexity (nearly linear time in many cases) than the simple linear search.
*/

func SearchDown(text [][]rune, pattern string, startLine int, startcol int) (int, int) {
	start := time.Now()
	defer Log.Info("search up end, elapsed:", time.Since(start).String())

	if len(pattern) == 0 { return -1, -1 }
	if startLine < 0 || startLine >= len(text) { return -1, -1 }


	for i := startLine; i < len(text); i++ {
		line := string(text[i])
		if startcol < 0 || startcol >= len(line) { continue }
		if startcol > 0 { line = line[startcol:] }

		pos := strings.Index(line, pattern)
		if pos != -1 { return i, pos+startcol }
		startcol = 0
	}
	return -1, -1
}

type SearchResult struct {
	Line     int
	Position int
}

func Search(text [][]rune, pattern string) []SearchResult {
	results := []SearchResult{}

	if len(pattern) == 0 || len(text) == 0 { return results }

	for i := 0; i < len(text); i++ {
		from := 0
		line := string(text[i])
		for {
			pos := strings.Index(line[from:], pattern)
			if pos == -1 { break } else {
				pos = from + pos
				results = append(results, SearchResult{i, pos})
				from = pos + 1
			}
		}
	}
	return results
}


func SearchOnFile(filename string, pattern string) ([]SearchResult, int) {
	file, err := os.Open(filename)
	if err != nil { return nil, 0 }
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

func ParsePattern(pattern string) (string, []string) {
	extensions := make([]string, 0)

	if strings.Contains(pattern, " -f ") {
		split := strings.Split(pattern, " -f ")
		pattern = strings.TrimSpace(split[0])
		fileExtensions := split[1]
		extensionList := strings.Split(fileExtensions, ",")
		return pattern, extensionList
	}

	return pattern, extensions
}

func SearchOnDir(dir string, pattern string) ([]FileSearchResult, int) {
	var files []string

	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil { return err }
		if !info.IsDir() {
			files = append(files, path)
		}
		return nil
	})

	if err != nil { return []FileSearchResult{}, 0 }

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


var IgnoreDirs = []string{
	".git", ".idea", "node_modules", "dist", "target", "__pycache__", "build",
	".DS_Store", ".venv", "venv",
}

var IgnoreExts = []string{ "",
	".doc", ".docx", ".pdf", ".rtf", ".odt", ".xlsx", ".pptx",
	".jpg", ".png", ".gif", ".bmp", ".svg", ".tiff",
	".mp3", ".wav", ".aac", ".flac", ".ogg",
	".mp4", ".avi", ".mov", ".wmv", ".mkv",
	".zip", ".rar", ".tar.gz", ".7z",
	".exe", ".msi", ".bat", ".sh",
	".ttf", ".otf",
}

func SearchOnDirParallel(dir string, pattern string) ([]FileSearchResult, int, int) {
	var files []string

	searchPattern, allowedExtensions := ParsePattern(pattern)

	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil { return err }
		if info.IsDir() && utils.IsIgnored(info.Name(), IgnoreDirs) { return filepath.SkipDir }
	
		if !info.IsDir() && !utils.IsMatchExt(info.Name(), IgnoreExts) {
			if len(allowedExtensions) > 0 {
				if utils.IsMatchExt(info.Name(), allowedExtensions) {
					files = append(files, path)
				}
			} else {
				files = append(files, path)
			}
		}
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
			fileResults, rowsProcessed := SearchOnFile(file, searchPattern)
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
