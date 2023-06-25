package main

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"sort"
)

func (e *Editor) readFile(fileToRead string) string {
	/// if file is big, read only first 1000 lines and read rest async
	fileSize := getFileSize(fileToRead)
	fileSizeMB := fileSize / (1024 * 1024) // Convert size to megabytes

	var code string
	if fileSizeMB >= 1 {
		//colorize = false
		code = e.buildContent(fileToRead, 1000)

		go func() { // sync?? no need yet
			code = e.buildContent(fileToRead, 1000000)
			code, _ = getFirstLines(code, 20000)
			e.Colors = highlighter.colorize(code, e.Filename);
			e.drawEverything();
			e.Screen.Show()

		}()

	} else {
		code = e.buildContent(fileToRead, 1000000)
	}
	return code
}

func (e *Editor) writeFile() {

	// Create a new file, or open it if it exists
	f, err := os.Create(e.AbsoluteFilePath)
	if err != nil { panic(err) }

	// Create a buffered writer from the file
	w := bufio.NewWriter(f)

	for i, row := range e.Content {
		for j := 0; j < len(row); {
			if _, err := w.WriteRune(row[j]); err != nil { panic(err) }
			j++
		}

		if i != len(e.Content) - 1 { // do not write \n at the end
			if _, err := w.WriteRune('\n'); err != nil { panic(err) }
		}

	}

	// Don't forget to flush the buffered writer to ensure all data is written
	if err := w.Flush(); err != nil { panic(err) }
	if err := f.Close(); err != nil { panic(err) }

	e.IsContentChanged = false

	if e.Lang != "" && lsp.IsLangReady(e.Lang) {
		go lsp.didOpen(e.AbsoluteFilePath, e.Lang) // todo remove it in future
		//go lsp.didChange(AbsoluteFilePath)
		//go lsp.didSave(AbsoluteFilePath)
	}

}

func (e *Editor) buildContent(filename string, limit int) string {
	//start := time.Now()
	//logger.info("read file start", Filename, string(limit))
	//defer logger.info("read file end",   Filename, string(limit), "elapsed", time.Since(start).String())

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

	// if no e.Content, consider it like one line for next editing
	if e.Content == nil || len(e.Content) == 0 {
		e.Content = make([][]rune, 1)
		e.Colors = make([][]int, 1)
	}

	return convertToString(e.Content)
}


func (e *Editor) readUpdateFiles() {
	ignoreDirs := []string{
		".git", ".idea", "node_modules", "dist", "target", "__pycache__", "build",
		".DS_Store",
	}

	filesTree, err := getFiles("./", ignoreDirs)
	if err != nil { fmt.Printf("Unable to get files: %v\n", err); os.Exit(1) }

	if filesTree != nil {
		if e.Files == nil && len(e.Files) == 0 {
			e.Files = make([]FileInfo, len(filesTree))
			for i, f := range filesTree {
				abs, _ := filepath.Abs(f)
				e.Files[i] = FileInfo{f, abs, 0}
			}
		} else {
			originalFiles := make([]string, len(e.Files))
			for i, f := range e.Files {
				originalFiles[i] = f.filename
			}
			newFiles, deletedFiles := findNewAndDeletedFiles(originalFiles, filesTree)
			for _, f := range newFiles {
				abs, _ := filepath.Abs(f)
				e.Files = append(e.Files, FileInfo{f, abs, 0})
			}

			// Remove deleted files from originalFiles
			for i := 0; i < len(e.Files); i++ {
				if contains(deletedFiles, e.Files[i].filename) {
					e.Files = remove(e.Files, i)
					i-- // Adjust index after removal
				}
			}
		}
	}
}

func (e *Editor) updateFilesOpenStats(file string) {
	if e.Files == nil || len(e.Files) == 0 { return }

	for i := 0; i < len(e.Files); i++ {
		ti := e.Files[i]
		if file == ti.fullfilename {
			ti.openCount += 1
			e.Files[i] = ti
			break
		}
	}

	sort.SliceStable(e.Files, func(i, j int) bool {
		return e.Files[i].openCount > e.Files[j].openCount
	})
}