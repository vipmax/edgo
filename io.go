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
			e.colors = highlighter.colorize(code, e.filename);
			e.drawEverything();
			e.screen.Show()

		}()

	} else {
		code = e.buildContent(fileToRead, 1000000)
	}
	return code
}

func (e *Editor) writeFile() {

	// Create a new file, or open it if it exists
	f, err := os.Create(e.absoluteFilePath)
	if err != nil { panic(err) }

	// Create a buffered writer from the file
	w := bufio.NewWriter(f)

	for i, row := range e.content {
		for j := 0; j < len(row); {
			if _, err := w.WriteRune(row[j]); err != nil { panic(err) }
			j++
		}

		if i != len(e.content) - 1 { // do not write \n at the end
			if _, err := w.WriteRune('\n'); err != nil { panic(err) }
		}

	}

	// Don't forget to flush the buffered writer to ensure all data is written
	if err := w.Flush(); err != nil { panic(err) }
	if err := f.Close(); err != nil { panic(err) }

	e.isContentChanged = false

	if e.lang != "" && lsp.IsLangReady(e.lang) {
		go lsp.didOpen(e.absoluteFilePath, e.lang) // todo remove it in future
		//go lsp.didChange(absoluteFilePath)
		//go lsp.didSave(absoluteFilePath)
	}

}

func (e *Editor) buildContent(filename string, limit int) string {
	//start := time.Now()
	//logger.info("read file start", filename, string(limit))
	//defer logger.info("read file end",   filename, string(limit), "elapsed", time.Since(start).String())

	file, err := os.Open(filename)
	if err != nil {
		filec, err2 := os.Create(filename)
		if err2 != nil {fmt.Printf("Failed to create file: %v\n", err2)}
		defer filec.Close()
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)

	e.content = make([][]rune, 0)
	e.colors = make([][]int, 0)

	for scanner.Scan() {
		var line = scanner.Text()
		var lineChars = []rune{}
		for _, char := range line { lineChars = append(lineChars, char) }
		e.content = append(e.content, lineChars)
		if len(e.content) > limit { break }
	}

	// if no e.content, consider it like one line for next editing
	if e.content == nil || len(e.content) == 0 {
		e.content = make([][]rune, 1)
		e.colors = make([][]int, 1)
	}

	return convertToString(e.content)
}


func (e *Editor) readUpdateFiles() {
	ignoreDirs := []string{
		".git", ".idea", "node_modules", "dist", "target", "__pycache__", "build",
		".DS_Store",
	}

	filesTree, err := getFiles("./", ignoreDirs)
	if err != nil { fmt.Printf("Unable to get files: %v\n", err); os.Exit(1) }

	if filesTree != nil {
		if e.files == nil && len(e.files) == 0 {
			e.files = make([]FileInfo, len(filesTree))
			for i, f := range filesTree {
				abs, _ := filepath.Abs(f)
				e.files[i] = FileInfo{f, abs, 0}
			}
		} else {
			originalFiles := make([]string, len(e.files))
			for i, f := range e.files {
				originalFiles[i] = f.filename
			}
			newFiles, deletedFiles := findNewAndDeletedFiles(originalFiles, filesTree)
			for _, f := range newFiles {
				abs, _ := filepath.Abs(f)
				e.files = append(e.files, FileInfo{f, abs, 0})
			}

			// Remove deleted files from originalFiles
			for i := 0; i < len(e.files); i++ {
				if contains(deletedFiles, e.files[i].filename) {
					e.files = remove(e.files, i)
					i-- // Adjust index after removal
				}
			}
		}
	}
}

func (e *Editor) updateFilesOpenStats(file string) {
	if e.files == nil || len(e.files) == 0 { return }

	for i := 0; i < len(e.files); i++ {
		ti := e.files[i]
		if file == ti.fullfilename {
			ti.openCount += 1
			e.files[i] = ti
			break
		}
	}

	sort.SliceStable(e.files, func(i, j int) bool {
		return e.files[i].openCount > e.files[j].openCount
	})
}