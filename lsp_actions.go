package main

import (
	"fmt"
	. "github.com/gdamore/tcell"
	"sort"
	"strings"
	"time"
)

func (e *Editor) onDefinition() {
	definition, err := lsp.definition(e.absoluteFilePath, e.r, e.c)

	if err != nil || len(definition.Result) == 0{
		return
	}

	if definition.Result[0].URI != "file://" + e.absoluteFilePath {
		e.inputFile = strings.Split(definition.Result[0].URI, "file://")[1]
		e.openFile(e.inputFile)
	}

	if int(definition.Result[0].Range.Start.Line) > len(e.content) ||  // not out of e.content
		int(definition.Result[0].Range.Start.Character) > len(e.content[int(definition.Result[0].Range.Start.Line)]) {
		return
	}

	e.r = int(definition.Result[0].Range.Start.Line)
	e.c = int(definition.Result[0].Range.Start.Character)
	e.selection.ssx = e.c; e.selection.ssy = e.r;
	e.selection.sey = int(definition.Result[0].Range.End.Line)
	e.selection.sex = int(definition.Result[0].Range.End.Character)
	e.r = e.selection.sey; e.c = e.selection.sex
	e.selection.isSelected = true
	e.focus()
}

func (e *Editor) onHover() {
	if !lsp.IsLangReady(e.lang) { return }

	e.isOverlay = true
	defer e.overlayFalse()

	var hoverEnd = false

	// loop until escape or enter pressed
	for !hoverEnd {

		start := time.Now()
		hover, err := lsp.hover(e.absoluteFilePath, e.r, e.c)
		elapsed := time.Since(start)

		lspStatus := "lsp hover, elapsed " + elapsed.String()
		status := fmt.Sprintf(" %s %s %d %d %s ", lspStatus, e.lang, e.r + 1, e.c + 1, e.inputFile)
		e.drawStatus(status)

		if err != nil || len(hover.Result.Contents.Value) == 0 { return }
		options := strings.Split(hover.Result.Contents.Value, "\n")
		if len(options) == 0 { return }

		tabs := countTabsTo(e.content[e.r], e.c)
		width := max(30, maxString(options))                                                                         // width depends on max option len or 30 at min
		height := minMany(10, len(options))                                                                          // depends on min option len or 5 at min or how many rows to the end of e.screen
		atx := (e.c - tabs) + e.LINES_WIDTH + tabs * (e.langTabWidth) + e.filesPanelWidth; aty := e.r - height - e.y // Define the window  position and dimensions
		style := StyleDefault.Foreground(ColorWhite)
		if len(options) > e.r - e.y { aty = e.r + 1 }

		var selectionEnd = false; var selected = 0; var selectedOffset = 0

		for !selectionEnd {
			if selected < selectedOffset { selectedOffset = selected }  // calculate offsets for scrolling completion
			if selected >= selectedOffset+height { selectedOffset = selected - height + 1 }

			e.drawCompletion(atx,aty, height, width, options, selected, selectedOffset, style)
			e.screen.Show()

			switch ev := e.screen.PollEvent().(type) { // poll and handle event
			case *EventKey:
				key := ev.Key()
				if key == KeyEscape || key == KeyEnter ||
					key == KeyBackspace || key == KeyBackspace2 {
					e.screen.Clear(); selectionEnd = true; hoverEnd = true
				}
				if key == KeyDown { selected = min(len(options)-1, selected+1) }
				if key == KeyUp { selected = max(0, selected-1) }
				if key == KeyRight { e.onRight(); e.screen.Clear(); e.drawEverything(); selectionEnd = true }
				if key == KeyLeft { e.onLeft(); e.screen.Clear(); e.drawEverything(); selectionEnd = true }
			}
		}
	}
}

