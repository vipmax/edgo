package main

import (
	"github.com/atotto/clipboard"
	"strings"
)


func (e *Editor) onDown() {
	if len(e.Content) == 0 { return }
	if e.Row+1 >= len(e.Content) {
		e.Y = e.Row - e.ROWS + 1;
		if e.Y < 0 { e.Y = 0 };
		return 
	}
	e.Row++
	if e.Col > len(e.Content[e.Row]) { e.Col = len(e.Content[e.Row]) } // fit to e.Content
	if e.Row < e.Y { e.Y = e.Row
	}
	if e.Row >= e.Y+ e.ROWS { e.Y = e.Row - e.ROWS + 1  }
}

func (e *Editor) onUp() {
	if len(e.Content) == 0 { return }
	if e.Row == 0 { e.Y = 0; return }
	e.Row--
	if e.Col > len(e.Content[e.Row]) { e.Col = len(e.Content[e.Row]) } // fit to e.Content
	if e.Row < e.Y { e.Y = e.Row
	}
	if e.Row > e.Y+ e.ROWS { e.Y = e.Row - e.ROWS + 1  }
}

func (e *Editor) onLeft() {
	if len(e.Content) == 0 { return }

	if e.Col > 0 {
		e.Col--
	} else if e.Row > 0 {
		e.Row--
		e.Col = len(e.Content[e.Row]) // fit to e.Content
		if e.Row < e.Y { e.Y = e.Row
		}
	}
}

func (e *Editor) onRight() {
	if len(e.Content) == 0 { return }

	if e.Col < len(e.Content[e.Row]) {
		e.Col++
	} else if e.Row < len(e.Content)-1 {
		e.Row++
		e.Col = 0
		if e.Row > e.Y+ e.ROWS { e.Y++  }
	}
}

func (e *Editor) onScrollUp() {
	if len(e.Content) == 0 { return }
	if e.Y == 0 { return }
	e.Y--
}

func (e *Editor) onScrollDown() {
	if len(e.Content) == 0 { return }
	if e.Y+ e.ROWS >= len(e.Content) { return }
	e.Y++
}

func (e *Editor) focus() {
	if e.Row > e.Y+ e.ROWS { e.Y = e.Row + e.ROWS }
	if e.Row < e.Y { e.Y = e.Row
	}
}

func (e *Editor) onEnter() {

	var ops = EditOperation{{Enter, '\n', e.Row, e.Col}}
	tabs := countTabs(e.Content[e.Row], e.Col)
	spaces := countSpaces(e.Content[e.Row], e.Col)

	after := e.Content[e.Row][e.Col:]
	before := e.Content[e.Row][:e.Col]
	e.Content[e.Row] = before
	e.updateColorsAtLine(e.Row)
	e.Row++
	e.Col = 0

	countToInsert := tabs
	characterToInsert := '\t'
	if tabs == 0 && spaces != 0 { characterToInsert = ' '; countToInsert = spaces }

	begining := []rune{}
	for i := 0; i < countToInsert; i++ {
		begining = append(begining, characterToInsert)
		ops = append(ops, Operation{Insert, characterToInsert, e.Row, e.Col + i})
	}
	e.Col = countToInsert

	newline := append(begining, after...)
	e.Content = insert(e.Content, e.Row, newline)

	if e.IsColorize && e.Lang != "" {
		e.Colors = insert(e.Colors, e.Row, []int{})
		e.updateColorsAtLine(e.Row)
	}

	e.Undo = append(e.Undo, ops)
	e.focus(); if e.Row- e.Y == e.ROWS { e.onScrollDown() }
	if len(e.Redo) > 0 { e.Redo = []EditOperation{} }
	e.Update = true
	e.IsContentChanged = true
	if len(e.Content) <= 10000 { go e.writeFile() }
}

