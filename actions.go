package main

import (
	"github.com/atotto/clipboard"
	"strings"
)


func (e *Editor) onDown() {
	if len(content) == 0 { return }
	if r+1 >= len(content) { y = r - e.ROWS + 1; if y < 0 { y = 0 }; return }
	r++
	if c > len(content[r]) { c = len(content[r]) } // fit to content
	if r < y { y = r }
	if r >= y + e.ROWS { y = r - e.ROWS + 1  }
}

func (e *Editor) onUp() {
	if len(content) == 0 { return }
	if r == 0 { y = 0; return }
	r--
	if c > len(content[r]) { c = len(content[r]) } // fit to content
	if r < y { y = r }
	if r > y + e.ROWS { y = r - e.ROWS + 1  }
}

func (e *Editor) onLeft() {
	if len(content) == 0 { return }

	if c > 0 {
		c--
	} else if r > 0 {
		r -= 1
		c = len(content[r]) // fit to content
		if r < y { y = r }
	}
}
func (e *Editor) onRight() {
	if len(content) == 0 { return }

	if c < len(content[r]) {
		c++
	} else if r < len(content)-1 {
		r += 1 // to newline
		c = 0
		if r > y + e.ROWS { y ++  }
	}
}
func (e *Editor) onScrollUp() {
	if len(content) == 0 { return }
	if y == 0 { return }
	y--
}
func (e *Editor) onScrollDown() {
	if len(content) == 0 { return }
	if y + e.ROWS >= len(content) { return }
	y++
}

func (e *Editor) focus() {
	if r > y+e.ROWS { y = r + e.ROWS }
	if r < y { y = r }
}

func (e *Editor) onEnter() {

	var ops = EditOperation{{Enter, '\n', r, c}}
	tabs := countTabs(content[r], c)
	spaces := countSpaces(content[r], c)

	after := content[r][c:]
	before := content[r][:c]
	content[r] = before
	e.updateColorsAtLine(r)
	r++
	c = 0

	countToInsert := tabs
	characterToInsert := '\t'
	if tabs == 0 && spaces != 0 { characterToInsert = ' '; countToInsert = spaces }

	begining := []rune{}
	for i := 0; i < countToInsert; i++ {
		begining = append(begining, characterToInsert)
		ops = append(ops, Operation{Insert, characterToInsert, r, c+i})
	}
	c = countToInsert

	newline := append(begining, after...)
	content = insert(content, r, newline)

	if e.isColorize && e.lang != "" {
		colors = insert(colors, r, []int{})
		e.updateColorsAtLine(r)
	}

	e.undo = append(e.undo, ops)
	e.focus(); if r - y == e.ROWS { e.onScrollDown() }
	if len(e.redoStack) > 0 { e.redoStack = []EditOperation{} }
	e.update = true
	e.isContentChanged = true
	if len(content) <= 10000 { go e.writeFile() }
}

func (e *Editor) onDelete() {

	if len(e.selection.getSelectionString(content)) > 0 { e.cut(); return }

	if c > 0 {
		c--
		e.deleteCharacter(r,c)
		e.updateColorsAtLine(r)
	} else if r > 0 { // delete line
		e.undo = append(e.undo, EditOperation{{DeleteLine, ' ', r-1, len(content[r-1])}})
		l := content[r][c:]
		content = remove(content, r)
		if e.isColorize && e.lang != "" {
			if r < len(colors) { colors = remove(colors, r) }
			e.updateColorsAtLine(r)
		}

		r--
		c = len(content[r])
		content[r] = append(content[r], l...)
		e.updateColorsAtLine(r)
	}

	e.focus()
	if len(e.redoStack) > 0 { e.redoStack = []EditOperation{} }
	e.update = true
	e.isContentChanged = true
	if len(content) <= 10000 { go e.writeFile() }
}

func (e *Editor) onTab() {
	e.focus()

	selectedLines := e.selection.getSelectedLines(content)

	if len(selectedLines) == 0 {
		ch := '\t'
		e.insertCharacter(r, c, ch)
		e.updateColorsAtLine(r)
		c++
	} else  {
		var ops = EditOperation{}
		e.selection.ssx = 0
		for _, linenumber := range selectedLines {
			r = linenumber
			content[r] = insert(content[r], 0, '\t')
			e.updateColorsAtLine(r)
			ops = append(ops, Operation{Insert, '\t', r, 0})
			c = len(content[r])
		}
		e.selection.sex = c
		e.undo = append(e.undo, ops)
	}

	if len(e.redoStack) > 0 { e.redoStack = []EditOperation{} }
	e.update = true
	e.isContentChanged = true
	if len(content) <= 10000 { go e.writeFile() }
}

