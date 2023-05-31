package main

import (
	"bufio"
	"fmt"
	"github.com/atotto/clipboard"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/gdamore/tcell"
	"github.com/gdamore/tcell/encoding"
)

var content [][]rune     // characters
var COLUMNS, ROWS = 0, 0 // term size
var r, c = 0, 0          // cursor position, row and column
var y, x = 0, 0          // offset for scrolling for row and column
var LS = 5               // left shift for line number
var ssx, ssy = -1, -1    // left shift for line number
var filename = "main.go" // file name to show
var colors [][]int       // characters colors
var highlighter = Highlighter{}

type Editor struct {
}

func (e *Editor) start() {
	if len(os.Args) > 1 {
		filename = os.Args[1]
	}
	code := e.readFile(filename)
	colors = highlighter.colorize(code, filename)

	s := e.initScreen()
	s.EnableMouse()

	for {
		e.drawEverything(s)
		e.handleEvents(s)
	}

	time.Sleep(time.Second * 200000)
}

func (e *Editor) initScreen() tcell.Screen {
	encoding.Register()
	s, err := tcell.NewScreen()
	s.Init()
	if s.HasMouse() {
		s.EnableMouse()
	}

	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}
	s.Clear()

	COLUMNS, ROWS = s.Size()
	ROWS -= 2

	return s
}

func (e *Editor) handleEvents(s tcell.Screen) {
	ev := s.PollEvent()      // Poll event
	switch ev := ev.(type) { // Process event
	case *tcell.EventMouse:
		mx, my := ev.Position()
		buttons := ev.Buttons()
		mx -= LS

		if ev.Buttons()&tcell.WheelDown != 0 {
			e.onDown()
			return
		} else if ev.Buttons()&tcell.WheelUp != 0 {
			e.onUp()
			return
		}

		//fmt.Printf("Left button: %v\n", buttons&tcell.Button1 != 0)

		if buttons&tcell.Button1 == 1 {
			r = my + y
			c = mx + x
			if r > len(content)-1 {
				r = len(content) - 1
			}
			if c > len(content[r]) {
				c = len(content[r])
			}
			if c < 0 {
				c = 0
			}
			if ssx < 0 {
				ssx, ssy = c, r
			}
		}
		if buttons&tcell.Button1 == 0 {
			ssx, ssy = -1, -1
		}

	case *tcell.EventResize:
		COLUMNS, ROWS = s.Size()
		ROWS -= 2
		s.Sync()
		s.Clear()

	case *tcell.EventKey:
		key := ev.Key()
		if key == tcell.KeyCtrlC {
			clipboard.WriteAll(getSelection())
		}
		if key == tcell.KeyCtrlV {
			e.paste()
		}
		if key == tcell.KeyCtrlX {
			e.cut()
			s.Clear()
		}
		if key == tcell.KeyCtrlD {
			e.duplicate()
		}

		if ev.Modifiers()&tcell.ModShift != 0 {
			if ssx < 0 {
				ssx, ssy = c, r
			}
			if key == tcell.KeyRight {
				e.onRight()
			}
			if key == tcell.KeyLeft {
				e.onLeft()
			}
			if key == tcell.KeyUp {
				e.onUp()
			}
			if key == tcell.KeyDown {
				e.onDown()
			}
			return
		}

		if key == tcell.KeyRune && ev.Modifiers()&tcell.ModAlt != 0 {
			if len(content) == 0 {
				return
			}
			e.handleSmartMove(ev.Rune())
			return
		}
		if key == tcell.KeyRune {
			e.addChar(ev.Rune())
		}

		ssx, ssy = -1, -1

		if key == tcell.KeyEscape || key == tcell.KeyCtrlQ {
			s.Fini()
			os.Exit(1)
		}
		if key == tcell.KeyCtrlS {
			e.writeFile(filename)
		}
		if key == tcell.KeyEnter {
			e.onEnter(true)
			s.Clear()
		}
		if key == tcell.KeyBackspace || key == tcell.KeyBackspace2 {
			e.onDelete()
			s.Clear()
		}
		if key == tcell.KeyDown {
			e.onDown()
		}
		if key == tcell.KeyUp {
			e.onUp()
		}
		if key == tcell.KeyLeft {
			e.onLeft()
		}
		if key == tcell.KeyRight {
			e.onRight()
		}
		if key == tcell.KeyTab {
			e.handleTabPress()
		}
		if key == tcell.KeyCtrlT {
		} // TODO: tree
		if key == tcell.KeyCtrlF {
		} // TODO: find

	}
}

