package main

import (
	"fmt"
	. "github.com/gdamore/tcell"
	"sort"
	"strings"
	"time"
)

func (e *Editor) onDefinition() {
	definition, err := lsp.definition(e.absoluteFilePath, r, c )

	if err != nil || len(definition.Result) == 0{
		return
	}

	if definition.Result[0].URI != "file://" + e.absoluteFilePath {
		e.inputFile = strings.Split(definition.Result[0].URI, "file://")[1]
		e.openFile(e.inputFile)
	}

	if int(definition.Result[0].Range.Start.Line) > len(content) ||  // not out of content
		int(definition.Result[0].Range.Start.Character) > len(content[int(definition.Result[0].Range.Start.Line)]) {
		return
	}

	r = int(definition.Result[0].Range.Start.Line)
	c = int(definition.Result[0].Range.Start.Character)
	e.selection.ssx = c; e.selection.ssy = r;
	e.selection.sey = int(definition.Result[0].Range.End.Line)
	e.selection.sex = int(definition.Result[0].Range.End.Character)
	r = e.selection.sey; c = e.selection.sex
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
		hover, err := lsp.hover(e.absoluteFilePath, r, c)
		elapsed := time.Since(start)

		lspStatus := "lsp hover, elapsed " + elapsed.String()
		status := fmt.Sprintf(" %s %s %d %d %s ", lspStatus, e.lang, r+1, c+1, e.inputFile)
		e.drawStatus(status)

		if err != nil || len(hover.Result.Contents.Value) == 0 { return }
		options := strings.Split(hover.Result.Contents.Value, "\n")
		if len(options) == 0 { return }

		tabs := countTabsTo(content[r], c)
		width := max(30, maxString(options)) // width depends on max option len or 30 at min
		height := minMany(10, len(options)) // depends on min option len or 5 at min or how many rows to the end of screen
		atx := (c-tabs) + e.LS + tabs*(e.tabWidth) + e.filesPanelWidth; aty := r - height - y // Define the window  position and dimensions
		style := StyleDefault.Foreground(ColorWhite)
		if len(options) > r - y { aty = r + 1 }

		var selectionEnd = false; var selected = 0; var selectedOffset = 0

		for !selectionEnd {
			if selected < selectedOffset { selectedOffset = selected }  // calculate offsets for scrolling completion
			if selected >= selectedOffset+height { selectedOffset = selected - height + 1 }

			e.drawCompletion(atx,aty, height, width, options, selected, selectedOffset, style)
			s.Show()

			switch ev := s.PollEvent().(type) { // poll and handle event
			case *EventKey:
				key := ev.Key()
				if key == KeyEscape || key == KeyEnter ||
					key == KeyBackspace || key == KeyBackspace2 {
					s.Clear(); selectionEnd = true; hoverEnd = true
				}
				if key == KeyDown { selected = min(len(options)-1, selected+1) }
				if key == KeyUp { selected = max(0, selected-1) }
				if key == KeyRight { e.onRight(); s.Clear(); e.drawEverything(); selectionEnd = true }
				if key == KeyLeft { e.onLeft(); s.Clear(); e.drawEverything(); selectionEnd = true }
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
		signatureHelpResponse, err := lsp.signatureHelp(e.absoluteFilePath, r, c)
		elapsed := time.Since(start)

		lspStatus := "lsp signature help, elapsed " + elapsed.String()
		status := fmt.Sprintf(" %s %s %d %d %s ", lspStatus, e.lang, r+1, c+1, e.inputFile)
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

		tabs := countTabsTo(content[r], c)
		width := max(30, maxString(options)) // width depends on max option len or 30 at min
		height := minMany(10, len(options)) // depends on min option len or 5 at min or how many rows to the end of screen
		atx := (c-tabs) + e.LS + tabs*(e.tabWidth) + e.filesPanelWidth; aty := r - height - y // Define the window  position and dimensions
		style := StyleDefault.Foreground(ColorWhite)
		if len(options) > r - y { aty = r + 1 }

		var selectionEnd = false; var selected = 0; var selectedOffset = 0

		for !selectionEnd {
			if selected < selectedOffset { selectedOffset = selected }  // calculate offsets for scrolling completion
			if selected >= selectedOffset+height { selectedOffset = selected - height + 1 }

			e.drawCompletion(atx,aty, height, width, options, selected, selectedOffset, style)
			s.Show()

			switch ev := s.PollEvent().(type) { // poll and handle event
			case *EventKey:
				key := ev.Key()
				if key == KeyEscape || key == KeyEnter ||
					key == KeyBackspace || key == KeyBackspace2 { s.Clear(); selectionEnd = true; end = true }
				if key == KeyDown { selected = min(len(options)-1, selected+1) }
				if key == KeyUp { selected = max(0, selected-1) }
				if key == KeyRight { e.onRight(); s.Clear(); e.drawEverything(); selectionEnd = true }
				if key == KeyLeft { e.onLeft(); s.Clear(); e.drawEverything(); selectionEnd = true }
				if key == KeyRune { e.addChar(ev.Rune()); e.writeFile(); s.Clear(); e.drawEverything(); selectionEnd = true  }
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
		referencesResponse, err := lsp.references(e.absoluteFilePath, r, c)
		elapsed := time.Since(start)

		lspStatus := "lsp references, elapsed " + elapsed.String()
		status := fmt.Sprintf(" %s %s %d %d %s ", lspStatus, e.lang, r+1, c+1, e.inputFile)
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

		tabs := countTabsTo(content[r], c)
		width := max(30, maxString(options)) // width depends on max option len or 30 at min
		height := minMany(10, len(options)) // depends on min option len or 5 at min or how many rows to the end of screen
		atx := (c-tabs) + e.LS + tabs*(e.tabWidth) + e.filesPanelWidth; aty := r - height - y // Define the window  position and dimensions
		style := StyleDefault.Foreground(ColorWhite)
		if len(options) > r - y { aty = r + 1 }

		var selectionEnd = false; var selected = 0; var selectedOffset = 0

		for !selectionEnd {
			if selected < selectedOffset { selectedOffset = selected }  // calculate offsets for scrolling completion
			if selected >= selectedOffset+height { selectedOffset = selected - height + 1 }

			e.drawCompletion(atx,aty, height, width, options, selected, selectedOffset, style)
			s.Show()

			switch ev := s.PollEvent().(type) { // poll and handle event
			case *EventKey:
				key := ev.Key()
				if key == KeyEscape || key == KeyEnter ||
					key == KeyBackspace || key == KeyBackspace2 { s.Clear(); selectionEnd = true; end = true }
				if key == KeyDown { selected = min(len(options)-1, selected+1) }
				if key == KeyUp { selected = max(0, selected-1) }
				if key == KeyRight { e.onRight(); s.Clear(); e.drawEverything(); selectionEnd = true }
				if key == KeyLeft { e.onLeft(); s.Clear(); e.drawEverything(); selectionEnd = true }
				if key == KeyRune { e.addChar(ev.Rune()); e.writeFile(); s.Clear(); e.drawEverything(); selectionEnd = true  }
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

	r = referencesResult.Range.Start.Line
	c = referencesResult.Range.Start.Character
	e.selection.ssx = c; e.selection.ssy = r;
	e.selection.sey = referencesResult.Range.End.Line
	e.selection.sex = referencesResult.Range.End.Character
	e.selection.isSelected = true
	r = e.selection.sey; c = e.selection.sex
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
		completion, err := lsp.completion(e.absoluteFilePath, r, c)
		elapsed := time.Since(start)

		lspStatus := "lsp completion, elapsed " + elapsed.String()
		status := fmt.Sprintf(" %s %s %d %d %s ", lspStatus, e.lang, r+1, c+1, e.filename)
		e.drawStatus(status)

		options := e.buildCompletionOptions(completion)
		if err != nil || len(options) == 0 { return }

		tabs := countTabsTo(content[r], c)
		atx := (c-tabs) + e.LS + tabs*(e.tabWidth) + e.filesPanelWidth; aty := r + 1 - y // Define the window  position and dimensions
		width := max(30, maxString(options)) // width depends on max option len or 30 at min
		height := minMany(5, len(options), e.ROWS - (r - y)) // depends on min option len or 5 at min or how many rows to the end of screen
		style := StyleDefault
		// if completion on last two rows of the screen - move window up
		if r - y  >= e.ROWS - 1 { aty -= min(5, len(options)); aty--; height = min(5, len(options)) }

		var selectionEnd = false; var selected = 0; var selectedOffset = 0


		for !selectionEnd {
			if selected < selectedOffset { selectedOffset = selected }  // calculate offsets for scrolling completion
			if selected >= selectedOffset+height { selectedOffset = selected - height + 1 }

			e.drawCompletion(atx,aty, height, width, options, selected, selectedOffset, style)
			s.Show()

			switch ev := s.PollEvent().(type) { // poll and handle event
			case *EventKey:
				key := ev.Key()
				if key == KeyEscape || key == KeyCtrlSpace { selectionEnd = true; completionEnd = true }
				if key == KeyDown { selected = min(len(options)-1, selected+1); s.Clear(); e.drawEverything() }
				if key == KeyUp { selected = max(0, selected-1); s.Clear(); e.drawEverything(); }
				if key == KeyRight { e.onRight(); s.Clear(); e.drawEverything(); selectionEnd = true }
				if key == KeyLeft { e.onLeft(); s.Clear(); e.drawEverything(); selectionEnd = true }
				if key == KeyRune { e.addChar(ev.Rune()); s.Clear(); e.drawEverything(); selectionEnd = true  }
				if key == KeyBackspace || key == KeyBackspace2 {
					e.onDelete(); s.Clear(); e.drawEverything(); selectionEnd = true
				}
				if key == KeyEnter {
					selectionEnd = true; completionEnd = true
					e.completionApply(completion, selected)
					s.Show()
				}
			}
		}
	}
}


func (e *Editor) buildCompletionOptions(completion CompletionResponse) []string {
	var options []string
	var maxOptlen = 5

	prev := findPrevWord(content[r], c)
	filterword := string(content[r][prev:c])

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

		s.SetContent(atx-1, row+aty, ' ', nil, style)
		for col, char := range option {
			s.SetContent(col+atx, row+aty, char, nil, style)
		}
		for col := len(option); col < width; col++ { // Fill the remaining space
			s.SetContent(col+atx, row+aty, ' ', nil, style)
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
		c = int(from)
		// remove chars between from and end
		content[r] = append(content[r][:c], content[r][int(end):]...)
		newText = item.TextEdit.NewText
	}

	if from == 0 && end == 0 {
		// text edit not supported by lsp
		prev := findPrevWord(content[r], c)
		next := findNextWord(content[r], c)
		from = float64(prev)
		newText = item.InsertText
		if len(newText) == 0 { newText = item.Label }
		end = float64(next)
		c = prev
		content[r] = append(content[r][:c], content[r][int(end) :]...)
	}

	// add newText
	for _, char := range newText { e.insertCharacter(r,c,char); c++ }
	e.updateColorsAtLine(r)

	e.update = true
	e.isContentChanged = true
	if len(content) <= 10000 { go e.writeFile() }
}

func (e *Editor) onRename() {
	var end = false
	var renameTo = []rune{}
	var patternx = 0
	var prefix = []rune("rename: ")

	prepareRenameResponse, err := lsp.prepareRename(e.absoluteFilePath, r, c)
	if err != nil  { return }
	placeHolder := prepareRenameResponse.Result.Placeholder
	renameTo = []rune(placeHolder)
	patternx = len(renameTo)

	// loop until escape or enter pressed
	for !end {

		for i := 0; i < len(prefix); i++ {
			s.SetContent(e.filesPanelWidth + e.LS + i, e.ROWS-1, prefix[i], nil, StyleDefault)
		}

		s.SetContent(e.filesPanelWidth + e.LS + len(prefix), e.ROWS-1, ' ', nil, StyleDefault)

		for i := 0; i < len(renameTo); i++ {
			s.SetContent(e.filesPanelWidth + e.LS + len(prefix) + i, e.ROWS-1, renameTo[i], nil, StyleDefault)
		}

		s.ShowCursor(e.filesPanelWidth + e.LS + len(prefix) + patternx , e.ROWS-1)

		for i := e.filesPanelWidth + e.LS + len(prefix) + len(renameTo); i < e.COLUMNS; i++ {
			s.SetContent(i, e.ROWS-1, ' ', nil, StyleDefault)
		}

		s.Show()


		switch ev := s.PollEvent().(type) { // poll and handle event
		case *EventResize:
			e.COLUMNS, e.ROWS = s.Size()

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
				renameResponse, err := lsp.rename(e.absoluteFilePath, string(renameTo), r, c)
				if err != nil  { return }
				e.applyRename(renameResponse)
				end = true
			}
		}
	}
}

func (e *Editor) applyRename(renameResponse RenameResponse) {
	inputFileTmp := e.inputFile
	tmpr := r; tmpc := c;  tmpy := y;  tmpx := x;

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
				after := content[line][endc:]
				newText := []rune(edit.NewText)
				newTextAndAfter := append(newText, after...)
				before := content[line][:startc]
				wholeNewLine := append(before, newTextAndAfter...)
				content[line] = wholeNewLine

				e.updateColorsAtLine(line)
			}

			e.writeFile()
		}
	}

	if e.inputFile != inputFileTmp {
		e.inputFile = inputFileTmp
		e.openFile(e.inputFile)
		r = tmpr; c = tmpc;  y = tmpy;  x = tmpx;
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
			if index >= len(content) {
				content = append(content,  []rune(line))
			} else  {
				content[index] = []rune(line)
			}
		}

		//lsp.didChange(absoluteFilePath, version, startLine, startChar, endLine, endChar, edit.NewText)
	}

	//e.updateColors()
	e.updateNeeded()
}