func (e *Editor) onBackTab() {
	e.focus()

	selectedLines := e.selection.getSelectedLines(content)

	// deleting tabs from beginning
	if len(selectedLines) == 0 {
		if content[r][0] == '\t'  {
			e.deleteCharacter(r,0)
			colors[r] = remove(colors[r], 0)
			c--
		}
	} else {
		e.selection.ssx = 0
		for _, linenumber := range selectedLines {
			r = linenumber
			if len(content[r]) > 0 && content[r][0] == '\t'  {
				e.deleteCharacter(r,0)
				colors[r] = remove(colors[r], 0)
				c = len(content[r])
			}
		}
	}

	if len(e.redoStack) > 0 { e.redoStack = []EditOperation{} }
	e.update = true
	e.isContentChanged = true
	if len(content) <= 10000 { go e.writeFile() }
}

func (e *Editor) addChar(ch rune) {
	if len(e.selection.getSelectionString(content)) != 0 { e.cut() }

	e.focus()
	e.insertCharacter(r, c, ch)
	c++

	e.maybeAddPair(ch)

	if len(e.redoStack) > 0 { e.redoStack = []EditOperation{} }

	e.update = true
	e.isContentChanged = true
	if len(content) <= 10000 { go e.writeFile() }
	e.updateColorsAtLine(r)
}

func (e *Editor) insertCharacter(line, pos int, ch rune) {
	content[line] = insert(content[line], pos, ch)
	//if lsp.isReady { go lsp.didChange(absoluteFilePath, line, pos, line, pos, string(ch)) }
	e.undo = append(e.undo, EditOperation{{Insert, ch, r, c}})
}

func (e *Editor) insertString(line, pos int, linestring string) {
	// Convert the string to insert to a slice of runes
	insertRunes := []rune(linestring)

	// Record the operation on the undo stack. Note that we're creating a new EditOperation
	// and adding all the Operations to it
	var ops = EditOperation{}
	for _, ch := range insertRunes {
		content[line] = insert(content[line], pos, ch)
		ops = append(ops, Operation{Insert, ch, line, pos})
		pos++
	}
	c = pos
	e.undo = append(e.undo, ops)
}

func (e *Editor) insertLines(line, pos int, lines []string) {
	var ops = EditOperation{}

	tabs := countTabs(content[r], c) // todo: spaces also can be
	if len(content[r]) > 0 { r++ }
	//ops = append(ops, Operation{Enter, '\n', r, c})
	for _, linestr := range lines {
		c = 0
		if r >= len(content)  { content = append(content, []rune{}) } // if last line adding empty line before

		nl := strings.Repeat("\t", tabs) + linestr
		content = insert(content, r, []rune(nl))

		ops = append(ops, Operation{Enter, '\n', r, c})
		for _, ch := range nl {
			ops = append(ops, Operation{Insert, ch, r, c})
			c++
		}
		r++
	}
	r--
	e.undo = append(e.undo, ops)
}

func (e *Editor) deleteCharacter(line, pos int) {
	e.undo = append(e.undo, EditOperation{
		{MoveCursor, content[line][pos], line, pos+1},
		{Delete, content[line][pos], line, pos},
	})
	content[line] = remove(content[line], pos)
	//if lsp.isReady { go lsp.didChange(absoluteFilePath, line,pos,line,pos+1, "")}
}



func (e *Editor) onSwapLinesUp() {
	e.focus()

	if r == 0 { return }
	var ops = EditOperation{}
	ops = append(ops, Operation{MoveCursor, ' ', r, c})

	line1 := content[r]; line2 := content[r-1]
	line1c := colors[r]; line2c := colors[r-1]

	for i := len(line1)-1; i >= 0; i-- { ops = append(ops, Operation{Delete, line1[i], r, i}) }
	for i := len(line2)-1; i >= 0; i-- { ops = append(ops, Operation{Delete, line2[i], r-1, i}) }
	for i, ch := range line1 { ops = append(ops, Operation{Insert, ch, r-1, i}) }
	for i, ch := range line2 { ops = append(ops, Operation{Insert, ch, r, i}) }

	content[r] = line2; content[r-1] = line1 // swap
	colors[r] = line2c; colors[r-1] = line1c // swap colors
	r--

	e.undo = append(e.undo, ops)
	e.selection.cleanSelection()
	e.update = true
	e.isContentChanged = true
	if len(content) <= 10000 { go e.writeFile() }
}

