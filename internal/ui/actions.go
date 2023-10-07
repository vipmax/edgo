package ui

import (
	"edgo/internal/highlighter"
	. "edgo/internal/operations"
	. "edgo/internal/utils"
	"github.com/atotto/clipboard"
	"strings"
)


func (e *Editor) OnDown() {
	if len(e.Content) == 0 { return }
	if e.Row+1 >= len(e.Content) {
		e.Y = e.Row - e.ROWS + 1
		if e.Y < 0 { e.Y = 0 }
		return
	}
	e.Row++
	if e.Col > len(e.Content[e.Row]) { e.Col = len(e.Content[e.Row]) } // fit to e.Content
	if e.Row < e.Y { e.Y = e.Row }
	if e.Row >= e.Y+ e.ROWS { e.Y = e.Row - e.ROWS + 1  }
	clear(e.HighlightElements)
}

func (e *Editor) OnUp() {
	if len(e.Content) == 0 { return }
	if e.Row == 0 { e.Y = 0; return }
	e.Row--
	if e.Col > len(e.Content[e.Row]) { e.Col = len(e.Content[e.Row]) } // fit to e.Content
	if e.Row < e.Y { e.Y = e.Row }
	if e.Row > e.Y+ e.ROWS { e.Y = e.Row - e.ROWS + 1  }
	clear(e.HighlightElements)
}

func (e *Editor) OnLeft() {
	if len(e.Content) == 0 { return }

	if e.Col > 0 {
		e.Col--

	} else if e.Row > 0 {
		e.Row--
		e.Col = len(e.Content[e.Row]) // fit to e.Content
		if e.Row < e.Y { e.Y = e.Row }
	}
	clear(e.HighlightElements)
}

func (e *Editor) OnRight() {
	if len(e.Content) == 0 { return }

	if e.Col < len(e.Content[e.Row]) {
		e.Col++
	} else if e.Row < len(e.Content)-1 {
		e.Row++
		e.Col = 0
		if e.Row > e.Y+ e.ROWS { e.Y++  }
	}
	clear(e.HighlightElements)
}

func (e *Editor) OnScrollUp() {
	if len(e.Content) == 0 { return }
	if e.Y == 0 { return }
	e.Y--
	e.Update = true
}

func (e *Editor) OnScrollDown() {
	if len(e.Content) == 0 { return }
	if e.Y+ e.ROWS >= len(e.Content) { return }
	e.Y++
	e.Update = true
}

func (e *Editor) Focus() {
	if e.Row > e.Y+ e.ROWS { e.Y = e.Row + e.ROWS }
	if e.Row < e.Y { e.Y = e.Row }
}
func (e *Editor) FocusCenter() {
	e.Screen.Show()
	if e.Row > e.Y + e.ROWS {
		e.Y = e.Row + e.ROWS
	}
	if e.Row < e.Y {
		e.Y = e.Row
	}

	e.Y -= e.ROWS/2
	if e.Y < 0 { e.Y = 0 }

	centerRow := e.ROWS / 2
	// Update the cursor row to the center row if necessary
	if e.Row - e.Y > centerRow {
		e.Y += e.Row - e.Y - centerRow
	}
}

