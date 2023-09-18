package io

import (
	"cmp"
	. "edgo/internal/logger"
	. "edgo/internal/search"
	. "edgo/internal/utils"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"
)

type FileInfo struct {
	Name      string
	FullName  string
	OpenCount int
	IsDir     bool
	IsDirOpen bool
	Childs    []FileInfo
	Level     int
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

	SortTree(fileInfo)

	return fileInfo, nil
}

func SortTree(fileInfo FileInfo) {
	slices.SortFunc(fileInfo.Childs, func(a, b FileInfo) int {
		if a.IsDir == true && b.IsDir == false {
			return -1
		}
		if a.IsDir == false && b.IsDir == true {
			return 1
		}
		return cmp.Compare(a.Name, b.Name)
	})
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

func FindByFullName(node *FileInfo, fileName string) *FileInfo {
	if node == nil { return nil }
	if node.FullName == fileName { return node }

	for i := range node.Childs {
		foundNode := FindByFullName(&node.Childs[i], fileName)
		if foundNode != nil { return foundNode }
	}

	return nil
}