func (e *Editor) drawEverything(s tcell.Screen) {
	if r < y {
		y = r
	}
	if r >= y+ROWS {
		y = r - ROWS
	}
	if c < x {
		x = c
	}
	if c >= x+COLUMNS {
		x = c - COLUMNS + 1
	}

	for row := 0; row <= ROWS; row++ {
		ry := row + y
		e.cleanLineAfter(s, 0, row)
		if row >= len(content) || ry >= len(content) {
			break
		}
		e.drawLineNumber(s, ry, row)

		for col := 0; col <= COLUMNS; col++ {
			cx := col + x
			if cx >= len(content[ry]) {
				break
			}
			ch := content[ry][cx]
			style := e.getStyle(ry, cx)
			s.SetContent(col+LS, row, ch, nil, style)
		}
	}

	status := fmt.Sprintf("%d %d %s", c+1, r+1, filename)
	e.drawText(s, 0, ROWS+1, COLUMNS, ROWS+1, status)
	e.cleanLineAfter(s, len(status), ROWS+1)
	s.ShowCursor(c-x+LS, r-y)
	s.Show()
}

func (e *Editor) getStyle(ry int, cx int) tcell.Style {
	var style = tcell.StyleDefault
	color := colors[ry][cx]
	if color > 0 {
		style = tcell.StyleDefault.Foreground(tcell.Color(color))
	}
	if isUnderSelection(cx, ry) {
		style = style.Background(56)
	}
	return style
}

func (e *Editor) drawLineNumber(s tcell.Screen, brw int, row int) {
	lineNumber := strconv.Itoa(brw + 1)
	var style = tcell.StyleDefault.Foreground(tcell.ColorDimGray)

	if brw == r {
		style = tcell.StyleDefault
	}
	for index, char := range lineNumber {
		s.SetContent(index, row, char, nil, style)
	}
}

func (e *Editor) addChar(ch rune) {
	if ssx != -1 {
		e.cut()
	}

	content[r] = insert(content[r], c, ch)
	c++
	e.updateColors()
}

func (e *Editor) onDown() {
	if len(content) == 0 {
		return
	}
	if r+1 >= len(content) {
		return
	}
	r++
	if c > len(content[r]) {
		c = len(content[r])
	} // fit to content
}

func (e *Editor) onUp() {
	if len(content) == 0 {
		return
	}
	if r == 0 {
		return
	}
	r--
	if c > len(content[r]) {
		c = len(content[r])
	} // fit to content
}

func (e *Editor) onLeft() {
	if len(content) == 0 {
		return
	}

	if c != 0 {
		if c >= 2 && content[r][c-1] == ' ' && content[r][c-2] == ' ' {
			c -= 2
		} else {
			c--
		}

	} else if r > 0 {
		r -= 1
		c = len(content[r]) // fit to content
	}
}
func (e *Editor) onRight() {
	if len(content) == 0 {
		return
	}

	if c < len(content[r]) {
		if len(content[r]) > c+1 && content[r][c] == ' ' && content[r][c+1] == ' ' {
			c += 2
		} else {
			c++
		}

	} else if r < len(content)-1 {
		r += 1 // to newline
		c = 0
	}
}

func (e *Editor) onDelete() {
	if c > 0 {
		if c >= 2 && content[r][c-1] == ' ' && content[r][c-2] == ' ' {
			c--
			content[r] = remove(content[r], c)
			c--
			content[r] = remove(content[r], c)
		} else {
			c--
			content[r] = remove(content[r], c)
		}

	} else if r > 0 {
		l := content[r][c:]
		content = remove(content, r)
		r--
		c = len(content[r])
		content[r] = append(content[r], l...)
	}

	e.updateColors()
}

func (e *Editor) onEnter(isSaveTabs bool) {
	after := content[r][c:]
	before := content[r][:c]
	content[r] = before
	r++
	c = 0

	begining := []rune{}
	if isSaveTabs {
		tabs := countTabsInRow(r - 1)
		for i := 0; i < tabs; i++ {
			begining = append(begining, ' ')
		}
		c = tabs
	}

	newline := append(begining, after...)
	content = insert(content, r, newline)

	e.updateColors()
}

func (e *Editor) cleanLine(c int, C int, s tcell.Screen, row int) {
	for x := c; x < C; x++ {
		s.SetContent(x, row, ' ', nil, tcell.StyleDefault)
	}
}

func (e *Editor) cleanLineAfter(s tcell.Screen, x, y int) {
	for i := x; i < COLUMNS; i++ {
		s.SetContent(i, y, ' ', nil, tcell.StyleDefault)
	}
}

func (e *Editor) writeFile(filename string) {
	// Convert content to a string
	contentStr := ""
	for _, row := range content {
		contentStr += string(row) + "\n"
	}
	contentStr = strings.ReplaceAll(contentStr, "  ", "\t")

	// Write content to a file
	err := os.WriteFile(filename, []byte(contentStr), 0644)
	if err != nil {
		fmt.Println("Error writing to file:", err)
	}
}

