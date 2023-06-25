package main

import (
	"fmt"
	. "github.com/gdamore/tcell"
	"sort"
	"strings"
	"time"
)

func (e *Editor) OnDefinition() {
	definition, err := Lsp.definition(e.AbsoluteFilePath, e.Row, e.Col)

	if err != nil || len(definition.Result) == 0{
		return
	}

	if definition.Result[0].URI != "file://" + e.AbsoluteFilePath {
		e.InputFile = strings.Split(definition.Result[0].URI, "file://")[1]
		e.OpenFile(e.InputFile)
	}

	if int(definition.Result[0].Range.Start.Line) > len(e.Content) || // not out of e.Content
		int(definition.Result[0].Range.Start.Character) > len(e.Content[int(definition.Result[0].Range.Start.Line)]) {
		return
	}

	e.Row = int(definition.Result[0].Range.Start.Line)
	e.Col = int(definition.Result[0].Range.Start.Character)
	e.Selection.Ssx = e.Col; e.Selection.Ssy = e.Row;
	e.Selection.Sey = int(definition.Result[0].Range.End.Line)
	e.Selection.Sex = int(definition.Result[0].Range.End.Character)
	e.Row = e.Selection.Sey; e.Col = e.Selection.Sex
	e.Selection.IsSelected = true
	e.Focus()
}

func (e *Editor) OnHover() {
	if !Lsp.IsLangReady(e.Lang) { return }

	e.IsOverlay = true
	defer e.OverlayFalse()

	var hoverEnd = false

	// loop until escape or enter pressed
	for !hoverEnd {

		start := time.Now()
		hover, err := Lsp.hover(e.AbsoluteFilePath, e.Row, e.Col)
		elapsed := time.Since(start)

		lspStatus := "lsp hover, elapsed " + elapsed.String()
		status := fmt.Sprintf(" %s %s %d %d %s ", lspStatus, e.Lang, e.Row+ 1, e.Col+ 1, e.InputFile)
		e.DrawStatus(status)

		if err != nil || len(hover.Result.Contents.Value) == 0 { return }
		options := strings.Split(hover.Result.Contents.Value, "\n")
		if len(options) == 0 { return }

		tabs := CountTabsTo(e.Content[e.Row], e.Col)
		width := Max(30, MaxString(options))                                                                             // width depends on max option len or 30 at min
		height := MinMany(10, len(options))                                                                              // depends on min option len or 5 at min or how many rows to the end of e.Screen
		atx := (e.Col - tabs) + e.LINES_WIDTH + tabs * (e.langTabWidth) + e.FilesPanelWidth; aty := e.Row - height - e.Y // Define the window  position and dimensions
		style := StyleDefault.Foreground(ColorWhite)
		if len(options) > e.Row- e.Y { aty = e.Row + 1 }

		var selectionEnd = false; var selected = 0; var selectedOffset = 0

		for !selectionEnd {
			if selected < selectedOffset { selectedOffset = selected }  // calculate offsets for scrolling completion
			if selected >= selectedOffset+height { selectedOffset = selected - height + 1 }

			e.drawCompletion(atx,aty, height, width, options, selected, selectedOffset, style)
			e.Screen.Show()

			switch ev := e.Screen.PollEvent().(type) { // poll and handle event
			case *EventKey:
				key := ev.Key()
				if key == KeyEscape || key == KeyEnter ||
					key == KeyBackspace || key == KeyBackspace2 {
					e.Screen.Clear(); selectionEnd = true; hoverEnd = true
				}
				if key == KeyDown { selected = Min(len(options)-1, selected+1) }
				if key == KeyUp { selected = Max(0, selected-1) }
				if key == KeyRight { e.OnRight(); e.Screen.Clear(); e.DrawEverything(); selectionEnd = true }
				if key == KeyLeft { e.OnLeft(); e.Screen.Clear(); e.DrawEverything(); selectionEnd = true }
			}
		}
	}
}

