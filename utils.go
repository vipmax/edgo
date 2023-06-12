package main

import (
	"fmt"
	"github.com/alecthomas/chroma/lexers"
	"os"
	"strconv"
	"strings"
	"unicode/utf8"
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

var matched = []rune{' ', '.', ',', '=', '+', '-', '[', '(', '{', '"'}

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

func GreaterThan(x, y, x1, y1 int) bool {
	if y > y1 {
		return true
	}
	return y == y1 && x > x1
}

func LessThan(x, y, x1, y1 int) bool {
	if y < y1 {
		return true
	}
	return y == y1 && x < x1
}

func GreaterEqual(x, y, x1, y1 int) bool {
	if y > y1 {
		return true
	}
	if y == y1 && x >= x1 {
		return true
	}
	return false
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

func countTabsFromString(str string, stopIndex int) int {
	count := 0
	for i, char := range str {
		if i > stopIndex { break }
		if char == '\t' { count++ }
	}
	return count
}

func formatText(left, right string, maxWidth int) string {
	left = fmt.Sprintf("%-*s", maxWidth, left)
	right = fmt.Sprintf("%s",  right)
	return fmt.Sprintf("%s %s", left, right)
}
func limitString(s string, maxLength int) string {
	if utf8.RuneCountInString(s) <= maxLength { return s }
	runes := []rune(s)
	return string(runes[:maxLength])
}

func detectLang(filename string) string {
	lexer := lexers.Match(filename)
	if lexer == nil { return "" }
	config := lexer.Config()
	if config == nil { return "" }
	return strings.ToLower(config.Name)
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