func (e *Editor) readFile(filename string) string {
	file, err := os.Open(filename)
	if err != nil {
		file, err := os.Create(filename)
		if err != nil {
			fmt.Printf("Failed to create file: %v\n", err)
		}

		defer file.Close()
		//
		//fmt.Fprintf(os.Stderr, "Failed to open file: %v", err)
		//os.Exit(1)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)

	lines := ""
	for scanner.Scan() {
		var line = scanner.Text()
		line = strings.ReplaceAll(line, "\t", "  ")
		lines += line + "\n"
		var lineChars = []rune{}
		for _, char := range line {
			lineChars = append(lineChars, char)
		}
		content = append(content, lineChars)
	}

	// if no content, consider it like one line for next editing
	if content == nil {
		content = make([][]rune, 1)
		colors = make([][]int, 1)
	}

	return lines
}

func (e *Editor) drawText(s tcell.Screen, x1, y1, x2, y2 int, text string) {
	row := y1
	col := x1
	var style = tcell.StyleDefault.Foreground(tcell.ColorDimGray)
	for _, r := range []rune(text) {
		s.SetContent(col, row, r, nil, style)
		col++
		if col >= x2 {
			row++
			col = x1
		}
		if row > y2 {
			break
		}
	}
}

func (e *Editor) updateColors() {
	code := convertToString()
	colors = highlighter.colorize(code, filename)
}

func (e *Editor) handleTabPress() {
	e.addChar(' ')
	e.addChar(' ')
}

func (e *Editor) paste() {
	text, _ := clipboard.ReadAll()
	text = strings.ReplaceAll(text, "\t", `  `)
	lines := strings.Split(text, "\n")

	for i, line := range lines {
		for _, char := range line {
			e.addChar(char)
		}
		if i < len(lines)-2 {
			e.onEnter(false)
		}
	}
}

func (e *Editor) cut() {
	if len(content) <= 1 {
		content[0] = []rune{}
		r, c = 0, 0
		return
	}

	if ssx == -1 && ssy == -1 {
		content = append(content[:r], content[r+1:]...)
		e.onUp()
	} else {
		var selectedIndices [][]int
		// calculate elements to remove
		for j := 0; j < len(content); j++ {
			for i := 0; i < len(content[j]); i++ {
				if isUnderSelection(i, j) {
					selectedIndices = append(selectedIndices, []int{i, j})
				}
			}
		}

		// Sort selectedIndices in reverse order to delete characters from the end
		for i := len(selectedIndices) - 1; i >= 0; i-- {
			indices := selectedIndices[i]
			x := indices[0]
			y := indices[1]
			c, r = x, y

			// Delete the character at indices (x, j)
			content[y] = append(content[y][:x], content[y][x+1:]...)
			if len(content[y]) == 0 {
				content = append(content[:y], content[y+1:]...)
			}
		}
	}

	e.updateColors()
}

func (e *Editor) duplicate() {
	if len(content) == 0 {
		return
	}

	if ssx == -1 && ssy == -1 {
		duplicatedSlice := make([]rune, len(content[r]))
		copy(duplicatedSlice, content[r])
		content = insert(content, r+1, duplicatedSlice)
		r++
	} else {
		clipboard.WriteAll(getSelection())
		e.paste()
	}
	e.updateColors()
}

func (e *Editor) handleSmartMove(char rune) {
	if char == 'f' || char == 'F' {
		nw := findNextWord(content[r], c+1)
		c = nw
		c = min(c, len(content[r]))
	}
	if char == 'b' || char == 'B' {
		nw := findPrevWord(content[r], c-1)
		c = nw
	}
}

func isUnderSelection(x, y int) bool {
	// Check if there is an active selection
	if ssx == -1 || ssy == -1 {
		return false
	}

	var startx, starty = ssx, ssy
	var endx, endy = c, r

	if GreaterThan(startx, starty, endx, endy) {
		startx, endx = endx, startx
		starty, endy = endy, starty
	}

	return GreaterEqual(x, y, startx, starty) && LessThan(x, y, endx, endy)
}

func getSelection() string {
	var ret = ""

	var in = false
	for j, row := range content {
		for i, char := range row {
			if isUnderSelection(i, j) {
				ret += string(char)
				in = true
			} else {
				in = false
			}
		}
		if in {
			ret += "\n"
		}
	}
	return ret
}

func convertToString() string {
	var result strings.Builder
	for _, row := range content {
		result.WriteString(string(row) + "\n")
	}
	return result.String()
}

func countTabsInRow(i int) int {
	if i < 0 || i >= len(content) {
		return 0
	} // Invalid row index

	row := content[i]
	count := 0
	for _, char := range row {
		if char == ' ' {
			count++
		} else {
			return count
		}
	}
	return count
}
