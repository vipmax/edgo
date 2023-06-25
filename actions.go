package main

import (
	"github.com/atotto/clipboard"
	"strings"
)


func (e *Editor) onDown() {
	if len(e.content) == 0 { return }
	if e.r+1 >= len(e.content) { e.y = e.r - e.ROWS + 1; if e.y < 0 { e.y = 0 }; return }
	e.r++
	if e.c > len(e.content[e.r]) { e.c = len(e.content[e.r]) } // fit to e.content
	if e.r < e.y { e.y = e.r }
	if e.r >= e.y + e.ROWS { e.y = e.r - e.ROWS + 1  }
}

func (e *Editor) onUp() {
	if len(e.content) == 0 { return }
	if e.r == 0 { e.y = 0; return }
	e.r--
	if e.c > len(e.content[e.r]) { e.c = len(e.content[e.r]) } // fit to e.content
	if e.r < e.y { e.y = e.r }
	if e.r > e.y + e.ROWS { e.y = e.r - e.ROWS + 1  }
}

func (e *Editor) onLeft() {
	if len(e.content) == 0 { return }

	if e.c > 0 {
		e.c--
	} else if e.r > 0 {
		e.r--
		e.c = len(e.content[e.r]) // fit to e.content
		if e.r < e.y { e.y = e.r }
	}
}
func (e *Editor) onRight() {
	if len(e.content) == 0 { return }

	if e.c < len(e.content[e.r]) {
		e.c++
	} else if e.r < len(e.content)-1 {
		e.r++
		e.c = 0
		if e.r > e.y + e.ROWS { e.y ++  }
	}
}
func (e *Editor) onScrollUp() {
	if len(e.content) == 0 { return }
	if e.y == 0 { return }
	e.y--
}
func (e *Editor) onScrollDown() {
	if len(e.content) == 0 { return }
	if e.y + e.ROWS >= len(e.content) { return }
	e.y++
}

func (e *Editor) focus() {
	if e.r > e.y + e.ROWS { e.y = e.r + e.ROWS }
	if e.r < e.y { e.y = e.r }
}

func (e *Editor) onEnter() {

	var ops = EditOperation{{Enter, '\n', e.r, e.c}}
	tabs := countTabs(e.content[e.r], e.c)
	spaces := countSpaces(e.content[e.r], e.c)

	after := e.content[e.r][e.c:]
	before := e.content[e.r][:e.c]
	e.content[e.r] = before
	e.updateColorsAtLine(e.r)
	e.r++
	e.c = 0

	countToInsert := tabs
	characterToInsert := '\t'
	if tabs == 0 && spaces != 0 { characterToInsert = ' '; countToInsert = spaces }

	begining := []rune{}
	for i := 0; i < countToInsert; i++ {
		begining = append(begining, characterToInsert)
		ops = append(ops, Operation{Insert, characterToInsert, e.r, e.c + i})
	}
	e.c = countToInsert

	newline := append(begining, after...)
	e.content = insert(e.content, e.r, newline)

	if e.isColorize && e.lang != "" {
		e.colors = insert(e.colors, e.r, []int{})
		e.updateColorsAtLine(e.r)
	}

	e.undo = append(e.undo, ops)
	e.focus(); if e.r - e.y == e.ROWS { e.onScrollDown() }
	if len(e.redo) > 0 { e.redo = []EditOperation{} }
	e.update = true
	e.isContentChanged = true
	if len(e.content) <= 10000 { go e.writeFile() }
}

func (e *Editor) onDelete() {

	if len(e.selection.getSelectionString(e.content)) > 0 { e.cut(); return }

	if e.c > 0 {
		e.c--
		e.deleteCharacter(e.r, e.c)
		e.updateColorsAtLine(e.r)
	} else if e.r > 0 { // delete line
		e.undo = append(e.undo, EditOperation{{DeleteLine, ' ', e.r-1, len(e.content[e.r-1])}})
		l := e.content[e.r][e.c:]
		e.content = remove(e.content, e.r)
		if e.isColorize && e.lang != "" {
			if e.r < len(e.colors) { e.colors = remove(e.colors, e.r) }
			e.updateColorsAtLine(e.r)
		}

		e.r--
		e.c = len(e.content[e.r])
		e.content[e.r] = append(e.content[e.r], l...)
		e.updateColorsAtLine(e.r)
	}

	e.focus()
	if len(e.redo) > 0 { e.redo = []EditOperation{} }
	e.update = true
	e.isContentChanged = true
	if len(e.content) <= 10000 { go e.writeFile() }
}