func (e *Editor) onDelete() {

	if len(e.Selection.getSelectionString(e.Content)) > 0 { e.cut(); return }

	if e.Col > 0 {
		e.Col--
		e.deleteCharacter(e.Row, e.Col)
		e.updateColorsAtLine(e.Row)
	} else if e.Row > 0 { // delete line
		e.Undo = append(e.Undo, EditOperation{{DeleteLine, ' ', e.Row -1, len(e.Content[e.Row-1])}})
		l := e.Content[e.Row][e.Col:]
		e.Content = remove(e.Content, e.Row)
		if e.IsColorize && e.Lang != "" {
			if e.Row < len(e.Colors) { e.Colors = remove(e.Colors, e.Row) }
			e.updateColorsAtLine(e.Row)
		}

		e.Row--
		e.Col = len(e.Content[e.Row])
		e.Content[e.Row] = append(e.Content[e.Row], l...)
		e.updateColorsAtLine(e.Row)
	}

	e.focus()
	if len(e.Redo) > 0 { e.Redo = []EditOperation{} }
	e.Update = true
	e.IsContentChanged = true
	if len(e.Content) <= 10000 { go e.writeFile() }
}

func (e *Editor) onTab() {
	e.focus()

	selectedLines := e.Selection.getSelectedLines(e.Content)

	if len(selectedLines) == 0 {
		ch := '\t'
		e.insertCharacter(e.Row, e.Col, ch)
		e.updateColorsAtLine(e.Row)
		e.Col++
	} else  {
		var ops = EditOperation{}
		e.Selection.ssx = 0
		for _, linenumber := range selectedLines {
			e.Row = linenumber
			e.Content[e.Row] = insert(e.Content[e.Row], 0, '\t')
			e.updateColorsAtLine(e.Row)
			ops = append(ops, Operation{Insert, '\t', e.Row, 0})
			e.Col = len(e.Content[e.Row])
		}
		e.Selection.sex = e.Col
		e.Undo = append(e.Undo, ops)
	}

	if len(e.Redo) > 0 { e.Redo = []EditOperation{} }
	e.Update = true
	e.IsContentChanged = true
	if len(e.Content) <= 10000 { go e.writeFile() }
}

func (e *Editor) onBackTab() {
	e.focus()

	selectedLines := e.Selection.getSelectedLines(e.Content)

	// deleting tabs from beginning
	if len(selectedLines) == 0 {
		if e.Content[e.Row][0] == '\t'  {
			e.deleteCharacter(e.Row,0)
			e.Colors[e.Row] = remove(e.Colors[e.Row], 0)
			e.Col--
		}
	} else {
		e.Selection.ssx = 0
		for _, linenumber := range selectedLines {
			e.Row = linenumber
			if len(e.Content[e.Row]) > 0 && e.Content[e.Row][0] == '\t'  {
				e.deleteCharacter(e.Row,0)
				e.Colors[e.Row] = remove(e.Colors[e.Row], 0)
				e.Col = len(e.Content[e.Row])
			}
		}
	}

	if len(e.Redo) > 0 { e.Redo = []EditOperation{} }
	e.Update = true
	e.IsContentChanged = true
	if len(e.Content) <= 10000 { go e.writeFile() }
}

func (e *Editor) addChar(ch rune) {
	if len(e.Selection.getSelectionString(e.Content)) != 0 { e.cut() }

	e.focus()
	e.insertCharacter(e.Row, e.Col, ch)
	e.Col++

	e.maybeAddPair(ch)

	if len(e.Redo) > 0 { e.Redo = []EditOperation{} }

	e.Update = true
	e.IsContentChanged = true
	if len(e.Content) <= 10000 { go e.writeFile() }
	e.updateColorsAtLine(e.Row)
}

func (e *Editor) insertCharacter(line, pos int, ch rune) {
	e.Content[line] = insert(e.Content[line], pos, ch)
	//if lsp.isReady { go lsp.didChange(AbsoluteFilePath, line, pos, line, pos, string(ch)) }
	e.Undo = append(e.Undo, EditOperation{{Insert, ch, e.Row, e.Col}})
}