func (e *Editor) OnSignatureHelp() {
	if !Lsp.IsLangReady(e.Lang) { return }

	e.IsOverlay = true
	defer e.OverlayFalse()

	var end = false

	// loop until escape or enter pressed
	for !end {

		start := time.Now()
		signatureHelpResponse, err := Lsp.signatureHelp(e.AbsoluteFilePath, e.Row, e.Col)
		elapsed := time.Since(start)

		lspStatus := "lsp signature help, elapsed " + elapsed.String()
		status := fmt.Sprintf(" %s %s %d %d %s ", lspStatus, e.Lang, e.Row+1, e.Col+ 1, e.InputFile)
		e.DrawStatus(status)

		if err != nil || signatureHelpResponse.Result.Signatures == nil ||
			len(signatureHelpResponse.Result.Signatures) == 0 { return }

		var options = []string{}
		for _, signature := range signatureHelpResponse.Result.Signatures {
			var text = []string {}
			for _, parameter := range signature.Parameters {
				text =  append(text, parameter.Label)
			}
			options = append(options, strings.Join(text, ", "))
		}

		if len(options) == 0 { return }

		tabs := CountTabsTo(e.Content[e.Row], e.Col)
		width := Max(30, MaxString(options))                                                                           // width depends on max option len or 30 at min
		height := MinMany(10, len(options))                                                                            // depends on min option len or 5 at min or how many rows to the end of e.Screen
		atx := (e.Col - tabs) + e.LINES_WIDTH + tabs*(e.langTabWidth) + e.FilesPanelWidth; aty := e.Row - height - e.Y // Define the window  position and dimensions
		style := StyleDefault.Foreground(ColorWhite)
		if len(options) > e.Row- e.Y { aty = e.Row + 1 }

		var selectionEnd = false; var selected = 0; var selectedOffset = 0

		for !selectionEnd {
			if selected < selectedOffset { selectedOffset = selected }  // calculate offsets for scrolling completion
			if selected >= selectedOffset+height { selectedOffset = selected - height + 1 }

			e.drawCompletion(atx,aty, height, width, options, selected, selectedOffset, style)
			e.Screen.Show()

			switch ev := e.Screen.PollEvent().(type) { // poll and handle event
			case *EventKey:
				key := ev.Key()
				if key == KeyEscape || key == KeyEnter ||
					key == KeyBackspace || key == KeyBackspace2 { e.Screen.Clear(); selectionEnd = true; end = true }
				if key == KeyDown { selected = Min(len(options)-1, selected+1) }
				if key == KeyUp { selected = Max(0, selected-1) }
				if key == KeyRight { e.OnRight(); e.Screen.Clear(); e.DrawEverything(); selectionEnd = true }
				if key == KeyLeft { e.OnLeft(); e.Screen.Clear(); e.DrawEverything(); selectionEnd = true }
				if key == KeyRune { e.AddChar(ev.Rune()); e.WriteFile(); e.Screen.Clear(); e.DrawEverything(); selectionEnd = true  }
			}
		}
	}
}