func (e *Editor) OnEnter() {

	var ops = EditOperation{{Enter, '\n', e.Row, e.Col}}
	tabs := CountTabs(e.Content[e.Row], e.Col)
	spaces := CountSpaces(e.Content[e.Row], e.Col)

	after := e.Content[e.Row][e.Col:]
	before := e.Content[e.Row][:e.Col]
	e.Content[e.Row] = before

	contentToString := ConvertContentToString(e.Content)
	e.treeSitterHighlighter.EnterEdit(contentToString, e.Row, max(e.Col,0))
	e.treeSitterHighlighter.Colors[e.Row] = e.treeSitterHighlighter.Colors[e.Row][:e.Col]

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
	e.Content = InsertTo(e.Content, e.Row, newline)

	e.treeSitterHighlighter.Colors = InsertTo(e.treeSitterHighlighter.Colors, e.Row, make([]int, len(newline)))

	code := ConvertContentToString(e.Content)

	if countToInsert > 0 {
		e.treeSitterHighlighter.AddMultipleCharEdit(code, e.Row, 0, e.Row, countToInsert)
	}

	if e.IsColorize && e.Lang != "" {
		e.treeSitterHighlighter.ColorizeRange(code,  e.Row-1, len(e.Content[e.Row-1]), e.Row, len(newline))
		e.Colors = e.treeSitterHighlighter.Colors
	}

	e.Undo = append(e.Undo, ops)
	e.Focus(); if e.Row- e.Y == e.ROWS { e.OnScrollDown() }
	e.OnCursorChanged()
	if len(e.Redo) > 0 { e.Redo = []EditOperation{} }
	e.Update = true
	e.IsContentChanged = true
	e.FindTests()

	if len(e.Content) <= 10000 { go e.WriteFile() }
}

func (e *Editor) OnDelete() {

	if e.Selection.IsSelectionNonEmpty() {
		e.Cut(false)
		return
	}

	if e.Col > 0 {
		e.Col--
		e.DeleteCharacter(e.Row, e.Col)
		e.OnCursorChanged()
		//e.UpdateColorsAtLine(e.Row)
	} else if e.Row > 0 { // delete line
		e.Undo = append(e.Undo, EditOperation{{DeleteLine, ' ', e.Row -1, len(e.Content[e.Row-1])}})
		left := e.Content[e.Row][e.Col:]
		leftColors := e.Colors[e.Row][e.Col:]
		e.Content = Remove(e.Content, e.Row)
		if e.IsColorize && e.Lang != "" {
			if e.Row < len(e.Colors) {
				e.treeSitterHighlighter.Colors = Remove(e.treeSitterHighlighter.Colors, e.Row)
			}
			//e.UpdateColorsAtLine(e.Row)
		}

		e.Row--
		e.Col = len(e.Content[e.Row])
		e.Content[e.Row] = append(e.Content[e.Row], left...)
		e.treeSitterHighlighter.Colors[e.Row] = append(e.treeSitterHighlighter.Colors[e.Row], leftColors...)


		code := ConvertContentToString(e.Content)
		e.treeSitterHighlighter.RemoveLineEdit(code, e.Row, e.Col)
		e.treeSitterHighlighter.ColorizeRange(code, e.Row, e.Col, e.Row,  len(e.Content[e.Row]))
		e.Colors = e.treeSitterHighlighter.Colors
		e.OnCursorChanged()
		//e.UpdateColorsAtLine(e.Row)
	}

	//code := ConvertContentToString(e.Content)
	//e.Colors = e.treeSitterHighlighter.Colorize(code)

	e.Focus()
	if len(e.Redo) > 0 { e.Redo = []EditOperation{} }
	e.Update = true
	e.IsContentChanged = true
	e.FindTests()
	if len(e.Content) <= 10000 { go e.WriteFile() }
}

func (e *Editor) OnTab() {
	e.Focus()

	selectedLines := e.Selection.GetSelectedLines(e.Content)

	if len(selectedLines) == 0 {
		ch := '\t'
		e.InsertCharacter(e.Row, e.Col, ch)
		e.UpdateColorsAtLine(e.Row)
		e.Col++
		e.OnCursorChanged()
	} else  {
		var ops = EditOperation{}
		e.Selection.Ssx = 0
		for _, linenumber := range selectedLines {
			e.Row = linenumber
			e.Content[e.Row] = InsertTo(e.Content[e.Row], 0, '\t')
			e.UpdateColorsAtLine(e.Row)
			ops = append(ops, Operation{Insert, '\t', e.Row, 0})
			e.Col = len(e.Content[e.Row])
		}
		e.Selection.Sex = e.Col
		e.Undo = append(e.Undo, ops)
	}

	if len(e.Redo) > 0 { e.Redo = []EditOperation{} }
	e.Update = true
	e.IsContentChanged = true
	e.FindTests()
	if len(e.Content) <= 10000 { go e.WriteFile() }
}