func (e *Editor) insertString(line, pos int, linestring string) {
	// Convert the string to insert to a slice of runes
	insertRunes := []rune(linestring)

	// Record the operation on the undo stack. Note that we're creating a new EditOperation
	// and adding all the Operations to it
	var ops = EditOperation{}
	for _, ch := range insertRunes {
		e.Content[line] = insert(e.Content[line], pos, ch)
		ops = append(ops, Operation{Insert, ch, line, pos})
		pos++
	}
	e.Col = pos
	e.Undo = append(e.Undo, ops)
}

func (e *Editor) insertLines(line, pos int, lines []string) {
	var ops = EditOperation{}

	tabs := countTabs(e.Content[e.Row], e.Col) // todo: spaces also can be
	if len(e.Content[e.Row]) > 0 { e.Row++ }
	//ops = append(ops, Operation{Enter, '\n', e.Row, e.Col})
	for _, linestr := range lines {
		e.Col = 0
		if e.Row >= len(e.Content)  { e.Content = append(e.Content, []rune{}) } // if last line adding empty line before

		nl := strings.Repeat("\t", tabs) + linestr
		e.Content = insert(e.Content, e.Row, []rune(nl))

		ops = append(ops, Operation{Enter, '\n', e.Row, e.Col})
		for _, ch := range nl {
			ops = append(ops, Operation{Insert, ch, e.Row, e.Col})
			e.Col++
		}
		e.Row++
	}
	e.Row--
	e.Undo = append(e.Undo, ops)
}

func (e *Editor) deleteCharacter(line, pos int) {
	e.Undo = append(e.Undo, EditOperation{
		{MoveCursor, e.Content[line][pos], line, pos+1},
		{Delete, e.Content[line][pos], line, pos},
	})
	e.Content[line] = remove(e.Content[line], pos)
	//if lsp.isReady { go lsp.didChange(AbsoluteFilePath, line,pos,line,pos+1, "")}
}



func (e *Editor) onSwapLinesUp() {
	e.focus()

	if e.Row == 0 { return }
	var ops = EditOperation{}
	ops = append(ops, Operation{MoveCursor, ' ', e.Row, e.Col})

	line1 := e.Content[e.Row]; line2 := e.Content[e.Row-1]
	line1c := e.Colors[e.Row]; line2c := e.Colors[e.Row-1]

	for i := len(line1)-1; i >= 0; i-- { ops = append(ops, Operation{Delete, line1[i], e.Row, i}) }
	for i := len(line2)-1; i >= 0; i-- { ops = append(ops, Operation{Delete, line2[i], e.Row -1, i}) }
	for i, ch := range line1 { ops = append(ops, Operation{Insert, ch, e.Row -1, i}) }
	for i, ch := range line2 { ops = append(ops, Operation{Insert, ch, e.Row, i}) }

	e.Content[e.Row] = line2; e.Content[e.Row-1] = line1 // swap
	e.Colors[e.Row] = line2c; e.Colors[e.Row-1] = line1c // swap e.Colors
	e.Row--

	e.Undo = append(e.Undo, ops)
	e.Selection.cleanSelection()
	e.Update = true
	e.IsContentChanged = true
	if len(e.Content) <= 10000 { go e.writeFile() }
}

func (e *Editor) onSwapLinesDown() {
	e.focus()

	if e.Row+1 == len(e.Content) { return }

	var ops = EditOperation{}
	ops = append(ops, Operation{MoveCursor, ' ', e.Row, e.Col})

	line1 := e.Content[e.Row]; line2 := e.Content[e.Row+1]
	line1c := e.Colors[e.Row]; line2c := e.Colors[e.Row+1]

	for i := len(line1)-1; i >= 0; i-- { ops = append(ops, Operation{Delete, line1[i], e.Row, i}) }
	for i := len(line2)-1; i >= 0; i-- { ops = append(ops, Operation{Delete, line2[i], e.Row +1, i}) }
	for i, ch := range line1 { ops = append(ops, Operation{Insert, ch, e.Row +1, i}) }
	for i, ch := range line2 { ops = append(ops, Operation{Insert, ch, e.Row, i}) }

	e.Content[e.Row] = line2; e.Content[e.Row+1] = line1 // swap
	e.Colors[e.Row] = line2c; e.Colors[e.Row+1] = line1c // swap
	e.Row++

	e.Undo = append(e.Undo, ops)
	e.Selection.cleanSelection()
	e.Update = true
	e.IsContentChanged = true
	if len(e.Content) <= 10000 { go e.writeFile() }
}