func (e *Editor) onTab() {
	e.focus()

	selectedLines := e.selection.getSelectedLines(e.content)

	if len(selectedLines) == 0 {
		ch := '\t'
		e.insertCharacter(e.r, e.c, ch)
		e.updateColorsAtLine(e.r)
		e.c++
	} else  {
		var ops = EditOperation{}
		e.selection.ssx = 0
		for _, linenumber := range selectedLines {
			e.r = linenumber
			e.content[e.r] = insert(e.content[e.r], 0, '\t')
			e.updateColorsAtLine(e.r)
			ops = append(ops, Operation{Insert, '\t', e.r, 0})
			e.c = len(e.content[e.r])
		}
		e.selection.sex = e.c
		e.undo = append(e.undo, ops)
	}

	if len(e.redo) > 0 { e.redo = []EditOperation{} }
	e.update = true
	e.isContentChanged = true
	if len(e.content) <= 10000 { go e.writeFile() }
}

func (e *Editor) onBackTab() {
	e.focus()

	selectedLines := e.selection.getSelectedLines(e.content)

	// deleting tabs from beginning
	if len(selectedLines) == 0 {
		if e.content[e.r][0] == '\t'  {
			e.deleteCharacter(e.r,0)
			e.colors[e.r] = remove(e.colors[e.r], 0)
			e.c--
		}
	} else {
		e.selection.ssx = 0
		for _, linenumber := range selectedLines {
			e.r = linenumber
			if len(e.content[e.r]) > 0 && e.content[e.r][0] == '\t'  {
				e.deleteCharacter(e.r,0)
				e.colors[e.r] = remove(e.colors[e.r], 0)
				e.c = len(e.content[e.r])
			}
		}
	}

	if len(e.redo) > 0 { e.redo = []EditOperation{} }
	e.update = true
	e.isContentChanged = true
	if len(e.content) <= 10000 { go e.writeFile() }
}

func (e *Editor) addChar(ch rune) {
	if len(e.selection.getSelectionString(e.content)) != 0 { e.cut() }

	e.focus()
	e.insertCharacter(e.r, e.c, ch)
	e.c++

	e.maybeAddPair(ch)

	if len(e.redo) > 0 { e.redo = []EditOperation{} }

	e.update = true
	e.isContentChanged = true
	if len(e.content) <= 10000 { go e.writeFile() }
	e.updateColorsAtLine(e.r)
}

func (e *Editor) insertCharacter(line, pos int, ch rune) {
	e.content[line] = insert(e.content[line], pos, ch)
	//if lsp.isReady { go lsp.didChange(absoluteFilePath, line, pos, line, pos, string(ch)) }
	e.undo = append(e.undo, EditOperation{{Insert, ch, e.r, e.c}})
}

func (e *Editor) insertString(line, pos int, linestring string) {
	// Convert the string to insert to a slice of runes
	insertRunes := []rune(linestring)

	// Record the operation on the undo stack. Note that we're creating a new EditOperation
	// and adding all the Operations to it
	var ops = EditOperation{}
	for _, ch := range insertRunes {
		e.content[line] = insert(e.content[line], pos, ch)
		ops = append(ops, Operation{Insert, ch, line, pos})
		pos++
	}
	e.c = pos
	e.undo = append(e.undo, ops)
}

func (e *Editor) insertLines(line, pos int, lines []string) {
	var ops = EditOperation{}

	tabs := countTabs(e.content[e.r], e.c) // todo: spaces also can be
	if len(e.content[e.r]) > 0 { e.r++ }
	//ops = append(ops, Operation{Enter, '\n', e.r, e.c})
	for _, linestr := range lines {
		e.c = 0
		if e.r >= len(e.content)  { e.content = append(e.content, []rune{}) } // if last line adding empty line before

		nl := strings.Repeat("\t", tabs) + linestr
		e.content = insert(e.content, e.r, []rune(nl))

		ops = append(ops, Operation{Enter, '\n', e.r, e.c})
		for _, ch := range nl {
			ops = append(ops, Operation{Insert, ch, e.r, e.c})
			e.c++
		}
		e.r++
	}
	e.r--
	e.undo = append(e.undo, ops)
}