func (e *Editor) onSwapLinesDown() {
	e.focus()

	if r+1 == len(content) { return }

	var ops = EditOperation{}
	ops = append(ops, Operation{MoveCursor, ' ', r, c})

	line1 := content[r]; line2 := content[r+1]
	line1c := colors[r]; line2c := colors[r+1]

	for i := len(line1)-1; i >= 0; i-- { ops = append(ops, Operation{Delete, line1[i], r, i}) }
	for i := len(line2)-1; i >= 0; i-- { ops = append(ops, Operation{Delete, line2[i], r+1, i}) }
	for i, ch := range line1 { ops = append(ops, Operation{Insert, ch, r+1, i}) }
	for i, ch := range line2 { ops = append(ops, Operation{Insert, ch, r, i}) }

	content[r] = line2; content[r+1] = line1 // swap
	colors[r] = line2c; colors[r+1] = line1c // swap
	r++

	e.undo = append(e.undo, ops)
	e.selection.cleanSelection()
	e.update = true
	e.isContentChanged = true
	if len(content) <= 10000 { go e.writeFile() }
}

func (e *Editor) onCopy() {
	selectionString := e.selection.getSelectionString(content)
	clipboard.WriteAll(selectionString)
}

func (e *Editor) onSelectAll() {
	if len(content) == 0 { return }
	e.selection.ssx = 0; e.selection.ssy = 0
	e.selection.sey = len(content)
	lastElement := len(content[len(content)-1])
	e.selection.sex = lastElement
	e.selection.sey = len(content)
	e.selection.isSelected = true
}

func (e *Editor) onPaste() {
	e.focus()

	if len(e.selection.getSelectionString(content)) > 0 { e.cut() }

	text, _ := clipboard.ReadAll()
	lines := strings.Split(text, "\n")

	if len(lines) == 0 { return }

	if len(lines) == 1 { // single line paste
		e.insertString(r,c, lines[0])
	}

	if len(lines) > 1 { // multiple line paste
		e.insertLines(r,c, lines)
	}

	e.update = true
	e.updateNeeded()
}

func (e *Editor) cut() {
	e.focus()

	if len(content) <= 1 {
		content[0] = []rune{};
		r, c = 0, 0
		return
	}
	var ops = EditOperation{}

	if len(e.selection.getSelectionString(content)) == 0 { // cut single line
		ops = append(ops, Operation{MoveCursor, ' ', r, c})

		for i := len(content[r])-1; i >= 0; i-- {
			ops = append(ops, Operation{Delete, content[r][i], r, i})
		}

		if r == 0 {
			ops = append(ops, Operation{DeleteLine, '\n', 0, 0})
			c = 0
		} else {
			newc := 0
			if c > len(content[r-1]) { newc = len(content[r-1])} else { newc = c }
			ops = append(ops, Operation{DeleteLine, '\n', r-1, newc})
			c = newc
		}

		content = remove(content, r)
		if e.isColorize && e.lang != "" {
			colors = remove(colors, r)
			e.updateColorsAtLine(r)
		}
		if r > 0 { r-- }

		e.update = true
		e.isContentChanged = true
		if len(content) <= 10000 { go e.writeFile() }

	} else { // cut selection

		//selectionString := getSelectionString(content, ssx, ssy, sex, sey)
		//clipboard.WriteAll(selectionString)

		ops = append(ops, Operation{MoveCursor, ' ', r, c})

		selectedIndices := e.selection.getSelectedIndices(content)

		// Sort selectedIndices in reverse order to delete characters from the end
		for i := len(selectedIndices) - 1; i >= 0; i-- {
			indices := selectedIndices[i]
			xd := indices[0]
			yd := indices[1]
			c, r = xd, yd

			// Delete the character at index (x, j)
			ops = append(ops, Operation{Delete, content[yd][xd], yd, xd})
			content[yd] = append(content[yd][:xd], content[yd][xd+1:]...)
			colors[yd] = append(colors[yd][:xd], colors[yd][xd+1:]...)

			if len(content[yd]) == 0 { // delete line
				if r == 0 { ops = append(ops, Operation{DeleteLine, '\n', 0, 0}) } else {
					ops = append(ops, Operation{DeleteLine, '\n', r-1, len(content[r-1])})
				}

				content = append(content[:yd], content[yd+1:]...)
				colors = append(colors[:yd], colors[yd+1:]...)
			}
		}

		if len(content) == 0 {
			content = make([][]rune, 1)
			colors = make([][]int, 1)
		}

		if r >= len(content)  {
			r = len(content) - 1
			if c >= len(content[r]) { c = len(content[r]) - 1 }
		}
		e.selection.cleanSelection()
		e.update = true
		e.isContentChanged = true
		if len(content) <= 10000 { go e.writeFile() }

		//e.updateNeeded()
	}

	e.undo = append(e.undo, ops)
}