func (e *Editor) OnBackTab() {
	e.Focus()

	selectedLines := e.Selection.GetSelectedLines(e.Content)

	// deleting tabs from beginning
	if len(selectedLines) == 0 {
		if e.Content[e.Row][0] == '\t'  {
			e.DeleteCharacter(e.Row,0)
			e.Colors[e.Row] = Remove(e.Colors[e.Row], 0)
			e.Col--
			e.UpdateColorsAtLine(e.Row)
		}
	} else {
		e.Selection.Ssx = 0
		for _, linenumber := range selectedLines {
			e.Row = linenumber
			if len(e.Content[e.Row]) > 0 && e.Content[e.Row][0] == '\t'  {
				e.DeleteCharacter(e.Row,0)
				e.Colors[e.Row] = Remove(e.Colors[e.Row], 0)
				e.Col = len(e.Content[e.Row])
				e.UpdateColorsAtLine(e.Row)
			}
		}
	}

	if len(e.Redo) > 0 { e.Redo = []EditOperation{} }
	e.Update = true
	e.IsContentChanged = true
	e.FindTests()
	if len(e.Content) <= 10000 { go e.WriteFile() }
}

func (e *Editor) AddChar(ch rune) {
	if len(e.Selection.GetSelectionString(e.Content)) != 0 { e.Cut(false) }

	e.Focus()
	e.InsertCharacter(e.Row, e.Col, ch)
	e.Col++

	e.MaybeAddPair(ch)
	e.OnCursorChanged()

	if len(e.Redo) > 0 { e.Redo = []EditOperation{} }

	e.Update = true
	e.IsContentChanged = true
	e.FindTests()
	if len(e.Content) <= 10000 { go e.WriteFile() }
	//e.UpdateColorsAtLine(e.Row)
}

func (e *Editor) InsertCharacter(line, pos int, ch rune) {
	e.Content[line] = InsertTo(e.Content[line], pos, ch)
	e.treeSitterHighlighter.Colors[line] = InsertTo(e.treeSitterHighlighter.Colors[line], pos, -1)
	//if lsp.isReady { go lsp.didChange(AbsoluteFilePath, Line, pos, Line, pos, string(ch)) }
	e.Undo = append(e.Undo, EditOperation{{Insert, ch, e.Row, e.Col}})

	code := ConvertContentToString(e.Content)
	e.treeSitterHighlighter.AddCharEdit(code, line, pos)
	//e.treeSitterHighlighter.ColorizeRange(code, line, line, line, pos)
	e.treeSitterHighlighter.ColorizeRange(code, line, 0, line, len(e.Content[line]))
	e.Colors = e.treeSitterHighlighter.Colors
}

func (e *Editor) InsertString(line, pos int, linestring string) {
	// Convert the string to insert to a slice of runes
	l := RemoveLeadingTabsSpaces(linestring)
	insertRunes := []rune(l)

	// Record the operation on the undo stack. Note that we're creating a new EditOperation
	// and adding all the Operations to it
	var ops = EditOperation{}
	for _, ch := range insertRunes {
		e.Content[line] = InsertTo(e.Content[line], pos, ch)
		ops = append(ops, Operation{Insert, ch, line, pos})
		pos++
	}
	e.Col = pos
	e.Undo = append(e.Undo, ops)
}