func (e *Editor) deleteCharacter(line, pos int) {
	e.undo = append(e.undo, EditOperation{
		{MoveCursor, e.content[line][pos], line, pos+1},
		{Delete, e.content[line][pos], line, pos},
	})
	e.content[line] = remove(e.content[line], pos)
	//if lsp.isReady { go lsp.didChange(absoluteFilePath, line,pos,line,pos+1, "")}
}



func (e *Editor) onSwapLinesUp() {
	e.focus()

	if e.r == 0 { return }
	var ops = EditOperation{}
	ops = append(ops, Operation{MoveCursor, ' ', e.r, e.c})

	line1 := e.content[e.r]; line2 := e.content[e.r-1]
	line1c := e.colors[e.r]; line2c := e.colors[e.r-1]

	for i := len(line1)-1; i >= 0; i-- { ops = append(ops, Operation{Delete, line1[i], e.r, i}) }
	for i := len(line2)-1; i >= 0; i-- { ops = append(ops, Operation{Delete, line2[i], e.r-1, i}) }
	for i, ch := range line1 { ops = append(ops, Operation{Insert, ch, e.r-1, i}) }
	for i, ch := range line2 { ops = append(ops, Operation{Insert, ch, e.r, i}) }

	e.content[e.r] = line2; e.content[e.r-1] = line1 // swap
	e.colors[e.r] = line2c; e.colors[e.r-1] = line1c // swap e.colors
	e.r--

	e.undo = append(e.undo, ops)
	e.selection.cleanSelection()
	e.update = true
	e.isContentChanged = true
	if len(e.content) <= 10000 { go e.writeFile() }
}

func (e *Editor) onSwapLinesDown() {
	e.focus()

	if e.r+1 == len(e.content) { return }

	var ops = EditOperation{}
	ops = append(ops, Operation{MoveCursor, ' ', e.r, e.c})

	line1 := e.content[e.r]; line2 := e.content[e.r+1]
	line1c := e.colors[e.r]; line2c := e.colors[e.r+1]

	for i := len(line1)-1; i >= 0; i-- { ops = append(ops, Operation{Delete, line1[i], e.r, i}) }
	for i := len(line2)-1; i >= 0; i-- { ops = append(ops, Operation{Delete, line2[i], e.r+1, i}) }
	for i, ch := range line1 { ops = append(ops, Operation{Insert, ch, e.r+1, i}) }
	for i, ch := range line2 { ops = append(ops, Operation{Insert, ch, e.r, i}) }

	e.content[e.r] = line2; e.content[e.r+1] = line1 // swap
	e.colors[e.r] = line2c; e.colors[e.r+1] = line1c // swap
	e.r++

	e.undo = append(e.undo, ops)
	e.selection.cleanSelection()
	e.update = true
	e.isContentChanged = true
	if len(e.content) <= 10000 { go e.writeFile() }
}

func (e *Editor) onCopy() {
	selectionString := e.selection.getSelectionString(e.content)
	clipboard.WriteAll(selectionString)
}

func (e *Editor) onSelectAll() {
	if len(e.content) == 0 { return }
	e.selection.ssx = 0; e.selection.ssy = 0
	e.selection.sey = len(e.content)
	lastElement := len(e.content[len(e.content)-1])
	e.selection.sex = lastElement
	e.selection.sey = len(e.content)
	e.selection.isSelected = true
}

func (e *Editor) onPaste() {
	e.focus()

	if len(e.selection.getSelectionString(e.content)) > 0 { e.cut() }

	text, _ := clipboard.ReadAll()
	lines := strings.Split(text, "\n")

	if len(lines) == 0 { return }

	if len(lines) == 1 { // single line paste
		e.insertString(e.r, e.c, lines[0])
	}

	if len(lines) > 1 { // multiple line paste
		e.insertLines(e.r, e.c, lines)
	}

	e.update = true
	e.updateNeeded()
}