func (e *Editor) duplicate() {
	e.focus()

	if len(content) == 0 { return }

	if e.selection.ssx == -1 && e.selection.ssy == -1 ||
		e.selection.ssx == e.selection.sex && e.selection.ssy == e.selection.sey  {
		var ops = EditOperation{}
		ops = append(ops, Operation{MoveCursor, ' ', r, c})
		ops = append(ops, Operation{Enter, '\n', r, len(content[r])})

		duplicatedSlice := make([]rune, len(content[r]))
		copy(duplicatedSlice, content[r])
		for i, ch := range duplicatedSlice {
			ops = append(ops, Operation{Insert, ch, r, i})
		}
		r++
		content = insert(content, r, duplicatedSlice)
		if e.isColorize && e.lang != "" {
			colors = insert(colors, r, []int{})
			e.updateColorsAtLine(r)
		}
		e.undo = append(e.undo, ops)
		e.update = true
		e.isContentChanged = true
		if len(content) <= 10000 { go e.writeFile() }

	} else {
		selection := e.selection.getSelectionString(content)
		if len(selection) == 0 { return }
		lines := strings.Split(selection, "\n")

		if len(lines) == 0 { return }

		if len(lines) == 1 { // single line
			lines[0] = " " + lines[0]// add space before
			e.insertString(r,c, lines[0])
		}

		if len(lines) > 1 { // multiple line
			e.insertLines(r,c, lines)
		}
		e.selection.cleanSelection()
		e.updateNeeded()
	}

}
func (e *Editor) onUndo() {
	if len(e.undo) == 0 { return }

	lastOperation := e.undo[len(e.undo)-1]
	e.undo = e.undo[:len(e.undo)-1]
	e.focus()
	for i := len(lastOperation) - 1; i >= 0; i-- {
		o := lastOperation[i]

		if o.action == Insert {
			r = o.line; c = o.column
			content[r] = append(content[r][:c], content[r][c+1:]...)

		} else if o.action == Delete {
			r = o.line; c = o.column
			content[r] = insert(content[r], c, o.char)

		} else if o.action == Enter {
			// Merge lines
			content[o.line] = append(content[o.line], content[o.line+1]...)
			content = append(content[:o.line+1], content[o.line+2:]...)
			r = o.line; c = o.column

		} else if o.action == DeleteLine {
			// Insert enter
			r = o.line; c = o.column
			after := content[r][c:]
			before := content[r][:c]
			content[r] = before
			r++; c = 0
			newline := append([]rune{}, after...)
			content = insert(content, r, newline)
		} else if o.action == MoveCursor {
			r = o.line; c = o.column
		}
	}

	e.redoStack = append(e.redoStack, lastOperation)
	e.updateNeeded()
}
func (e *Editor) redo() {
	if len(e.redoStack) == 0 { return }

	lastRedoOperation := e.redoStack[len(e.redoStack)-1]
	e.redoStack = e.redoStack[:len(e.redoStack)-1]

	for i := 0; i < len(lastRedoOperation); i++ {
		o := lastRedoOperation[i]

		if o.action == Insert {
			r = o.line; c = o.column
			content[r] = insert(content[r], c, o.char)
			c++
		} else if o.action == Delete {
			r = o.line; c = o.column
			content[r] = append(content[r][:c], content[r][c+1:]...)
		} else if o.action == Enter {
			r = o.line; c = o.column
			after := content[r][c:]
			before := content[r][:c]
			content[r] = before
			r++; c = 0
			newline := append([]rune{}, after...)
			content = insert(content, r, newline)
		} else if o.action == DeleteLine {
			// Merge lines
			content[o.line] = append(content[o.line], content[o.line+1]...)
			content = append(content[:o.line+1], content[o.line+2:]...)
			r = o.line; c = o.column
		} else if o.action == MoveCursor {
			r = o.line; c = o.column
		}
	}

	e.undo = append(e.undo, lastRedoOperation)
	e.updateNeeded()
}
func (e *Editor) onCommentLine() {
	e.focus()

	found := false

	for i, ch := range content[r] {
		if len(content[r]) == 0 { break }
		if len(e.langConf.Comment) == 1 && ch == rune(e.langConf.Comment[0]) {
			// found 1 char comment, uncomment
			c = i
			e.undo = append(e.undo, EditOperation{
				{MoveCursor, content[r][i], r, i+1},
				{Delete, content[r][i], r, i},
			})
			content[r] = remove(content[r], i)
			e.updateColorsAtLine(r)
			found = true
			break
		}
		if len(e.langConf.Comment) == 2 && ch == rune(e.langConf.Comment[0]) && content[r][i+1] == rune(e.langConf.Comment[1]) {
			// found 2 char comment, uncomment
			c = i
			e.undo = append(e.undo, EditOperation{
				{MoveCursor, content[r][i], r, i+1},
				{Delete, content[r][i], r, i},
				{MoveCursor, content[r][i+1], r, i+1},
				{Delete, content[r][i], r, i},
			})
			content[r] = remove(content[r], i)
			content[r] = remove(content[r], i)
			e.updateColorsAtLine(r)
			found = true
			break
		}
	}

	if found {
		if c < 0 { c = 0 }
		e.onDown()
		e.update = true
		e.isContentChanged = true
		if len(content) <= 10000 { go e.writeFile() }
		return
	}

	tabs := countTabs(content[r], c)
	spaces := countSpaces(content[r], c)

	from := tabs
	if tabs == 0 && spaces != 0 { from = spaces }

	c = from
	ops := EditOperation{}
	for _, ch := range e.langConf.Comment {
		content[r] = insert(content[r], c, ch)
		ops = append(ops, Operation{Insert, ch, r, c})
	}

	e.updateColorsAtLine(r)

	e.undo = append(e.undo, ops)
	if c < 0 { c = 0 }
	e.onDown()
	e.update = true
	e.isContentChanged = true
	if len(content) <= 10000 { go e.writeFile() }
}

