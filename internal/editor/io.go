package editor

import (
	"bufio"
	"cmp"
	. "edgo/internal/highlighter"
	. "edgo/internal/logger"
	. "edgo/internal/search"
	. "edgo/internal/utils"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"sort"
	"strings"
	"sync"
	"time"
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
			e.Colors = HighlighterGlobal.Colorize(code, e.Filename);
			e.DrawEverything();
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
	e.FileWatcher.Update()

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

func GetFiles(path string, ignoreDirs []string) ([]string, error) {
	var files []string
	err := filepath.Walk(path, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			dir := filepath.Base(path)
			if IsIgnored(dir, ignoreDirs) {
				return filepath.SkipDir
			}
		} else {
			if !IsIgnored(path, ignoreDirs) {
				files = append(files, path)
			}

		}
		return nil
	})
	return files, err
}


func (e *Editor) ReadFilesUpdate() {

	filesTree, err := GetFiles("./", IgnoreDirs)
	if err != nil { fmt.Printf("Unable to get files: %v\n", err); os.Exit(1) }

	if filesTree != nil {
		if e.Files == nil && len(e.Files) == 0 {
			e.Files = make([]FileInfo, len(filesTree))
			for i, f := range filesTree {
				abs, _ := filepath.Abs(f)
				e.Files[i] = FileInfo{f, abs, 0, false, false, nil, 0 }
			}
		} else {
			originalFiles := make([]string, len(e.Files))
			for i, f := range e.Files {
				originalFiles[i] = f.Name
			}
			newFiles, deletedFiles := FindNewAndDeletedFiles(originalFiles, filesTree)
			for _, f := range newFiles {
				abs, _ := filepath.Abs(f)
				e.Files = append(e.Files, FileInfo{f, abs, 0, false, false, nil, 0})
			}

			// Remove deleted files from originalFiles
			for i := 0; i < len(e.Files); i++ {
				if Contains(deletedFiles, e.Files[i].Name) {
					e.Files = Remove(e.Files, i)
					i-- // Adjust index after removal
				}
			}
		}
	}
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


func FindNewAndDeletedFiles(originalFiles []string, newFiles []string) ([]string, []string) {

	originalFilesMap := make(map[string]bool, len(originalFiles))
	newFilesMap := make(map[string]bool, len(newFiles))

	// Add original files to map
	for _, file := range originalFiles { originalFilesMap[file] = true }

	// Add new files to map
	for _, file := range newFiles { newFilesMap[file] = true }

	// Check for new files
	var newlyCreated []string
	for _, file := range newFiles {
		if !originalFilesMap[file] {
			newlyCreated = append(newlyCreated, file)
		}
	}

	// Check for deleted files
	var deleted []string
	for _, file := range originalFiles {
		if !newFilesMap[file] {
			deleted = append(deleted, file)
		}
	}

	return newlyCreated, deleted
}




type FileInfo struct {
	Name      string
	FullName  string
	OpenCount int
	IsDir     bool
	IsDirOpen bool
	Childs    []FileInfo
	Level     int
}

func findMaxByFilenameLength(files []FileInfo) int {
	maxLength := 0
	var maxFile FileInfo

	for _, file := range files {
		if len(file.Name) > maxLength {
			maxLength = len(file.Name)
			maxFile = file
		}
	}

	return len(maxFile.Name)
}

func IsFileExists(filename string) bool {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}


func ReadDirTree(dirPath string, filter string, isOpen bool, level int) (FileInfo, error) {
	fileInfo := FileInfo{
		Name:      filepath.Base(dirPath),
		FullName:  dirPath,
		IsDir:     true,
		IsDirOpen: isOpen,
		Level: level,
	}

	// Read directory contents
	files, err := os.ReadDir(dirPath)
	if err != nil { return fileInfo, err }

	for _, file := range files {
		childPath := filepath.Join(dirPath, file.Name())

		if file.IsDir() && ! IsIgnored(file.Name(), IgnoreDirs) {
			childInfo, err2 := ReadDirTree(childPath, filter, isOpen, level + 1)
			if err2 != nil {
				Log.Info("Failed to process directory:", err2.Error())
				continue
			}
			fileInfo.Childs = append(fileInfo.Childs, childInfo)
		} else {
			foundMatch :=  strings.Contains(file.Name(), filter)
			if filter == "" || foundMatch {
				childInfo := FileInfo{
					Name:      file.Name(),
					FullName:  childPath,
					OpenCount: 0,
					IsDir:     false,
					IsDirOpen: false,
					Level:     level+1,
				}
				fileInfo.Childs = append(fileInfo.Childs, childInfo)
			}
		}
	}

	slices.SortFunc(fileInfo.Childs, func(a, b FileInfo) int {
		if a.IsDir == true && b.IsDir == false { return -1 }
		if a.IsDir == false && b.IsDir == true { return 1 }
		return cmp.Compare(a.Name, b.Name)
	})

	return fileInfo, nil
}

func PrintTree(fileInfo FileInfo, indent int) {
	// Print the current file/directory
	fmt.Printf("%s%s\n", strings.Repeat(" ", indent), fileInfo.Name)

	// Print child files/directories recursively
	for _, child := range fileInfo.Childs {
		PrintTree(child, indent+1)
	}
}

func TreeSize(fileInfo FileInfo, size int) int {
	size += 1

	if !fileInfo.IsDirOpen { return size }
	for _, child := range fileInfo.Childs {
		size += TreeSize(child, 0)
	}

	return  size
}

func GetSelected(fileInfo FileInfo, selected int) (bool, *FileInfo) {
	var i = 0
	found, info := countSelected(fileInfo, selected, &i)
	return found, info
}

func countSelected(fileInfo FileInfo, selected int, i *int) (bool, *FileInfo) {
	if selected == *i {
		return true, &fileInfo
	}

	*i++

	for j := 0; j < len(fileInfo.Childs); j++ {
		var child = &fileInfo.Childs[j]
		if selected == *i {
			return true, child
		}

		if child.IsDir && child.IsDirOpen {
			found, fi := countSelected(*child, selected, i)
			if found {
				return found, fi
			}
		} else {
			*i++
		}
	}

	return false, &FileInfo{}
}
func FilterIfLeafEmpty(fileInfo FileInfo) FileInfo {
	fileInfo.Childs = filterChilds(fileInfo.Childs)
	return fileInfo
}

func filterChilds(childs []FileInfo) []FileInfo {
	filteredChilds := make([]FileInfo, 0)

	for _, child := range childs {
		// Leaf node with an empty name, filter it out
		if !child.IsDir && child.Name == "" { continue }

		if child.IsDir {
			child.Childs = filterChilds(child.Childs)
			if len(child.Childs) == 0 { continue }
		}
		filteredChilds = append(filteredChilds, child)
	}

	return filteredChilds
}

func FindFirstFile(fileInfo FileInfo, index int) (*FileInfo, int) {
	if !fileInfo.IsDir {
		// If the current node is a file, return it with the index
		return &fileInfo, index
	}

	// Recursively search for the first file in child nodes
	for i, child := range fileInfo.Childs {
		foundFile, foundIndex := FindFirstFile(child, index+i+1)
		if foundFile != nil {
			return foundFile, foundIndex
		}
	}

	// No file found in the hierarchy
	return nil, -1
}


func SetDirOpenFlag(root *FileInfo, fileName string) bool {
	if root == nil { return false }
	if root.FullName == fileName { return true }

	for i := range root.Childs {
		didSet := SetDirOpenFlag(&root.Childs[i], fileName)
		if didSet {
			root.IsDirOpen = true
			return true
		}
	}

	return false
}

type FileWatcher struct {
	filePath  string
	lastStats os.FileInfo
	ticker    *time.Ticker
	mu        sync.Mutex
}

func NewFileWatcher(everyms int) *FileWatcher {
	duration := time.Millisecond * time.Duration(everyms)

	fw := &FileWatcher{
		filePath:  "",
		lastStats: nil,
		ticker:    time.NewTicker(duration),
		mu:        sync.Mutex{},
	}
	return fw
}

func (fw *FileWatcher) StartWatch(f func() ) {
	go func() {
		for range fw.ticker.C {
			if fw.filePath == "" { continue }
			stats, _ := os.Stat(fw.filePath)

			if stats.Size() != fw.lastStats.Size() ||
			   stats.ModTime() != fw.lastStats.ModTime() {
				f()
				fw.Update()
			}
		}
	}()
}


func (fw *FileWatcher) UpdateFile(filePath string) {
	if fw.filePath == filePath { return }
	fw.mu.Lock()
	defer fw.mu.Unlock()
	fw.filePath = filePath
}


func (fw *FileWatcher) Update() {
	fw.mu.Lock()
	defer fw.mu.Unlock()
	newStats, _ := os.Stat(fw.filePath)
	fw.lastStats = newStats
}

func (fw *FileWatcher) Stop() {
	fw.mu.Lock()
	defer fw.mu.Unlock()
	fw.ticker.Stop()
}