func (e *Editor) OnReferences() {
	if !Lsp.IsLangReady(e.Lang) { return }

	e.IsOverlay = true
	defer e.OverlayFalse()

	var end = false

	// loop until escape or enter pressed
	for !end {

		start := time.Now()
		referencesResponse, err := Lsp.references(e.AbsoluteFilePath, e.Row, e.Col)
		elapsed := time.Since(start)

		lspStatus := "lsp references, elapsed " + elapsed.String()
		status := fmt.Sprintf(" %s %s %d %d %s ", lspStatus, e.Lang, e.Row+ 1, e.Col+ 1, e.InputFile)
		e.DrawStatus(status)

		if err != nil || len(referencesResponse.Result) == 0 { return }

		var options = []string{}
		for i, ref := range referencesResponse.Result {
			text := fmt.Sprintf("%d/%d %s %d %d ", i+1, len(referencesResponse.Result),
				ref.URI, ref.Range.Start.Line + 1, ref.Range.Start.Character + 1,
			)
			options = append(options, text)
		}

		if len(options) == 0 { return }
		if len(options) == 1 {
			// if only one option, no need to draw options
			referencesResult := referencesResponse.Result[0]
			e.applyReferences(referencesResult)
			return
		}

		tabs := CountTabsTo(e.Content[e.Row], e.Col)
		width := Max(30, MaxString(options))                                                                          // width depends on max option len or 30 at min
		height := MinMany(10, len(options))                                                                           // depends on min option len or 5 at min or how many rows to the end of e.Screen
		atx := (e.Col -tabs) + e.LINES_WIDTH + tabs*(e.langTabWidth) + e.FilesPanelWidth; aty := e.Row - height - e.Y // Define the window  position and dimensions
		style := StyleDefault.Foreground(ColorWhite)
		if len(options) > e.Row- e.Y { aty = e.Row + 1 }

		var selectionEnd = false; var selected = 0; var selectedOffset = 0

		for !selectionEnd {
			if selected < selectedOffset { selectedOffset = selected }  // calculate offsets for scrolling completion
			if selected >= selectedOffset+height { selectedOffset = selected - height + 1 }

			e.drawCompletion(atx,aty, height, width, options, selected, selectedOffset, style)
			e.Screen.Show()

			switch ev := e.Screen.PollEvent().(type) { // poll and handle event
			case *EventKey:
				key := ev.Key()
				if key == KeyEscape || key == KeyEnter ||
					key == KeyBackspace || key == KeyBackspace2 { e.Screen.Clear(); selectionEnd = true; end = true }
				if key == KeyDown { selected = Min(len(options)-1, selected+1) }
				if key == KeyUp { selected = Max(0, selected-1) }
				if key == KeyRight { e.OnRight(); e.Screen.Clear(); e.DrawEverything(); selectionEnd = true }
				if key == KeyLeft { e.OnLeft(); e.Screen.Clear(); e.DrawEverything(); selectionEnd = true }
				if key == KeyRune { e.AddChar(ev.Rune()); e.WriteFile(); e.Screen.Clear(); e.DrawEverything(); selectionEnd = true  }
				if key == KeyEnter {
					selectionEnd = true
					referencesResult := referencesResponse.Result[selected]
					e.applyReferences(referencesResult)
				}
			}
		}
	}
}

func (e *Editor) applyReferences(referencesResult ReferencesRange) {
	if referencesResult.URI != "file://"+ e.AbsoluteFilePath { // if another file
		e.InputFile = strings.Split(referencesResult.URI, "file://")[1]
		e.OpenFile(e.InputFile)
	}

	e.Row = referencesResult.Range.Start.Line
	e.Col = referencesResult.Range.Start.Character
	e.Selection.Ssx = e.Col; e.Selection.Ssy = e.Row;
	e.Selection.Sey = referencesResult.Range.End.Line
	e.Selection.Sex = referencesResult.Range.End.Character
	e.Selection.IsSelected = true
	e.Row = e.Selection.Sey; e.Col = e.Selection.Sex
	e.Focus();
	e.DrawEverything()
}