func (e *Editor) InsertLines(line, pos int, lines []string) {
	var ops = EditOperation{}

	//tabs := CountTabs(e.Content[e.Row], e.Col) // todo: spaces also can be
	//if len(e.Content[e.Row]) > 0 { e.Row++ }
	//ops = append(ops, Operation{Enter, '\n', e.Row, e.Col})


	lines[0] = string(e.Content[e.Row][:e.Col]) + RemoveLeadingTabsSpaces(lines[0])


	for _, linestr := range lines {
		e.Col = 0
		if e.Row >= len(e.Content)  { e.Content = append(e.Content, []rune{}) } // if last Line adding empty Line before

		//l := RemoveLeadingTabsSpaces(linestr)
		l := linestr
		//nl := strings.Repeat("\t", tabs) + l
		nl := l
		e.Content = InsertTo(e.Content, e.Row, []rune(nl))

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

func (e *Editor) DeleteCharacter(line, pos int) {
	e.Undo = append(e.Undo, EditOperation{
		{MoveCursor, e.Content[line][pos], line, pos+1},
		{Delete, e.Content[line][pos], line, pos},
	})
	e.Content[line] = Remove(e.Content[line], pos)
	e.treeSitterHighlighter.Colors[line] = Remove(e.treeSitterHighlighter.Colors[line], pos)

	code := ConvertContentToString(e.Content)
	e.treeSitterHighlighter.RemoveCharEdit(code, line, pos)
	//e.treeSitterHighlighter.ColorizeRange(code, line, pos, line, pos)
	e.treeSitterHighlighter.ColorizeRange(code, line, 0, line, len(e.Content[line]))

	e.Colors = e.treeSitterHighlighter.Colors
}

func (e *Editor) OnSwapLinesUp() {
	e.Focus()

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
	e.Selection.CleanSelection()
	e.Update = true
	e.IsContentChanged = true
	e.FindTests()
	if len(e.Content) <= 10000 { go e.WriteFile() }
}

func (e *Editor) OnSwapLinesDown() {
	e.Focus()

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
	e.Selection.CleanSelection()
	e.Update = true
	e.IsContentChanged = true
	e.FindTests()
	if len(e.Content) <= 10000 { go e.WriteFile() }
}

func (e *Editor) OnCopy() {
	selectionString := e.Selection.GetSelectionString(e.Content)
	clipboard.WriteAll(selectionString)
}

func (e *Editor) OnSelectMoreAtCursor() {
	var node highlighter.NodeRange

	if !e.Selection.IsSelected || e.TreePath == nil || (e.TreePath.Aty != e.Row || e.TreePath.Atx != e.Col) {
		treepath := e.treeSitterHighlighter.GetNodePathAt(e.Row, e.Col, e.Row, e.Col)
		e.TreePath = &treepath
		node = e.TreePath.CurrentNode()
	} else {
		node = e.TreePath.Next()
	}

	e.Selection.Ssx = node.Ssx; e.Selection.Ssy = node.Ssy
	e.Selection.Sex = node.Sex; e.Selection.Sey = node.Sey
	e.Selection.IsSelected = true
}
func (e *Editor) OnSelectLessAtCursor() {
	if e.TreePath == nil { return }
	node := e.TreePath.Prev()
	e.Selection.Ssx = node.Ssx; e.Selection.Ssy = node.Ssy
	e.Selection.Sex = node.Sex; e.Selection.Sey = node.Sey
	e.Selection.IsSelected = true
}

func (e *Editor) OnSelectAll() {
	if len(e.Content) == 0 { return }
	e.Selection.Ssx = 0; e.Selection.Ssy = 0
	e.Selection.Sey = len(e.Content)
	lastElement := len(e.Content[len(e.Content)-1])
	e.Selection.Sex = lastElement
	e.Selection.Sey = len(e.Content)
	e.Selection.IsSelected = true
}

func (e *Editor) OnPaste() {
	// e.Focus()

	if e.Selection.IsSelectionNonEmpty() {
		e.Cut(false)
	}

	text, _ := clipboard.ReadAll()
	lines := strings.Split(text, "\n")

	if len(lines) == 0 { return }

	if len(lines) == 1 { // single Line paste
		e.InsertString(e.Row, e.Col, lines[0])
	}

	if len(lines) > 1 { // multiple Line paste
		e.InsertLines(e.Row, e.Col, lines)
	}
	
	e.Update = true
	e.UpdateNeeded()
}

func (e *Editor) Cut(isCopySelected bool) {
	e.Focus()

	if len(e.Content) < 1 {
		e.Content[0] = []rune{};
		e.Row, e.Col = 0, 0
		return
	}
	var ops = EditOperation{}

	if len(e.Selection.GetSelectionString(e.Content)) == 0 { // cut single Line
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

		e.Content = Remove(e.Content, e.Row)
		//if e.IsColorize && e.Lang != "" {
		//	if e.Row < len(e.Colors) { e.Colors = Remove(e.Colors, e.Row) }
		//	e.UpdateColorsAtLine(e.Row)
		//}
		if e.Row > 0 { e.Row-- }

		e.Update = true
		e.IsContentChanged = true
		if len(e.Content) <= 10000 { go e.WriteFile() }
		e.UpdateNeeded() // optimize

	} else { // cut selection

		if isCopySelected {
			selectionString := e.Selection.GetSelectionString(e.Content)
			clipboard.WriteAll(selectionString)
		}

		ops = append(ops, Operation{MoveCursor, ' ', e.Row, e.Col})

		selectedIndices := e.Selection.GetSelectedIndices(e.Content)

		// Sort selectedIndices in reverse order to delete characters from the end
		for i := len(selectedIndices) - 1; i >= 0; i-- {
			indices := selectedIndices[i]
			xd := indices[0]
			yd := indices[1]
			e.Col, e.Row = xd, yd

			if len(e.Content[yd]) > 0 {
				// Delete the character at index (x, j)
				ops = append(ops, Operation{Delete, e.Content[yd][xd], yd, xd})
				e.Content[yd] = append(e.Content[yd][:xd], e.Content[yd][xd+1:]...)
				//e.Colors[yd] = append(e.Colors[yd][:xd], e.Colors[yd][xd+1:]...)
			}


			if len(e.Content[yd]) == 0 { // delete Line
				if e.Row == 0 { ops = append(ops, Operation{DeleteLine, '\n', 0, 0}) } else {
					ops = append(ops, Operation{DeleteLine, '\n', e.Row -1, len(e.Content[e.Row-1])})
				}

				e.Content = append(e.Content[:yd], e.Content[yd+1:]...)
				//e.Colors = append(e.Colors[:yd], e.Colors[yd+1:]...)
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
		e.Selection.CleanSelection()
		e.Update = true
		e.IsContentChanged = true
		if len(e.Content) <= 10000 { go e.WriteFile() }

		e.UpdateNeeded() // optimize
	}

	e.Undo = append(e.Undo, ops)
}

func (e *Editor) Duplicate() {
	e.Focus()

	if len(e.Content) == 0 { return }

	if e.Selection.Ssx == -1 && e.Selection.Ssy == -1 ||
		e.Selection.Ssx == e.Selection.Sex && e.Selection.Ssy == e.Selection.Sey {
		var ops = EditOperation{}
		ops = append(ops, Operation{MoveCursor, ' ', e.Row, e.Col})
		ops = append(ops, Operation{Enter, '\n', e.Row, len(e.Content[e.Row])})

		duplicatedSlice := make([]rune, len(e.Content[e.Row]))
		copy(duplicatedSlice, e.Content[e.Row])
		for i, ch := range duplicatedSlice {
			ops = append(ops, Operation{Insert, ch, e.Row, i})
		}
		e.Row++
		e.Content = InsertTo(e.Content, e.Row, duplicatedSlice)
		if e.IsColorize && e.Lang != "" {
			e.Colors = InsertTo(e.Colors, e.Row, []int{})
			e.UpdateColorsAtLine(e.Row)
		}
		e.Undo = append(e.Undo, ops)
		e.Update = true
		e.IsContentChanged = true
		e.FindTests()
		if len(e.Content) <= 10000 { go e.WriteFile() }

	} else {
		selection := e.Selection.GetSelectionString(e.Content)
		if len(selection) == 0 { return }
		lines := strings.Split(selection, "\n")

		if len(lines) == 0 { return }

		if len(lines) == 1 { // single Line
			lines[0] = " " + lines[0]// add space before
			e.InsertString(e.Row, e.Col, lines[0])
		}

		if len(lines) > 1 { // multiple Line
			e.InsertLines(e.Row, e.Col, lines)
		}
		e.Selection.CleanSelection()
		e.UpdateNeeded()
	}

}

func (e *Editor) OnCursorBackUndo() {
	if len(e.CursorHistoryUndo) == 0 { return }

	lastCursor := e.CursorHistoryUndo[len(e.CursorHistoryUndo)-1]
	e.CursorHistoryUndo = e.CursorHistoryUndo[:len(e.CursorHistoryUndo)-1]


	if lastCursor.Filename != e.Filename {
		e.OpenFile(lastCursor.Filename)
	}

	e.Row = lastCursor.Row
	e.Col = lastCursor.Col
	e.Y = lastCursor.Y
	e.X = lastCursor.X
	e.Focus()
	e.OnCursorChanged()
	e.CursorHistory = append(e.CursorHistory, lastCursor)
}
func (e *Editor) OnCursorBack() {
	if len(e.CursorHistory) == 0 { return }

	lastCursor := e.CursorHistory[len(e.CursorHistory)-1]
	e.CursorHistory = e.CursorHistory[:len(e.CursorHistory)-1]
	e.CursorHistoryUndo = append(e.CursorHistoryUndo,
		 CursorMove{e.AbsoluteFilePath, e.Row, e.Col, e.Y, e.X},
	)

	if lastCursor.Filename != e.Filename {
		e.OpenFile(lastCursor.Filename)
	}

	e.Row = lastCursor.Row
	e.Col = lastCursor.Col
	e.Y = lastCursor.Y
	e.X = lastCursor.X
	e.Focus()
	e.OnCursorChanged()
}

func (e *Editor) OnUndo() {
	if len(e.Undo) == 0 { return }

	lastOperation := e.Undo[len(e.Undo)-1]
	e.Undo = e.Undo[:len(e.Undo)-1]
	e.Focus()
	for i := len(lastOperation) - 1; i >= 0; i-- {
		o := lastOperation[i]

		if o.Action == Insert {
			e.Row = o.Line; e.Col = o.Column
			e.Content[e.Row] = append(e.Content[e.Row][:e.Col], e.Content[e.Row][e.Col+1:]...)

		} else if o.Action == Delete {
			e.Row = o.Line; e.Col = o.Column
			e.Content[e.Row] = InsertTo(e.Content[e.Row], e.Col, o.Char)

		} else if o.Action == Enter {
			// Merge lines
			e.Content[o.Line] = append(e.Content[o.Line], e.Content[o.Line+1]...)
			e.Content = append(e.Content[:o.Line+1], e.Content[o.Line+2:]...)
			e.Row = o.Line; e.Col = o.Column

		} else if o.Action == DeleteLine {
			// Insert enter
			e.Row = o.Line; e.Col = o.Column
			after := e.Content[e.Row][e.Col:]
			before := e.Content[e.Row][:e.Col]
			e.Content[e.Row] = before
			e.Row++; e.Col = 0
			newline := append([]rune{}, after...)
			e.Content = InsertTo(e.Content, e.Row, newline)
		} else if o.Action == MoveCursor {
			e.Row = o.Line; e.Col = o.Column
		}
		e.OnCursorChanged()
	}

	e.Redo = append(e.Redo, lastOperation)
	e.UpdateNeeded()
}
func (e *Editor) OnRedo() {
	if len(e.Redo) == 0 { return }

	lastRedoOperation := e.Redo[len(e.Redo)-1]
	e.Redo = e.Redo[:len(e.Redo)-1]

	for i := 0; i < len(lastRedoOperation); i++ {
		o := lastRedoOperation[i]

		if o.Action == Insert {
			e.Row = o.Line; e.Col = o.Column
			e.Content[e.Row] = InsertTo(e.Content[e.Row], e.Col, o.Char)
			e.Col++
		} else if o.Action == Delete {
			e.Row = o.Line; e.Col = o.Column
			e.Content[e.Row] = append(e.Content[e.Row][:e.Col], e.Content[e.Row][e.Col+1:]...)
		} else if o.Action == Enter {
			e.Row = o.Line; e.Col = o.Column
			after := e.Content[e.Row][e.Col:]
			before := e.Content[e.Row][:e.Col]
			e.Content[e.Row] = before
			e.Row++; e.Col = 0
			newline := append([]rune{}, after...)
			e.Content = InsertTo(e.Content, e.Row, newline)
		} else if o.Action == DeleteLine {
			// Merge lines
			e.Content[o.Line] = append(e.Content[o.Line], e.Content[o.Line+1]...)
			e.Content = append(e.Content[:o.Line+1], e.Content[o.Line+2:]...)
			e.Row = o.Line; e.Col = o.Column
		} else if o.Action == MoveCursor {
			e.Row = o.Line; e.Col = o.Column
		}
	}

	e.Undo = append(e.Undo, lastRedoOperation)
	e.UpdateNeeded()
}
func (e *Editor) OnCommentLine() {
	e.Focus()

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
			e.Content[e.Row] = Remove(e.Content[e.Row], i)

			e.treeSitterHighlighter.Colors[e.Row] = Remove(e.treeSitterHighlighter.Colors[e.Row], i)
			for index := range e.treeSitterHighlighter.Colors[e.Row] { // reset colors on line to default
				e.treeSitterHighlighter.Colors[e.Row][index] = -1
			}

			code := ConvertContentToString(e.Content)
			e.treeSitterHighlighter.RemoveCharEdit(code, e.Row, i)
			e.treeSitterHighlighter.ColorizeRange(code, e.Row, 0, e.Row, len(e.Content[e.Row]))
			e.Colors = e.treeSitterHighlighter.Colors

			//e.UpdateColorsAtLine(e.Row)
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
			e.Content[e.Row] = Remove(e.Content[e.Row], i)
			e.Content[e.Row] = Remove(e.Content[e.Row], i)

			e.treeSitterHighlighter.Colors[e.Row] = Remove(e.treeSitterHighlighter.Colors[e.Row], i)
			e.treeSitterHighlighter.Colors[e.Row] = Remove(e.treeSitterHighlighter.Colors[e.Row], i)

			for index := range e.treeSitterHighlighter.Colors[e.Row] { // reset colors on line to default
				e.treeSitterHighlighter.Colors[e.Row][index] = -1
			}

			code := ConvertContentToString(e.Content)
			e.treeSitterHighlighter.RemoveCharEdit(code, e.Row, i)
			e.treeSitterHighlighter.RemoveCharEdit(code, e.Row, i)
			e.treeSitterHighlighter.ColorizeRange(code, e.Row, 0, e.Row, len(e.Content[e.Row]))
			e.Colors = e.treeSitterHighlighter.Colors

			//e.UpdateColorsAtLine(e.Row)
			found = true
			break
		}
	}

	if found {
		if e.Col < 0 { e.Col = 0 }
		e.OnDown()
		e.Update = true
		e.IsContentChanged = true
		if len(e.Content) <= 10000 { go e.WriteFile() }
		return
	}

	tabs := CountTabs(e.Content[e.Row], e.Col)
	spaces := CountSpaces(e.Content[e.Row], e.Col)

	from := tabs
	if tabs == 0 && spaces != 0 { from = spaces }

	e.Col = from
	ops := EditOperation{}
	for _, ch := range e.langConf.Comment {
		e.Content[e.Row] = InsertTo(e.Content[e.Row], from, ch)
		e.treeSitterHighlighter.Colors[e.Row] = InsertTo(e.treeSitterHighlighter.Colors[e.Row], from, -1)
		code := ConvertContentToString(e.Content)
		e.treeSitterHighlighter.AddCharEdit(code, e.Row, from)
		e.treeSitterHighlighter.ColorizeRange(code, e.Row, from, e.Row, from)
		e.Colors = e.treeSitterHighlighter.Colors
		ops = append(ops, Operation{Insert, ch, e.Row, from})
	}

	code := ConvertContentToString(e.Content)
	e.treeSitterHighlighter.ColorizeRange(code, e.Row, 0, e.Row, len(e.Content[e.Row]))
	e.Colors = e.treeSitterHighlighter.Colors
	//e.UpdateColorsAtLine(e.Row)

	e.Undo = append(e.Undo, ops)
	if e.Col < 0 { e.Col = 0 }
	e.OnDown()
	e.Update = true
	e.IsContentChanged = true
	if len(e.Content) <= 10000 { go e.WriteFile() }
}

func (e *Editor) HandleSmartMove(char rune) {
	e.Focus()
	if char == 'f' || char == 'F' {
		nw := FindNextWord(e.Content[e.Row], e.Col+ 1)
		e.Col = nw
		e.Col = Min(e.Col, len(e.Content[e.Row]))
	}
	if char == 'b' || char == 'B' {
		nw := FindPrevWord(e.Content[e.Row], e.Col-1)
		e.Col = nw
	}
}

func (e *Editor) HandleSmartMoveDown() {

	var ops = EditOperation{{Enter, '\n', e.Row, e.Col}}

	// moving down, insert new Line, add same amount of tabs
	tabs := CountTabs(e.Content[e.Row], e.Col)
	spaces := CountSpaces(e.Content[e.Row], e.Col)

	countToInsert := tabs
	characterToInsert := '\t'
	if tabs == 0 && spaces != 0 { characterToInsert = ' '; countToInsert = spaces }

	e.Row++; e.Col = 0
	e.Content = InsertTo(e.Content, e.Row, []rune{})
	for i := 0; i < countToInsert; i++ {
		e.Content[e.Row] = InsertTo(e.Content[e.Row], e.Col, characterToInsert)
		ops = append(ops, Operation{Insert, characterToInsert, e.Row, e.Col})
		e.Col++
	}

	if e.IsColorize && e.Lang != "" {
		e.Colors = InsertTo(e.Colors, e.Row, []int{})
		e.UpdateColorsAtLine(e.Row)
	}

	e.Focus(); e.OnScrollDown()
	e.Undo = append(e.Undo, ops)
	e.Update = true
	e.IsContentChanged = true
	if len(e.Content) <= 10000 { go e.WriteFile() }
}

func (e *Editor) HandleSmartMoveUp() {
	e.Focus()
	// add new Line and shift all lines, add same amount of tabs/spaces
	tabs := CountTabs(e.Content[e.Row], e.Col)
	spaces := CountSpaces(e.Content[e.Row], e.Col)

	countToInsert := tabs
	characterToInsert := '\t'
	if tabs == 0 && spaces != 0 { characterToInsert = ' '; countToInsert = spaces }

	var ops = EditOperation{{Enter, '\n', e.Row, e.Col}}
	e.Content = InsertTo(e.Content, e.Row, []rune{})

	e.Col = 0
	for i := 0; i < countToInsert; i++ {
		e.Content[e.Row] = InsertTo(e.Content[e.Row], e.Col, characterToInsert)
		ops = append(ops, Operation{Insert, characterToInsert, e.Row, e.Col})
		e.Col++
	}

	if e.IsColorize && e.Lang != "" {
		e.Colors = InsertTo(e.Colors, e.Row, []int{})
		e.UpdateColorsAtLine(e.Row)
	}

	e.Undo = append(e.Undo, ops)
	e.Update = true
	e.IsContentChanged = true
	if len(e.Content) <= 10000 { go e.WriteFile() }
}

func (e *Editor) MaybeAddPair(ch rune) {
	pairMap := map[rune]rune{
		'(': ')', '{': '}', '[': ']',
		'"': '"', '\'': '\'', '`': '`',
	}

	if closeChar, found := pairMap[ch]; found {
		noMoreChars := e.Col >= len(e.Content[e.Row])
		isSpaceNext := e.Col < len(e.Content[e.Row]) && e.Content[e.Row][e.Col] == ' '
		isStringAndClosedBracketNext := closeChar == '"' && e.Col < len(e.Content[e.Row]) && e.Content[e.Row][e.Col] == ')'

		if noMoreChars || isSpaceNext || isStringAndClosedBracketNext {
			e.InsertCharacter(e.Row, e.Col, closeChar)
		}
	}
}
