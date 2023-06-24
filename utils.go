package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
)

func max(x, y int) int {
	if x < y {
		return y
	}
	return x
}

func min(x, y int) int {
	if x <= y {
		return x
	}
	return y
}

func maxMany(nums ...int) int {
	if len(nums) == 0 {
		panic("max: no arguments provided")
	}
	maxValue := nums[0]
	for _, num := range nums[1:] {
		if num > maxValue {
			maxValue = num
		}
	}
	return maxValue
}
func minMany(nums ...int) int {
	if len(nums) == 0 {
		panic("min: no arguments provided")
	}
	minValue := nums[0]
	for _, num := range nums[1:] {
		if num < minValue {
			minValue = num
		}
	}
	return minValue
}

func insert[T any](a []T, index int, value T) []T {
	n := len(a)
	if index < 0 {
		index = (index%n + n) % n
	}
	switch {
	case index == n: // nil or empty slice or after last element
		return append(a, value)

	case index < n: // index < len(a)
		a = append(a[:index+1], a[index:]...)
		a[index] = value
		return a

	case index < cap(a): // index > len(a)
		a = a[:index+1]
		var zero T
		for i := n; i < index; i++ {
			a[i] = zero
		}
		a[index] = value
		return a

	default:
		b := make([]T, index+1) // malloc
		if n > 0 {
			copy(b, a)
		}
		b[index] = value
		return b
	}
}

var matched = []rune{
	' ', '.', ',', '=', '+', '-', '[', '(', '{', ']', ')', '}', '"', ':', '&', '?','!',';','\t',
}

func findNextWord(chars []rune, from int) int {
	// Find the next word index after the specified index
	for i := from; i < len(chars); i++ {
		if contains(matched, chars[i]) {
			return i
		}
	}

	return len(chars)
}

func findPrevWord(chars []rune, from int) int {
	// Find the previous word index before the specified index
	for i := from - 1; i >= 0; i-- {
		if contains(matched, chars[i]) {
			return i + 1
		}
	}

	return 0
}

func contains[T comparable](slice []T, e T) bool {
	for _, val := range slice {
		if val == e {
			return true
		}
	}
	return false
}

func remove[T any](slice []T, s int) []T {
	return append(slice[:s], slice[s+1:]...)
}



func maxString(arr []string) int {
	maxLength := 0
	for _, str := range arr {
		if len(str) > maxLength {
			maxLength = len(str)
		}
	}
	return maxLength
}

func readFileToString(filePath string) (string, error) {
	filecontent, err := os.ReadFile(filePath)
	if err != nil { return "", err }
	return string(filecontent), nil
}

func convertToString(content [][]rune) string {
	var result strings.Builder
	for i, row := range content {
		for _, ch := range row { result.WriteRune(ch) }
		if i != len(content)-1 { result.WriteByte('\n') }
	}
	return result.String()
}

func countTabs(str []rune, stopIndex int) int {
	if stopIndex == 0 { return 0 }

	count := 0
	for i, char := range str {
		if i >= stopIndex { break }
		if char == '\t' { count++ } else { break }
	}
	return count
}
func countTabsTo(str []rune, stopIndex int) int {
	if stopIndex == 0 { return 0 }

	count := 0
	for i, char := range str {
		if i >= stopIndex { break }
		if char == '\t' { count++ }
	}
	return count
}
func countSpaces(str []rune, stopIndex int) int {
	if stopIndex == 0 { return 0 }

	count := 0
	for i, char := range str {
		if i >= stopIndex { break }
		if char == ' ' { count++ } else { break }
	}
	return count
}

func formatText(left, right string, maxWidth int) string {
	left = fmt.Sprintf("%-*s", maxWidth, left)
	right = fmt.Sprintf("%s",  right)
	return fmt.Sprintf("%s %s", left, right)
}

func getFileSize(filename string) int64 {
	file, err := os.Open(filename) // replace with your file name
	if err != nil { return 0 }
	defer file.Close()

	fileInfo, err := file.Stat()
	if err != nil { return 0 }

	fileSize := fileInfo.Size() // get the size in bytes
	return fileSize
}

func centerNumber(brw int, width int) string {
	lineNumber := strconv.Itoa(brw )
	padding := width - len(lineNumber)
	leftPad := fmt.Sprintf("%*s", padding/2, "")
	rightPad := fmt.Sprintf("%*s", padding-(padding/2), "")
	lineNumber = leftPad + lineNumber + rightPad
	return lineNumber
}

type Set map[int]struct{}

// a value to the set.
func (this Set) Add(value int) { this[value] = struct{}{} }

// returns all keys in the set, sorted.
func (this Set) GetKeys() []int {
	keys := make([]int, 0, len(this))
	for key := range this { keys = append(keys, key) }
	sort.Ints(keys) // Sort the keys
	return keys
}

func PadLeft(str string, length int) string {
	format := fmt.Sprintf("%%%ds", length)
	return fmt.Sprintf(format, str)
}

func getFirstLines(s string, lineNum int) (string, error) {
	scanner := bufio.NewScanner(strings.NewReader(s))
	count := 0
	var builder strings.Builder
	for scanner.Scan() {
		builder.WriteString(scanner.Text())
		builder.WriteString("\n")
		count++
		if count == lineNum {
			break
		}
	}

	if scanner.Err() != nil {
		// handle error.
		return "", scanner.Err()
	}

	return builder.String(), nil
}

//func isIgnored(dir string, ignoreDirs []string) bool {
//	for _, ignore := range ignoreDirs {
//		if dir == ignore {
//			return true
//		}
//	}
//	return false
//}
func isIgnored(path string, ignorePatterns []string) bool {
	for _, pattern := range ignorePatterns {
		match, err := filepath.Match(pattern, filepath.Base(path))
		if err != nil { log.Println("Invalid pattern:", pattern); continue }
		if match { return true }
	}
	return false
}

func getFiles(path string, ignoreDirs []string) ([]string, error) {
	var files []string
	err := filepath.Walk(path, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			dir := filepath.Base(path)
			if isIgnored(dir, ignoreDirs) {
				return filepath.SkipDir
			}
		} else {
			if !isIgnored(path, ignoreDirs) {
				files = append(files, path)
			}

		}
		return nil
	})
	return files, err
}

const Phi = 1.61803398875 // The Golden Ratio

func goldenRatioPartition(totalSize int) (a int, b int) {
	b = int(float64(totalSize) / (Phi + 1))
	a = totalSize - b
	return
}

func findNewAndDeletedFiles(originalFiles []string, newFiles []string) ([]string, []string) {
     
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