func (e *Editor) OnCompletion() {
	if !Lsp.IsLangReady(e.Lang) { return }
	e.IsOverlay = true
	defer e.OverlayFalse()

	var completionEnd = false

	// loop until escape or enter pressed
	for !completionEnd {

		start := time.Now()
		completion, err := Lsp.completion(e.AbsoluteFilePath, e.Row, e.Col)
		elapsed := time.Since(start)

		lspStatus := "lsp completion, elapsed " + elapsed.String()
		status := fmt.Sprintf(" %s %s %d %d %s ", lspStatus, e.Lang, e.Row+1, e.Col+1, e.Filename)
		e.DrawStatus(status)

		options := e.buildCompletionOptions(completion)
		if err != nil || len(options) == 0 { return }

		tabs := CountTabsTo(e.Content[e.Row], e.Col)
		atx := (e.Col - tabs) + e.LINES_WIDTH + tabs*(e.langTabWidth) + e.FilesPanelWidth; aty := e.Row + 1 - e.Y // Define the window  position and dimensions
		width := Max(30, MaxString(options))                                                                      // width depends on Max option len or 30 at min
		height := MinMany(5, len(options), e.ROWS - (e.Row- e.Y))                                                 // depends on min option len or 5 at min or how many rows to the end of e.Screen
		style := StyleDefault
		// if completion on last two rows of the e.Screen - move window up
		if e.Row- e.Y >= e.ROWS - 1 { aty -= Min(5, len(options)); aty--; height = Min(5, len(options)) }

		var selectionEnd = false; var selected = 0; var selectedOffset = 0


		for !selectionEnd {
			if selected < selectedOffset { selectedOffset = selected }  // calculate offsets for scrolling completion
			if selected >= selectedOffset+height { selectedOffset = selected - height + 1 }

			e.drawCompletion(atx,aty, height, width, options, selected, selectedOffset, style)
			e.Screen.Show()

			switch ev := e.Screen.PollEvent().(type) { // poll and handle event
			case *EventKey:
				key := ev.Key()
				if key == KeyEscape || key == KeyCtrlSpace { selectionEnd = true; completionEnd = true }
				if key == KeyDown { selected = Min(len(options)-1, selected+1); e.Screen.Clear(); e.DrawEverything() }
				if key == KeyUp { selected = Max(0, selected-1); e.Screen.Clear(); e.DrawEverything(); }
				if key == KeyRight { e.OnRight(); e.Screen.Clear(); e.DrawEverything(); selectionEnd = true }
				if key == KeyLeft { e.OnLeft(); e.Screen.Clear(); e.DrawEverything(); selectionEnd = true }
				if key == KeyRune { e.AddChar(ev.Rune()); e.Screen.Clear(); e.DrawEverything(); selectionEnd = true  }
				if key == KeyBackspace || key == KeyBackspace2 {
					e.OnDelete(); e.Screen.Clear(); e.DrawEverything(); selectionEnd = true
				}
				if key == KeyEnter {
					selectionEnd = true; completionEnd = true
					e.completionApply(completion, selected)
					e.Screen.Show()
				}
			}
		}
	}
}


func (e *Editor) buildCompletionOptions(completion CompletionResponse) []string {
	var options []string
	var maxOptlen = 5

	prev := FindPrevWord(e.Content[e.Row], e.Col)
	filterword := string(e.Content[e.Row][prev:e.Col])

	sortItemsByMatchCount(&completion.Result, filterword)

	for _, item := range completion.Result.Items {
		if len(item.Label) > maxOptlen { maxOptlen = len(item.Label) }
	}
	for _, item := range completion.Result.Items {
		options = append(options, FormatText(item.Label, item.Detail, maxOptlen))
	}

	if options == nil { options = []string{} }
	return options
}

func sortItemsByMatchCount(cr *CompletionResult, matchStr string) {
	sort.Slice(cr.Items, func(i, j int) bool {
		return scoreMatches(cr.Items[i].Label, matchStr) > scoreMatches(cr.Items[j].Label, matchStr)
	})
}

// scoreMatches applies a scoring system based on different matching scenarios.
func scoreMatches(src, matchStr string) int {
	score := 0

	// If the match is at the beginning, we give it a high score.
	if strings.HasPrefix(src, matchStr) { score += 1000 }

	// Each occurrence of matchStr in src adds a smaller score.
	score += strings.Count(src, matchStr) * 10

	// Subtracting the square of the length of the string to give shorter strings a bigger boost.
	//score -= len(src) * len(src) * len(src)


	// If match is close to the Start of the string but not at the beginning, add some score.
	initialIndex := strings.Index(src, matchStr)
	if initialIndex > 0 && initialIndex < 5 { score += 500 }

	return score
}