func (e *Editor) onCopy() {
	selectionString := e.Selection.getSelectionString(e.Content)
	clipboard.WriteAll(selectionString)
}

func (e *Editor) onSelectAll() {
	if len(e.Content) == 0 { return }
	e.Selection.ssx = 0; e.Selection.ssy = 0
	e.Selection.sey = len(e.Content)
	lastElement := len(e.Content[len(e.Content)-1])
	e.Selection.sex = lastElement
	e.Selection.sey = len(e.Content)
	e.Selection.isSelected = true
}

func (e *Editor) onPaste() {
	e.focus()

	if len(e.Selection.getSelectionString(e.Content)) > 0 { e.cut() }

	text, _ := clipboard.ReadAll()
	lines := strings.Split(text, "\n")

	if len(lines) == 0 { return }

	if len(lines) == 1 { // single line paste
		e.insertString(e.Row, e.Col, lines[0])
	}

	if len(lines) > 1 { // multiple line paste
		e.insertLines(e.Row, e.Col, lines)
	}

	e.Update = true
	e.updateNeeded()
}

func (e *Editor) cut() {
	e.focus()

	if len(e.Content) <= 1 {
		e.Content[0] = []rune{};
		e.Row, e.Col = 0, 0
		return
	}
	var ops = EditOperation{}

	if len(e.Selection.getSelectionString(e.Content)) == 0 { // cut single line
		ops = append(ops, Operation{MoveCursor, ' ', e.Row, e.Col})

		for i := len(e.Content[e.Row])-1; i >= 0; i-- {
			ops = append(ops, Operation{Delete, e.Content[e.Row][i], e.Row, i})
		}

		if e.Row == 0 {
			ops = append(ops, Operation{DeleteLine, '\n', 0, 0})
			e.Col = 0
		} else {
			newc := 0
			if e.Col > len(e.Content[e.Row-1]) { newc = len(e.Content[e.Row-1])} else { newc = e.Col
			}
			ops = append(ops, Operation{DeleteLine, '\n', e.Row -1, newc})
			e.Col = newc
		}

		e.Content = remove(e.Content, e.Row)
		if e.IsColorize && e.Lang != "" {
			e.Colors = remove(e.Colors, e.Row)
			e.updateColorsAtLine(e.Row)
		}
		if e.Row > 0 { e.Row-- }

		e.Update = true
		e.IsContentChanged = true
		if len(e.Content) <= 10000 { go e.writeFile() }

	} else { // cut selection

		selectionString := e.Selection.getSelectionString(e.Content)
		clipboard.WriteAll(selectionString)

		ops = append(ops, Operation{MoveCursor, ' ', e.Row, e.Col})

		selectedIndices := e.Selection.getSelectedIndices(e.Content)

		// Sort selectedIndices in reverse order to delete characters from the end
		for i := len(selectedIndices) - 1; i >= 0; i-- {
			indices := selectedIndices[i]
			xd := indices[0]
			yd := indices[1]
			e.Col, e.Row = xd, yd

			// Delete the character at index (x, j)
			ops = append(ops, Operation{Delete, e.Content[yd][xd], yd, xd})
			e.Content[yd] = append(e.Content[yd][:xd], e.Content[yd][xd+1:]...)
			e.Colors[yd] = append(e.Colors[yd][:xd], e.Colors[yd][xd+1:]...)

			if len(e.Content[yd]) == 0 { // delete line
				if e.Row == 0 { ops = append(ops, Operation{DeleteLine, '\n', 0, 0}) } else {
					ops = append(ops, Operation{DeleteLine, '\n', e.Row -1, len(e.Content[e.Row-1])})
				}

				e.Content = append(e.Content[:yd], e.Content[yd+1:]...)
				e.Colors = append(e.Colors[:yd], e.Colors[yd+1:]...)
			}
		}

		if len(e.Content) == 0 {
			e.Content = make([][]rune, 1)
			e.Colors = make([][]int, 1)
		}

		if e.Row >= len(e.Content)  {
			e.Row = len(e.Content) - 1
			if e.Col >= len(e.Content[e.Row]) { e.Col = len(e.Content[e.Row]) - 1 }
		}
		e.Selection.cleanSelection()
		e.Update = true
		e.IsContentChanged = true
		if len(e.Content) <= 10000 { go e.writeFile() }

		//e.updateNeeded()
	}

	e.Undo = append(e.Undo, ops)
}

