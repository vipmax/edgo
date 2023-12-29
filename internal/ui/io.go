package ui

import (
	"bufio"
	. "edgo/internal/highlighter"
	. "edgo/internal/utils"
	"fmt"
	"os"
	"sort"
)

func (e *Editor) ReadFile(fileToRead string) string {
	/// if file is big, read only first 1000 lines and read rest async
	fileSize := GetFileSize(fileToRead)
	fileSizeMB := fileSize / (1024 * 1024) // Convert size to megabytes

	var code string
	if fileSizeMB >= 1 {
		//colorize = false
		code = e.BuildContent(fileToRead, 1000)

		go func() { // sync?? no need yet
			code = e.BuildContent(fileToRead, 1000000)
			code, _ = GetFirstLines(code, 20000)
			e.Colors = HighlighterGlobal.Colorize(code, e.Filename)
			e.DrawEverything()
			e.Screen.Show()
		}()

	} else {
		code = e.BuildContent(fileToRead, 1000000)
	}
	return code
}

func (e *Editor) WriteFile() {
	//to much cpu usage for big files
	//added, removed := Diff(e.LastCommitFileContent, ConvertContentToString(e.Content))
	//e.Added = added
	//e.Removed = removed

	// Create a new file, or open it if it exists
	f, err := os.Create(e.AbsoluteFilePath)
	if err != nil { panic(err) }

	// Create a buffered writer from the file
	w := bufio.NewWriter(f)

	for _, row := range e.Content {
		for j := 0; j < len(row); {
			if _, err := w.WriteRune(row[j]); err != nil { panic(err) }
			j++
		}

		//if i != len(e.Content) - 1 { // do not write \n at the end
			if _, err := w.WriteRune('\n'); err != nil { panic(err) }
		//}

	}

	// Don't forget to flush the buffered writer to ensure all data is written
	if err := w.Flush(); err != nil { panic(err) }
	if err := f.Close(); err != nil { panic(err) }

	e.IsContentChanged = false
	e.FileWatcher.UpdateStats()

	if e.Lang != "" {
		lsp := e.lsp2lang[e.Lang]
		if lsp.IsReady {
			go lsp.DidOpen(e.AbsoluteFilePath, e.Lang) // todo remove it in future
			//go lsp.didChange(AbsoluteFilePath)
			//go lsp.didSave(AbsoluteFilePath)
		}

	}
}

func (e *Editor) BuildContent(filename string, limit int) string {
	//Start := time.Now()
	//Log.info("read file Start", Name, string(limit))
	//defer Log.info("read file end",   Name, string(limit), "elapsed", time.Since(Start).String())

	file, err := os.Open(filename)
	if err != nil {
		filec, err2 := os.Create(filename)
		if err2 != nil {fmt.Printf("Failed to create file: %v\n", err2)}
		defer filec.Close()
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)

	e.Content = make([][]rune, 0)
	e.Colors = make([][]int, 0)

	for scanner.Scan() {
		var line = scanner.Text()
		var lineChars = []rune{}
		for _, char := range line { lineChars = append(lineChars, char) }
		e.Content = append(e.Content, lineChars)
		if len(e.Content) > limit { break }
	}

	// if no e.Content, consider it like one Line for next editing
	if e.Content == nil || len(e.Content) == 0 {
		e.Content = make([][]rune, 1)
		e.Colors = make([][]int, 1)
	}

	return ConvertContentToString(e.Content)
}

func (e *Editor) ReadContent(filename string, fromline int, toline int) [][]rune {

	file, err := os.Open(filename)
	if err != nil { return nil }
	defer file.Close()

	content := make([][]rune, 0)

	var i = 0
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		var line = scanner.Text()
		if i < fromline { i++; continue }
		var lineChars = []rune{}
		for _, char := range line { lineChars = append(lineChars, char) }
		content = append(content, lineChars)
		if i > toline { break }
		i++
	}

	return content
}

func (e *Editor) UpdateFilesOpenStats(file string) {
	if e.Files == nil || len(e.Files) == 0 { return }

	for i := 0; i < len(e.Files); i++ {
		ti := e.Files[i]
		if file == ti.FullName {
			ti.OpenCount += 1
			e.Files[i] = ti
			break
		}
	}

	sort.SliceStable(e.Files, func(i, j int) bool {
		return e.Files[i].OpenCount > e.Files[j].OpenCount
	})
}