func (e *Editor) cut() {
	e.focus()

	if len(e.content) <= 1 {
		e.content[0] = []rune{};
		e.r, e.c = 0, 0
		return
	}
	var ops = EditOperation{}

	if len(e.selection.getSelectionString(e.content)) == 0 { // cut single line
		ops = append(ops, Operation{MoveCursor, ' ', e.r, e.c})

		for i := len(e.content[e.r])-1; i >= 0; i-- {
			ops = append(ops, Operation{Delete, e.content[e.r][i], e.r, i})
		}

		if e.r == 0 {
			ops = append(ops, Operation{DeleteLine, '\n', 0, 0})
			e.c = 0
		} else {
			newc := 0
			if e.c > len(e.content[e.r-1]) { newc = len(e.content[e.r-1])} else { newc = e.c }
			ops = append(ops, Operation{DeleteLine, '\n', e.r-1, newc})
			e.c = newc
		}

		e.content = remove(e.content, e.r)
		if e.isColorize && e.lang != "" {
			e.colors = remove(e.colors, e.r)
			e.updateColorsAtLine(e.r)
		}
		if e.r > 0 { e.r-- }

		e.update = true
		e.isContentChanged = true
		if len(e.content) <= 10000 { go e.writeFile() }

	} else { // cut selection

		//selectionString := getSelectionString(e.content, ssx, ssy, sex, sey)
		//clipboard.WriteAll(selectionString)

		ops = append(ops, Operation{MoveCursor, ' ', e.r, e.c})

		selectedIndices := e.selection.getSelectedIndices(e.content)

		// Sort selectedIndices in reverse order to delete characters from the end
		for i := len(selectedIndices) - 1; i >= 0; i-- {
			indices := selectedIndices[i]
			xd := indices[0]
			yd := indices[1]
			e.c, e.r = xd, yd

			// Delete the character at index (x, j)
			ops = append(ops, Operation{Delete, e.content[yd][xd], yd, xd})
			e.content[yd] = append(e.content[yd][:xd], e.content[yd][xd+1:]...)
			e.colors[yd] = append(e.colors[yd][:xd], e.colors[yd][xd+1:]...)

			if len(e.content[yd]) == 0 { // delete line
				if e.r == 0 { ops = append(ops, Operation{DeleteLine, '\n', 0, 0}) } else {
					ops = append(ops, Operation{DeleteLine, '\n', e.r-1, len(e.content[e.r-1])})
				}

				e.content = append(e.content[:yd], e.content[yd+1:]...)
				e.colors = append(e.colors[:yd], e.colors[yd+1:]...)
			}
		}

		if len(e.content) == 0 {
			e.content = make([][]rune, 1)
			e.colors = make([][]int, 1)
		}

		if e.r >= len(e.content)  {
			e.r = len(e.content) - 1
			if e.c >= len(e.content[e.r]) { e.c = len(e.content[e.r]) - 1 }
		}
		e.selection.cleanSelection()
		e.update = true
		e.isContentChanged = true
		if len(e.content) <= 10000 { go e.writeFile() }

		//e.updateNeeded()
	}

	e.undo = append(e.undo, ops)
}