func (e *Editor) duplicate() {
	e.focus()

	if len(e.Content) == 0 { return }

	if e.Selection.ssx == -1 && e.Selection.ssy == -1 ||
		e.Selection.ssx == e.Selection.sex && e.Selection.ssy == e.Selection.sey  {
		var ops = EditOperation{}
		ops = append(ops, Operation{MoveCursor, ' ', e.Row, e.Col})
		ops = append(ops, Operation{Enter, '\n', e.Row, len(e.Content[e.Row])})

		duplicatedSlice := make([]rune, len(e.Content[e.Row]))
		copy(duplicatedSlice, e.Content[e.Row])
		for i, ch := range duplicatedSlice {
			ops = append(ops, Operation{Insert, ch, e.Row, i})
		}
		e.Row++
		e.Content = insert(e.Content, e.Row, duplicatedSlice)
		if e.IsColorize && e.Lang != "" {
			e.Colors = insert(e.Colors, e.Row, []int{})
			e.updateColorsAtLine(e.Row)
		}
		e.Undo = append(e.Undo, ops)
		e.Update = true
		e.IsContentChanged = true
		if len(e.Content) <= 10000 { go e.writeFile() }

	} else {
		selection := e.Selection.getSelectionString(e.Content)
		if len(selection) == 0 { return }
		lines := strings.Split(selection, "\n")

		if len(lines) == 0 { return }

		if len(lines) == 1 { // single line
			lines[0] = " " + lines[0]// add space before
			e.insertString(e.Row, e.Col, lines[0])
		}

		if len(lines) > 1 { // multiple line
			e.insertLines(e.Row, e.Col, lines)
		}
		e.Selection.cleanSelection()
		e.updateNeeded()
	}

}
func (e *Editor) onUndo() {
	if len(e.Undo) == 0 { return }

	lastOperation := e.Undo[len(e.Undo)-1]
	e.Undo = e.Undo[:len(e.Undo)-1]
	e.focus()
	for i := len(lastOperation) - 1; i >= 0; i-- {
		o := lastOperation[i]

		if o.action == Insert {
			e.Row = o.line; e.Col = o.column
			e.Content[e.Row] = append(e.Content[e.Row][:e.Col], e.Content[e.Row][e.Col+1:]...)

		} else if o.action == Delete {
			e.Row = o.line; e.Col = o.column
			e.Content[e.Row] = insert(e.Content[e.Row], e.Col, o.char)

		} else if o.action == Enter {
			// Merge lines
			e.Content[o.line] = append(e.Content[o.line], e.Content[o.line+1]...)
			e.Content = append(e.Content[:o.line+1], e.Content[o.line+2:]...)
			e.Row = o.line; e.Col = o.column

		} else if o.action == DeleteLine {
			// Insert enter
			e.Row = o.line; e.Col = o.column
			after := e.Content[e.Row][e.Col:]
			before := e.Content[e.Row][:e.Col]
			e.Content[e.Row] = before
			e.Row++; e.Col = 0
			newline := append([]rune{}, after...)
			e.Content = insert(e.Content, e.Row, newline)
		} else if o.action == MoveCursor {
			e.Row = o.line; e.Col = o.column
		}
	}

	e.Redo = append(e.Redo, lastOperation)
	e.updateNeeded()
}
func (e *Editor) onRedo() {
	if len(e.Redo) == 0 { return }

	lastRedoOperation := e.Redo[len(e.Redo)-1]
	e.Redo = e.Redo[:len(e.Redo)-1]

	for i := 0; i < len(lastRedoOperation); i++ {
		o := lastRedoOperation[i]

		if o.action == Insert {
			e.Row = o.line; e.Col = o.column
			e.Content[e.Row] = insert(e.Content[e.Row], e.Col, o.char)
			e.Col++
		} else if o.action == Delete {
			e.Row = o.line; e.Col = o.column
			e.Content[e.Row] = append(e.Content[e.Row][:e.Col], e.Content[e.Row][e.Col+1:]...)
		} else if o.action == Enter {
			e.Row = o.line; e.Col = o.column
			after := e.Content[e.Row][e.Col:]
			before := e.Content[e.Row][:e.Col]
			e.Content[e.Row] = before
			e.Row++; e.Col = 0
			newline := append([]rune{}, after...)
			e.Content = insert(e.Content, e.Row, newline)
		} else if o.action == DeleteLine {
			// Merge lines
			e.Content[o.line] = append(e.Content[o.line], e.Content[o.line+1]...)
			e.Content = append(e.Content[:o.line+1], e.Content[o.line+2:]...)
			e.Row = o.line; e.Col = o.column
		} else if o.action == MoveCursor {
			e.Row = o.line; e.Col = o.column
		}
	}

	e.Undo = append(e.Undo, lastRedoOperation)
	e.updateNeeded()
}
func (e *Editor) onCommentLine() {
	e.focus()

	found := false

	for i, ch := range e.Content[e.Row] {
		if len(e.Content[e.Row]) == 0 { break }
		if len(e.langConf.Comment) == 1 && ch == rune(e.langConf.Comment[0]) {
			// found 1 char comment, uncomment
			e.Col = i
			e.Undo = append(e.Undo, EditOperation{
				{MoveCursor, e.Content[e.Row][i], e.Row, i+1},
				{Delete, e.Content[e.Row][i], e.Row, i},
			})
			e.Content[e.Row] = remove(e.Content[e.Row], i)
			e.updateColorsAtLine(e.Row)
			found = true
			break
		}
		if len(e.langConf.Comment) == 2 && ch == rune(e.langConf.Comment[0]) && e.Content[e.Row][i+1] == rune(e.langConf.Comment[1]) {
			// found 2 char comment, uncomment
			e.Col = i
			e.Undo = append(e.Undo, EditOperation{
				{MoveCursor, e.Content[e.Row][i], e.Row, i+1},
				{Delete, e.Content[e.Row][i], e.Row, i},
				{MoveCursor, e.Content[e.Row][i+1], e.Row, i+1},
				{Delete, e.Content[e.Row][i], e.Row, i},
			})
			e.Content[e.Row] = remove(e.Content[e.Row], i)
			e.Content[e.Row] = remove(e.Content[e.Row], i)
			e.updateColorsAtLine(e.Row)
			found = true
			break
		}
	}

	if found {
		if e.Col < 0 { e.Col = 0 }
		e.onDown()
		e.Update = true
		e.IsContentChanged = true
		if len(e.Content) <= 10000 { go e.writeFile() }
		return
	}

	tabs := countTabs(e.Content[e.Row], e.Col)
	spaces := countSpaces(e.Content[e.Row], e.Col)

	from := tabs
	if tabs == 0 && spaces != 0 { from = spaces }

	e.Col = from
	ops := EditOperation{}
	for _, ch := range e.langConf.Comment {
		e.Content[e.Row] = insert(e.Content[e.Row], e.Col, ch)
		ops = append(ops, Operation{Insert, ch, e.Row, e.Col})
	}

	e.updateColorsAtLine(e.Row)

	e.Undo = append(e.Undo, ops)
	if e.Col < 0 { e.Col = 0 }
	e.onDown()
	e.Update = true
	e.IsContentChanged = true
	if len(e.Content) <= 10000 { go e.writeFile() }
}