func (e *Editor) drawCompletion(atx int, aty int, height int, width int,
	options []string, selected int, selectedOffset int, style Style) {

	for row := 0; row < aty+height; row++ {
		if row >= len(options) || row >= height { break }
		var option = options[row+selectedOffset]

		isRowSelected := selected == row+selectedOffset
		if isRowSelected { style = style.Background(Color(AccentColor)) } else {
			style = StyleDefault.Background(Color(OverlayColor))
		}
		//style = style.Foreground(ColorWhite)

		e.Screen.SetContent(atx-1, row+aty, ' ', nil, style)
		for col, char := range option {
			e.Screen.SetContent(col+atx, row+aty, char, nil, style)
		}
		for col := len(option); col < width; col++ { // Fill the remaining space
			e.Screen.SetContent(col+atx, row+aty, ' ', nil, style)
		}
	}
}

func (e *Editor) completionApply(completion CompletionResponse, selected int) {
	// parse completion
	item := completion.Result.Items[selected]
	from := item.TextEdit.Range.Start.Character
	end := item.TextEdit.Range.End.Character
	newText := item.TextEdit.NewText

	if item.TextEdit.Range.Start.Character != 0 && item.TextEdit.Range.End.Character != 0 {
		// text edit supported by lsp server
		// move cursor to beginning
		e.Col = int(from)
		// remove chars between from and end
		e.Content[e.Row] = append(e.Content[e.Row][:e.Col], e.Content[e.Row][int(end):]...)
		newText = item.TextEdit.NewText
	}

	if from == 0 && end == 0 {
		// text edit not supported by lsp
		prev := FindPrevWord(e.Content[e.Row], e.Col)
		next := FindNextWord(e.Content[e.Row], e.Col)
		from = float64(prev)
		newText = item.InsertText
		if len(newText) == 0 { newText = item.Label }
		end = float64(next)
		e.Col = prev
		e.Content[e.Row] = append(e.Content[e.Row][:e.Col], e.Content[e.Row][int(end) :]...)
	}

	// add newText
	for _, char := range newText { e.InsertCharacter(e.Row, e.Col, char); e.Col++ }
	e.UpdateColorsAtLine(e.Row)

	e.Update = true
	e.IsContentChanged = true
	if len(e.Content) <= 10000 { go e.WriteFile() }
}

func (e *Editor) OnRename() {
	var end = false
	var renameTo = []rune{}
	var patternx = 0
	var prefix = []rune("rename: ")

	prepareRenameResponse, err := Lsp.prepareRename(e.AbsoluteFilePath, e.Row, e.Col)
	if err != nil  { return }
	placeHolder := prepareRenameResponse.Result.Placeholder
	renameTo = []rune(placeHolder)
	patternx = len(renameTo)

	// loop until escape or enter pressed
	for !end {

		for i := 0; i < len(prefix); i++ {
			e.Screen.SetContent(e.FilesPanelWidth+ e.LINES_WIDTH+ i, e.ROWS-1, prefix[i], nil, StyleDefault)
		}

		e.Screen.SetContent(e.FilesPanelWidth+ e.LINES_WIDTH+ len(prefix), e.ROWS-1, ' ', nil, StyleDefault)

		for i := 0; i < len(renameTo); i++ {
			e.Screen.SetContent(e.FilesPanelWidth+ e.LINES_WIDTH+ len(prefix) + i, e.ROWS-1, renameTo[i], nil, StyleDefault)
		}

		e.Screen.ShowCursor(e.FilesPanelWidth+ e.LINES_WIDTH+ len(prefix) + patternx , e.ROWS-1)

		for i := e.FilesPanelWidth + e.LINES_WIDTH + len(prefix) + len(renameTo); i < e.COLUMNS; i++ {
			e.Screen.SetContent(i, e.ROWS-1, ' ', nil, StyleDefault)
		}

		e.Screen.Show()


		switch ev := e.Screen.PollEvent().(type) { // poll and handle event
		case *EventResize:
			e.COLUMNS, e.ROWS = e.Screen.Size()

		case *EventKey:
			key := ev.Key()

			if key == KeyRune {
				renameTo = InsertTo(renameTo, patternx, ev.Rune())
				patternx++
			}
			if key == KeyBackspace2 && patternx > 0 && len(renameTo) > 0 {
				patternx--
				renameTo = Remove(renameTo, patternx)
			}
			if key == KeyLeft  && patternx > 0 { patternx-- }
			if key == KeyRight && patternx < len(renameTo) { patternx++ }
			if key == KeyESC  || key == KeyCtrlF { end = true }
			if key == KeyEnter {
				renameResponse, err := Lsp.rename(e.AbsoluteFilePath, string(renameTo), e.Row, e.Col)
				if err != nil  { return }
				e.applyRename(renameResponse)
				end = true
			}
		}
	}
}