func (e *Editor) duplicate() {
	e.focus()

	if len(e.content) == 0 { return }

	if e.selection.ssx == -1 && e.selection.ssy == -1 ||
		e.selection.ssx == e.selection.sex && e.selection.ssy == e.selection.sey  {
		var ops = EditOperation{}
		ops = append(ops, Operation{MoveCursor, ' ', e.r, e.c})
		ops = append(ops, Operation{Enter, '\n', e.r, len(e.content[e.r])})

		duplicatedSlice := make([]rune, len(e.content[e.r]))
		copy(duplicatedSlice, e.content[e.r])
		for i, ch := range duplicatedSlice {
			ops = append(ops, Operation{Insert, ch, e.r, i})
		}
		e.r++
		e.content = insert(e.content, e.r, duplicatedSlice)
		if e.isColorize && e.lang != "" {
			e.colors = insert(e.colors, e.r, []int{})
			e.updateColorsAtLine(e.r)
		}
		e.undo = append(e.undo, ops)
		e.update = true
		e.isContentChanged = true
		if len(e.content) <= 10000 { go e.writeFile() }

	} else {
		selection := e.selection.getSelectionString(e.content)
		if len(selection) == 0 { return }
		lines := strings.Split(selection, "\n")

		if len(lines) == 0 { return }

		if len(lines) == 1 { // single line
			lines[0] = " " + lines[0]// add space before
			e.insertString(e.r, e.c, lines[0])
		}

		if len(lines) > 1 { // multiple line
			e.insertLines(e.r, e.c, lines)
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
			e.r = o.line; e.c = o.column
			e.content[e.r] = append(e.content[e.r][:e.c], e.content[e.r][e.c+1:]...)

		} else if o.action == Delete {
			e.r = o.line; e.c = o.column
			e.content[e.r] = insert(e.content[e.r], e.c, o.char)

		} else if o.action == Enter {
			// Merge lines
			e.content[o.line] = append(e.content[o.line], e.content[o.line+1]...)
			e.content = append(e.content[:o.line+1], e.content[o.line+2:]...)
			e.r = o.line; e.c = o.column

		} else if o.action == DeleteLine {
			// Insert enter
			e.r = o.line; e.c = o.column
			after := e.content[e.r][e.c:]
			before := e.content[e.r][:e.c]
			e.content[e.r] = before
			e.r++; e.c = 0
			newline := append([]rune{}, after...)
			e.content = insert(e.content, e.r, newline)
		} else if o.action == MoveCursor {
			e.r = o.line; e.c = o.column
		}
	}

	e.redo = append(e.redo, lastOperation)
	e.updateNeeded()
}
func (e *Editor) onRedo() {
	if len(e.redo) == 0 { return }

	lastRedoOperation := e.redo[len(e.redo)-1]
	e.redo = e.redo[:len(e.redo)-1]

	for i := 0; i < len(lastRedoOperation); i++ {
		o := lastRedoOperation[i]

		if o.action == Insert {
			e.r = o.line; e.c = o.column
			e.content[e.r] = insert(e.content[e.r], e.c, o.char)
			e.c++
		} else if o.action == Delete {
			e.r = o.line; e.c = o.column
			e.content[e.r] = append(e.content[e.r][:e.c], e.content[e.r][e.c+1:]...)
		} else if o.action == Enter {
			e.r = o.line; e.c = o.column
			after := e.content[e.r][e.c:]
			before := e.content[e.r][:e.c]
			e.content[e.r] = before
			e.r++; e.c = 0
			newline := append([]rune{}, after...)
			e.content = insert(e.content, e.r, newline)
		} else if o.action == DeleteLine {
			// Merge lines
			e.content[o.line] = append(e.content[o.line], e.content[o.line+1]...)
			e.content = append(e.content[:o.line+1], e.content[o.line+2:]...)
			e.r = o.line; e.c = o.column
		} else if o.action == MoveCursor {
			e.r = o.line; e.c = o.column
		}
	}

	e.undo = append(e.undo, lastRedoOperation)
	e.updateNeeded()
}
func (e *Editor) onCommentLine() {
	e.focus()

	found := false

	for i, ch := range e.content[e.r] {
		if len(e.content[e.r]) == 0 { break }
		if len(e.langConf.Comment) == 1 && ch == rune(e.langConf.Comment[0]) {
			// found 1 char comment, uncomment
			e.c = i
			e.undo = append(e.undo, EditOperation{
				{MoveCursor, e.content[e.r][i], e.r, i+1},
				{Delete, e.content[e.r][i], e.r, i},
			})
			e.content[e.r] = remove(e.content[e.r], i)
			e.updateColorsAtLine(e.r)
			found = true
			break
		}
		if len(e.langConf.Comment) == 2 && ch == rune(e.langConf.Comment[0]) && e.content[e.r][i+1] == rune(e.langConf.Comment[1]) {
			// found 2 char comment, uncomment
			e.c = i
			e.undo = append(e.undo, EditOperation{
				{MoveCursor, e.content[e.r][i], e.r, i+1},
				{Delete, e.content[e.r][i], e.r, i},
				{MoveCursor, e.content[e.r][i+1], e.r, i+1},
				{Delete, e.content[e.r][i], e.r, i},
			})
			e.content[e.r] = remove(e.content[e.r], i)
			e.content[e.r] = remove(e.content[e.r], i)
			e.updateColorsAtLine(e.r)
			found = true
			break
		}
	}

	if found {
		if e.c < 0 { e.c = 0 }
		e.onDown()
		e.update = true
		e.isContentChanged = true
		if len(e.content) <= 10000 { go e.writeFile() }
		return
	}

	tabs := countTabs(e.content[e.r], e.c)
	spaces := countSpaces(e.content[e.r], e.c)

	from := tabs
	if tabs == 0 && spaces != 0 { from = spaces }

	e.c = from
	ops := EditOperation{}
	for _, ch := range e.langConf.Comment {
		e.content[e.r] = insert(e.content[e.r], e.c, ch)
		ops = append(ops, Operation{Insert, ch, e.r, e.c})
	}

	e.updateColorsAtLine(e.r)

	e.undo = append(e.undo, ops)
	if e.c < 0 { e.c = 0 }
	e.onDown()
	e.update = true
	e.isContentChanged = true
	if len(e.content) <= 10000 { go e.writeFile() }
}