func (e *Editor) onSignatureHelp() {
	if !lsp.IsLangReady(e.lang) { return }

	e.isOverlay = true
	defer e.overlayFalse()

	var end = false

	// loop until escape or enter pressed
	for !end {

		start := time.Now()
		signatureHelpResponse, err := lsp.signatureHelp(e.absoluteFilePath, e.r, e.c)
		elapsed := time.Since(start)

		lspStatus := "lsp signature help, elapsed " + elapsed.String()
		status := fmt.Sprintf(" %s %s %d %d %s ", lspStatus, e.lang, e.r+1, e.c + 1, e.inputFile)
		e.drawStatus(status)

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

		tabs := countTabsTo(e.content[e.r], e.c)
		width := max(30, maxString(options))                                                                       // width depends on max option len or 30 at min
		height := minMany(10, len(options))                                                                        // depends on min option len or 5 at min or how many rows to the end of e.screen
		atx := (e.c - tabs) + e.LINES_WIDTH + tabs*(e.langTabWidth) + e.filesPanelWidth; aty := e.r - height - e.y // Define the window  position and dimensions
		style := StyleDefault.Foreground(ColorWhite)
		if len(options) > e.r - e.y { aty = e.r + 1 }

		var selectionEnd = false; var selected = 0; var selectedOffset = 0

		for !selectionEnd {
			if selected < selectedOffset { selectedOffset = selected }  // calculate offsets for scrolling completion
			if selected >= selectedOffset+height { selectedOffset = selected - height + 1 }

			e.drawCompletion(atx,aty, height, width, options, selected, selectedOffset, style)
			e.screen.Show()

			switch ev := e.screen.PollEvent().(type) { // poll and handle event
			case *EventKey:
				key := ev.Key()
				if key == KeyEscape || key == KeyEnter ||
					key == KeyBackspace || key == KeyBackspace2 { e.screen.Clear(); selectionEnd = true; end = true }
				if key == KeyDown { selected = min(len(options)-1, selected+1) }
				if key == KeyUp { selected = max(0, selected-1) }
				if key == KeyRight { e.onRight(); e.screen.Clear(); e.drawEverything(); selectionEnd = true }
				if key == KeyLeft { e.onLeft(); e.screen.Clear(); e.drawEverything(); selectionEnd = true }
				if key == KeyRune { e.addChar(ev.Rune()); e.writeFile(); e.screen.Clear(); e.drawEverything(); selectionEnd = true  }
			}
		}
	}
}