func (e *Editor) applyRename(renameResponse RenameResponse) {
	inputFileTmp := e.InputFile
	tmpr := e.Row; tmpc := e.Col;  tmpy := e.Y;  tmpx := e.X;

	for _, dc := range renameResponse.Result.DocumentChanges {
		if dc.TextDocument.URI != "file://" + e.AbsoluteFilePath {
			e.InputFile = strings.Split(dc.TextDocument.URI, "file://")[1]
			e.OpenFile(e.InputFile)
		}
		if dc.TextDocument.URI == "file://"+e.AbsoluteFilePath {
			// apply the changes in reverse order, its matter
			for ei := len(dc.Edits) - 1; ei >= 0; ei-- {
				edit := dc.Edits[ei]
				line := int(edit.Range.Start.Line)
				startc := int(edit.Range.Start.Character)
				endc := int(edit.Range.End.Character)

				// replace the old text with the new text
				after := e.Content[line][endc:]
				newText := []rune(edit.NewText)
				newTextAndAfter := append(newText, after...)
				before := e.Content[line][:startc]
				wholeNewLine := append(before, newTextAndAfter...)
				e.Content[line] = wholeNewLine

				e.UpdateColorsAtLine(line)
			}

			e.WriteFile()
		}
	}

	if e.InputFile != inputFileTmp {
		e.InputFile = inputFileTmp
		e.OpenFile(e.InputFile)
		e.Row = tmpr; e.Col = tmpc;  e.Y = tmpy;  e.X = tmpx;
	}

}


func (e *Editor) OnCodeAction() {
	codeAction, err := Lsp.codeAction(e.AbsoluteFilePath, e.Selection.Ssx, e.Selection.Ssy, e.Selection.Sex, e.Selection.Sey)
	if err != nil { return }
	if len(codeAction.Result) == 0 { return }
	//
	commandResponse, err := Lsp.command(codeAction.Result[0].Command)
	if err != nil { return }
	if len(commandResponse.Params.Edit.DocumentChanges) == 0 { return }

	e.handleEdits(commandResponse.Params.Edit.DocumentChanges[0].Edits, commandResponse.Params.Edit.DocumentChanges[0].TextDocument.Version+1)
	Lsp.applyEdit(commandResponse.ID)
}

func (e *Editor) handleEdits(edits []Edit, version int) {
	for _, edit := range edits {
		start := edit.Range.Start
		//end := edit.Range.End

		// Adjusting because slices are 0-indexed.
		//startLine, startChar := int(Start.Line), int(Start.Character)
		//endLine, endChar := int(end.Line), int(end.Character)

		newLines := strings.Split(edit.NewText, "\n")
		for i, line := range newLines {
			index := int(start.Line) + i
			if index >= len(e.Content) {
				e.Content = append(e.Content,  []rune(line))
			} else  {
				e.Content[index] = []rune(line)
			}
		}

		//lsp.didChange(AbsoluteFilePath, version, startLine, startChar, endLine, endChar, edit.NewText)
	}

	//e.UpdateColors()
	e.UpdateNeeded()
}