func (e *Editor) handleSmartMove(char rune) {
	e.focus()
	if char == 'f' || char == 'F' {
		nw := findNextWord(e.Content[e.Row], e.Col+ 1)
		e.Col = nw
		e.Col = min(e.Col, len(e.Content[e.Row]))
	}
	if char == 'b' || char == 'B' {
		nw := findPrevWord(e.Content[e.Row], e.Col-1)
		e.Col = nw
	}
}

func (e *Editor) handleSmartMoveDown() {

	var ops = EditOperation{{Enter, '\n', e.Row, e.Col}}

	// moving down, insert new line, add same amount of tabs
	tabs := countTabs(e.Content[e.Row], e.Col)
	spaces := countSpaces(e.Content[e.Row], e.Col)

	countToInsert := tabs
	characterToInsert := '\t'
	if tabs == 0 && spaces != 0 { characterToInsert = ' '; countToInsert = spaces }

	e.Row++; e.Col = 0
	e.Content = insert(e.Content, e.Row, []rune{})
	for i := 0; i < countToInsert; i++ {
		e.Content[e.Row] = insert(e.Content[e.Row], e.Col, characterToInsert)
		ops = append(ops, Operation{Insert, characterToInsert, e.Row, e.Col})
		e.Col++
	}

	if e.IsColorize && e.Lang != "" {
		e.Colors = insert(e.Colors, e.Row, []int{})
		e.updateColorsAtLine(e.Row)
	}

	e.focus(); e.onScrollDown()
	e.Undo = append(e.Undo, ops)
	e.Update = true
	e.IsContentChanged = true
	if len(e.Content) <= 10000 { go e.writeFile() }
}

