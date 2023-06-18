package main

import (
	"bufio"
	"fmt"
	"os"
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

func Equal(x, y, x1, y1 int) bool {
	return x == x1 && y == y1
}

func getSelectedIndices(content [][]rune, ssx, ssy, sex, sey int) [][]int {
	var selectedIndices = [][]int{}

	// check for empty selection
	if Equal(ssx, ssy, sex, sey) {
		return selectedIndices
	}

	// getting selection start point
	var startx, starty = ssx, ssy
	var endx, endy = sex, sey

	// swap points if selection is inversed
	if GreaterThan(startx, starty, endx, endy) {
		startx, endx = endx, startx
		starty, endy = endy, starty
	}

	var inside = false
	// iterate over content, starting from selection start point until out ouf selection
	for j := starty; j < len(content); j++ {
		for i := 0; i < len(content[j]); i++ {
			if isUnderSelection(i, j) {
				selectedIndices = append(selectedIndices, []int{i, j})
				inside = true
			} else  {
				if inside == true { // first time when out ouf selection
					return selectedIndices
				}
			}
		}
	}
	return selectedIndices
}

func getSelectionString(content [][]rune, ssx, ssy, sex, sey int) string {
	var ret = []rune {}
	var in = false

	// check for empty selection
	if Equal(ssx, ssy, sex, sey) { return "" }

	// getting selection start point
	var startx, starty = ssx, ssy
	var endx, endy = sex, sey

	if GreaterThan(startx, starty, endx, endy) {
		startx, endx = endx, startx // swap  points if selection inverse
		starty, endy = endy, starty
	}

	for j := starty; j < len(content); j++ {
		row := content[j]
		for i, char := range row {
			// if inside selection
			if GreaterEqual(i, j, startx, starty) && LessThan(i, j, endx, endy) {
				ret = append(ret, char)
				in = true
			} else {
				in = false
				// only one selection area can be, early return
				if len(ret) > 0 {
					// remove the last newline if present
					if len(ret) > 0 && ret[len(ret)-1] == '\n' { ret = ret[:len(ret)-1] }
					return string(ret)
				}
			}
		}
		if in && LessThan(0, j, endx, endy) {
			ret = append(ret, '\n')
		}
	}

	if len(ret) > 0 && ret[len(ret)-1] == '\n' { ret = ret[:len(ret)-1] }
	return string(ret)
}


func getSelectedLines(content [][]rune, ssx, ssy, sex, sey int)  []int {
	var lineNumbers = make(Set)
	var in = false

	// check for empty selection
	if Equal(ssx, ssy, sex, sey) { return lineNumbers.GetKeys() }

	// getting selection start point
	var startx, starty = ssx, ssy
	var endx, endy = sex, sey

	if GreaterThan(startx, starty, endx, endy) {
		startx, endx = endx, startx // swap  points if selection inverse
		starty, endy = endy, starty
	}

	for j := starty; j < len(content); j++ {
		row := content[j]
		for i, _ := range row {
			// if inside selection
			if GreaterEqual(i, j, startx, starty) && LessThan(i, j, endx, endy) {
				lineNumbers.Add(j)
				in = true
			} else {
				in = false
				// only one selection area can be, early return
				if len(lineNumbers) > 0 {
					return lineNumbers.GetKeys()
				}
			}
		}
		if in && LessThan(0, j, endx, endy) {
			lineNumbers.Add(j)
		}
	}
	return lineNumbers.GetKeys()
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