func (e *Editor) handleSmartMove(char rune) {
	e.focus()
	if char == 'f' || char == 'F' {
		nw := findNextWord(e.content[e.r], e.c + 1)
		e.c = nw
		e.c = min(e.c, len(e.content[e.r]))
	}
	if char == 'b' || char == 'B' {
		nw := findPrevWord(e.content[e.r], e.c-1)
		e.c = nw
	}
}

func (e *Editor) handleSmartMoveDown() {

	var ops = EditOperation{{Enter, '\n', e.r, e.c}}

	// moving down, insert new line, add same amount of tabs
	tabs := countTabs(e.content[e.r], e.c)
	spaces := countSpaces(e.content[e.r], e.c)

	countToInsert := tabs
	characterToInsert := '\t'
	if tabs == 0 && spaces != 0 { characterToInsert = ' '; countToInsert = spaces }

	e.r++; e.c = 0
	e.content = insert(e.content, e.r, []rune{})
	for i := 0; i < countToInsert; i++ {
		e.content[e.r] = insert(e.content[e.r], e.c, characterToInsert)
		ops = append(ops, Operation{Insert, characterToInsert, e.r, e.c })
		e.c++
	}

	if e.isColorize && e.lang != "" {
		e.colors = insert(e.colors, e.r, []int{})
		e.updateColorsAtLine(e.r)
	}

	e.focus(); e.onScrollDown()
	e.undo = append(e.undo, ops)
	e.update = true
	e.isContentChanged = true
	if len(e.content) <= 10000 { go e.writeFile() }
}

func (e *Editor) handleSmartMoveUp() {
	e.focus()
	// add new line and shift all lines, add same amount of tabs/spaces
	tabs := countTabs(e.content[e.r], e.c)
	spaces := countSpaces(e.content[e.r], e.c)

	countToInsert := tabs
	characterToInsert := '\t'
	if tabs == 0 && spaces != 0 { characterToInsert = ' '; countToInsert = spaces }

	var ops = EditOperation{{Enter, '\n', e.r, e.c}}
	e.content = insert(e.content, e.r, []rune{})

	e.c = 0
	for i := 0; i < countToInsert; i++ {
		e.content[e.r] = insert(e.content[e.r], e.c, characterToInsert)
		ops = append(ops, Operation{Insert, characterToInsert, e.r, e.c })
		e.c++
	}

	if e.isColorize && e.lang != "" {
		e.colors = insert(e.colors, e.r, []int{})
		e.updateColorsAtLine(e.r)
	}

	e.undo = append(e.undo, ops)
	e.update = true
	e.isContentChanged = true
	if len(e.content) <= 10000 { go e.writeFile() }
}

func (e *Editor) maybeAddPair(ch rune) {
	pairMap := map[rune]rune{
		'(': ')', '{': '}', '[': ']',
		'"': '"', '\'': '\'', '`': '`',
	}

	if closeChar, found := pairMap[ch]; found {
		noMoreChars := e.c >= len(e.content[e.r])
		isSpaceNext := e.c < len(e.content[e.r]) && e.content[e.r][e.c] == ' '
		isStringAndClosedBracketNext := closeChar == '"' && e.c < len(e.content[e.r]) && e.content[e.r][e.c] == ')'

		if noMoreChars || isSpaceNext || isStringAndClosedBracketNext {
			e.insertCharacter(e.r, e.c, closeChar)
		}
	}
}