func (e *Editor) handleSmartMoveUp() {
	e.focus()
	// add new line and shift all lines, add same amount of tabs/spaces
	tabs := countTabs(e.Content[e.Row], e.Col)
	spaces := countSpaces(e.Content[e.Row], e.Col)

	countToInsert := tabs
	characterToInsert := '\t'
	if tabs == 0 && spaces != 0 { characterToInsert = ' '; countToInsert = spaces }

	var ops = EditOperation{{Enter, '\n', e.Row, e.Col}}
	e.Content = insert(e.Content, e.Row, []rune{})

	e.Col = 0
	for i := 0; i < countToInsert; i++ {
		e.Content[e.Row] = insert(e.Content[e.Row], e.Col, characterToInsert)
		ops = append(ops, Operation{Insert, characterToInsert, e.Row, e.Col})
		e.Col++
	}

	if e.IsColorize && e.Lang != "" {
		e.Colors = insert(e.Colors, e.Row, []int{})
		e.updateColorsAtLine(e.Row)
	}

	e.Undo = append(e.Undo, ops)
	e.Update = true
	e.IsContentChanged = true
	if len(e.Content) <= 10000 { go e.writeFile() }
}

func (e *Editor) maybeAddPair(ch rune) {
	pairMap := map[rune]rune{
		'(': ')', '{': '}', '[': ']',
		'"': '"', '\'': '\'', '`': '`',
	}

	if closeChar, found := pairMap[ch]; found {
		noMoreChars := e.Col >= len(e.Content[e.Row])
		isSpaceNext := e.Col < len(e.Content[e.Row]) && e.Content[e.Row][e.Col] == ' '
		isStringAndClosedBracketNext := closeChar == '"' && e.Col < len(e.Content[e.Row]) && e.Content[e.Row][e.Col] == ')'

		if noMoreChars || isSpaceNext || isStringAndClosedBracketNext {
			e.insertCharacter(e.Row, e.Col, closeChar)
		}
	}
}