func (e *Editor) onReferences() {
	if !lsp.IsLangReady(e.lang) { return }

	e.isOverlay = true
	defer e.overlayFalse()

	var end = false

	// loop until escape or enter pressed
	for !end {

		start := time.Now()
		referencesResponse, err := lsp.references(e.absoluteFilePath, e.r, e.c)
		elapsed := time.Since(start)

		lspStatus := "lsp references, elapsed " + elapsed.String()
		status := fmt.Sprintf(" %s %s %d %d %s ", lspStatus, e.lang, e.r + 1, e. c + 1, e.inputFile)
		e.drawStatus(status)

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

		tabs := countTabsTo(e.content[e.r], e.c)
		width := max(30, maxString(options))                                                                     // width depends on max option len or 30 at min
		height := minMany(10, len(options))                                                                      // depends on min option len or 5 at min or how many rows to the end of e.screen
		atx := (e.c-tabs) + e.LINES_WIDTH + tabs*(e.langTabWidth) + e.filesPanelWidth; aty := e.r - height - e.y // Define the window  position and dimensions
		style := StyleDefault.Foreground(ColorWhite)
		if len(options) > e.r - e.y { aty = e.r + 1 }

		var selectionEnd = false; var selected = 0; var selectedOffset = 0

		for !selectionEnd {
			if selected < selectedOffset { selectedOffset = selected }  // calculate offsets for scrolling completion
			if selected >= selectedOffset+height { selectedOffset = selected - height + 1 }

			e.drawCompletion(atx,aty, height, width, options, selected, selectedOffset, style)
			e.screen.Show()

			switch ev := e.screen.PollEvent().(type) { // poll and handle event
			case *EventKey:
				key := ev.Key()
				if key == KeyEscape || key == KeyEnter ||
					key == KeyBackspace || key == KeyBackspace2 { e.screen.Clear(); selectionEnd = true; end = true }
				if key == KeyDown { selected = min(len(options)-1, selected+1) }
				if key == KeyUp { selected = max(0, selected-1) }
				if key == KeyRight { e.onRight(); e.screen.Clear(); e.drawEverything(); selectionEnd = true }
				if key == KeyLeft { e.onLeft(); e.screen.Clear(); e.drawEverything(); selectionEnd = true }
				if key == KeyRune { e.addChar(ev.Rune()); e.writeFile(); e.screen.Clear(); e.drawEverything(); selectionEnd = true  }
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
	if referencesResult.URI != "file://"+ e.absoluteFilePath {  // if another file
		e.inputFile = strings.Split(referencesResult.URI, "file://")[1]
		e.openFile(e.inputFile)
	}

	e.r = referencesResult.Range.Start.Line
	e.c = referencesResult.Range.Start.Character
	e.selection.ssx = e.c; e.selection.ssy = e.r;
	e.selection.sey = referencesResult.Range.End.Line
	e.selection.sex = referencesResult.Range.End.Character
	e.selection.isSelected = true
	e.r = e.selection.sey; e.c = e.selection.sex
	e.focus();
	e.drawEverything()
}


func (e *Editor) onCompletion() {
	if !lsp.IsLangReady(e.lang) { return }
	e.isOverlay = true
	defer e.overlayFalse()

	var completionEnd = false

	// loop until escape or enter pressed
	for !completionEnd {

		start := time.Now()
		completion, err := lsp.completion(e.absoluteFilePath, e.r, e.c)
		elapsed := time.Since(start)

		lspStatus := "lsp completion, elapsed " + elapsed.String()
		status := fmt.Sprintf(" %s %s %d %d %s ", lspStatus, e.lang, e.r+1, e.c+1, e.filename)
		e.drawStatus(status)

		options := e.buildCompletionOptions(completion)
		if err != nil || len(options) == 0 { return }

		tabs := countTabsTo(e.content[e.r], e.c)
		atx := (e.c - tabs) + e.LINES_WIDTH + tabs*(e.langTabWidth) + e.filesPanelWidth; aty := e.r + 1 - e.y // Define the window  position and dimensions
		width := max(30, maxString(options))                                                                  // width depends on max option len or 30 at min
		height := minMany(5, len(options), e.ROWS - (e.r - e.y))                                              // depends on min option len or 5 at min or how many rows to the end of e.screen
		style := StyleDefault
		// if completion on last two rows of the e.screen - move window up
		if e.r - e.y  >= e.ROWS - 1 { aty -= min(5, len(options)); aty--; height = min(5, len(options)) }

		var selectionEnd = false; var selected = 0; var selectedOffset = 0


		for !selectionEnd {
			if selected < selectedOffset { selectedOffset = selected }  // calculate offsets for scrolling completion
			if selected >= selectedOffset+height { selectedOffset = selected - height + 1 }

			e.drawCompletion(atx,aty, height, width, options, selected, selectedOffset, style)
			e.screen.Show()

			switch ev := e.screen.PollEvent().(type) { // poll and handle event
			case *EventKey:
				key := ev.Key()
				if key == KeyEscape || key == KeyCtrlSpace { selectionEnd = true; completionEnd = true }
				if key == KeyDown { selected = min(len(options)-1, selected+1); e.screen.Clear(); e.drawEverything() }
				if key == KeyUp { selected = max(0, selected-1); e.screen.Clear(); e.drawEverything(); }
				if key == KeyRight { e.onRight(); e.screen.Clear(); e.drawEverything(); selectionEnd = true }
				if key == KeyLeft { e.onLeft(); e.screen.Clear(); e.drawEverything(); selectionEnd = true }
				if key == KeyRune { e.addChar(ev.Rune()); e.screen.Clear(); e.drawEverything(); selectionEnd = true  }
				if key == KeyBackspace || key == KeyBackspace2 {
					e.onDelete(); e.screen.Clear(); e.drawEverything(); selectionEnd = true
				}
				if key == KeyEnter {
					selectionEnd = true; completionEnd = true
					e.completionApply(completion, selected)
					e.screen.Show()
				}
			}
		}
	}
}


func (e *Editor) buildCompletionOptions(completion CompletionResponse) []string {
	var options []string
	var maxOptlen = 5

	prev := findPrevWord(e.content[e.r], e.c)
	filterword := string(e.content[e.r][prev:e.c])

	sortItemsByMatchCount(&completion.Result, filterword)

	for _, item := range completion.Result.Items {
		if len(item.Label) > maxOptlen { maxOptlen = len(item.Label) }
	}
	for _, item := range completion.Result.Items {
		options = append(options, formatText(item.Label, item.Detail, maxOptlen))
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


	// If match is close to the start of the string but not at the beginning, add some score.
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

		e.screen.SetContent(atx-1, row+aty, ' ', nil, style)
		for col, char := range option {
			e.screen.SetContent(col+atx, row+aty, char, nil, style)
		}
		for col := len(option); col < width; col++ { // Fill the remaining space
			e.screen.SetContent(col+atx, row+aty, ' ', nil, style)
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
		e.c = int(from)
		// remove chars between from and end
		e.content[e.r] = append(e.content[e.r][:e.c], e.content[e.r][int(end):]...)
		newText = item.TextEdit.NewText
	}

	if from == 0 && end == 0 {
		// text edit not supported by lsp
		prev := findPrevWord(e.content[e.r], e.c)
		next := findNextWord(e.content[e.r], e.c)
		from = float64(prev)
		newText = item.InsertText
		if len(newText) == 0 { newText = item.Label }
		end = float64(next)
		e.c = prev
		e.content[e.r] = append(e.content[e.r][:e.c], e.content[e.r][int(end) :]...)
	}

	// add newText
	for _, char := range newText { e.insertCharacter(e.r, e.c, char); e.c++ }
	e.updateColorsAtLine(e.r)

	e.update = true
	e.isContentChanged = true
	if len(e.content) <= 10000 { go e.writeFile() }
}

func (e *Editor) onRename() {
	var end = false
	var renameTo = []rune{}
	var patternx = 0
	var prefix = []rune("rename: ")

	prepareRenameResponse, err := lsp.prepareRename(e.absoluteFilePath, e.r, e.c)
	if err != nil  { return }
	placeHolder := prepareRenameResponse.Result.Placeholder
	renameTo = []rune(placeHolder)
	patternx = len(renameTo)

	// loop until escape or enter pressed
	for !end {

		for i := 0; i < len(prefix); i++ {
			e.screen.SetContent(e.filesPanelWidth + e.LINES_WIDTH+ i, e.ROWS-1, prefix[i], nil, StyleDefault)
		}

		e.screen.SetContent(e.filesPanelWidth + e.LINES_WIDTH+ len(prefix), e.ROWS-1, ' ', nil, StyleDefault)

		for i := 0; i < len(renameTo); i++ {
			e.screen.SetContent(e.filesPanelWidth + e.LINES_WIDTH+ len(prefix) + i, e.ROWS-1, renameTo[i], nil, StyleDefault)
		}

		e.screen.ShowCursor(e.filesPanelWidth + e.LINES_WIDTH+ len(prefix) + patternx , e.ROWS-1)

		for i := e.filesPanelWidth + e.LINES_WIDTH + len(prefix) + len(renameTo); i < e.COLUMNS; i++ {
			e.screen.SetContent(i, e.ROWS-1, ' ', nil, StyleDefault)
		}

		e.screen.Show()


		switch ev := e.screen.PollEvent().(type) { // poll and handle event
		case *EventResize:
			e.COLUMNS, e.ROWS = e.screen.Size()

		case *EventKey:
			key := ev.Key()

			if key == KeyRune {
				renameTo = insert(renameTo, patternx, ev.Rune())
				patternx++
			}
			if key == KeyBackspace2 && patternx > 0 && len(renameTo) > 0 {
				patternx--
				renameTo = remove(renameTo, patternx)
			}
			if key == KeyLeft  && patternx > 0 { patternx-- }
			if key == KeyRight && patternx < len(renameTo) { patternx++ }
			if key == KeyESC  || key == KeyCtrlF { end = true }
			if key == KeyEnter {
				renameResponse, err := lsp.rename(e.absoluteFilePath, string(renameTo), e.r, e.c)
				if err != nil  { return }
				e.applyRename(renameResponse)
				end = true
			}
		}
	}
}

func (e *Editor) applyRename(renameResponse RenameResponse) {
	inputFileTmp := e.inputFile
	tmpr := e.r; tmpc := e.c;  tmpy := e.y;  tmpx := e.x;

	for _, dc := range renameResponse.Result.DocumentChanges {
		if dc.TextDocument.URI != "file://" + e.absoluteFilePath {
			e.inputFile = strings.Split(dc.TextDocument.URI, "file://")[1]
			e.openFile(e.inputFile)
		}
		if dc.TextDocument.URI == "file://"+e.absoluteFilePath {
			// apply the changes in reverse order, its matter
			for ei := len(dc.Edits) - 1; ei >= 0; ei-- {
				edit := dc.Edits[ei]
				line := int(edit.Range.Start.Line)
				startc := int(edit.Range.Start.Character)
				endc := int(edit.Range.End.Character)

				// replace the old text with the new text
				after := e.content[line][endc:]
				newText := []rune(edit.NewText)
				newTextAndAfter := append(newText, after...)
				before := e.content[line][:startc]
				wholeNewLine := append(before, newTextAndAfter...)
				e.content[line] = wholeNewLine

				e.updateColorsAtLine(line)
			}

			e.writeFile()
		}
	}

	if e.inputFile != inputFileTmp {
		e.inputFile = inputFileTmp
		e.openFile(e.inputFile)
		e.r = tmpr; e.c = tmpc;  e.y = tmpy;  e.x = tmpx;
	}

}


func (e *Editor) onCodeAction() {
	codeAction, err := lsp.codeAction(e.absoluteFilePath, e.selection.ssx, e.selection.ssy, e.selection.sex, e.selection.sey)
	if err != nil { return }
	if len(codeAction.Result) == 0 { return }
	//
	commandResponse, err := lsp.command(codeAction.Result[0].Command)
	if err != nil { return }
	if len(commandResponse.Params.Edit.DocumentChanges) == 0 { return }

	e.handleEdits(commandResponse.Params.Edit.DocumentChanges[0].Edits, commandResponse.Params.Edit.DocumentChanges[0].TextDocument.Version+1)
	lsp.applyEdit(commandResponse.ID)
}

func (e *Editor) handleEdits(edits []Edit, version int) {
	for _, edit := range edits {
		start := edit.Range.Start
		//end := edit.Range.End

		// Adjusting because slices are 0-indexed.
		//startLine, startChar := int(start.Line), int(start.Character)
		//endLine, endChar := int(end.Line), int(end.Character)

		newLines := strings.Split(edit.NewText, "\n")
		for i, line := range newLines {
			index := int(start.Line) + i
			if index >= len(e.content) {
				e.content = append(e.content,  []rune(line))
			} else  {
				e.content[index] = []rune(line)
			}
		}

		//lsp.didChange(absoluteFilePath, version, startLine, startChar, endLine, endChar, edit.NewText)
	}

	//e.updateColors()
	e.updateNeeded()
}