func (e *Editor) handleSmartMove(char rune) {
	e.focus()
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

func (e *Editor) handleSmartMoveDown() {

	var ops = EditOperation{{Enter, '\n', r, c}}

	// moving down, insert new line, add same amount of tabs
	tabs := countTabs(content[r], c)
	spaces := countSpaces(content[r], c)

	countToInsert := tabs
	characterToInsert := '\t'
	if tabs == 0 && spaces != 0 { characterToInsert = ' '; countToInsert = spaces }

	r++; c = 0
	content = insert(content, r, []rune{})
	for i := 0; i < countToInsert; i++ {
		content[r] = insert(content[r], c, characterToInsert)
		ops = append(ops, Operation{Insert, characterToInsert, r, c })
		c++
	}

	if e.isColorize && e.lang != "" {
		colors = insert(colors, r, []int{})
		e.updateColorsAtLine(r)
	}

	e.focus(); e.onScrollDown()
	e.undo = append(e.undo, ops)
	e.update = true
	e.isContentChanged = true
	if len(content) <= 10000 { go e.writeFile() }
}

func (e *Editor) handleSmartMoveUp() {
	e.focus()
	// add new line and shift all lines, add same amount of tabs/spaces
	tabs := countTabs(content[r], c)
	spaces := countSpaces(content[r], c)

	countToInsert := tabs
	characterToInsert := '\t'
	if tabs == 0 && spaces != 0 { characterToInsert = ' '; countToInsert = spaces }

	var ops = EditOperation{{Enter, '\n', r, c}}
	content = insert(content, r, []rune{})

	c = 0
	for i := 0; i < countToInsert; i++ {
		content[r] = insert(content[r], c, characterToInsert)
		ops = append(ops, Operation{Insert, characterToInsert, r, c })
		c++
	}

	if e.isColorize && e.lang != "" {
		colors = insert(colors, r, []int{})
		e.updateColorsAtLine(r)
	}

	e.undo = append(e.undo, ops)
	e.update = true
	e.isContentChanged = true
	if len(content) <= 10000 { go e.writeFile() }
}

func (e *Editor) maybeAddPair(ch rune) {
	pairMap := map[rune]rune{
		'(': ')', '{': '}', '[': ']',
		'"': '"', '\'': '\'', '`': '`',
	}

	if closeChar, found := pairMap[ch]; found {
		noMoreChars := c >= len(content[r])
		isSpaceNext := c < len(content[r]) && content[r][c] == ' '
		isStringAndClosedBracketNext := closeChar == '"' && c < len(content[r]) && content[r][c] == ')'

		if noMoreChars || isSpaceNext || isStringAndClosedBracketNext {
			e.insertCharacter(r, c, closeChar)
		}
	}
}
