package main

import (
	"fmt"
	"github.com/atotto/clipboard"
	. "github.com/gdamore/tcell"
	"github.com/gdamore/tcell/encoding"
	"math"
	"os"
	"path"
	"path/filepath"
	"strings"
)

type Editor struct {
	COLUMNS     int // terminal size columns
	ROWS        int // terminal size rows
	LINES_WIDTH int // draw file lines number

	r int // cursor position row
	c int // cursor position column
	y int // row offset for scrolling
	x int // col offset for scrolling

	content [][]rune // text characters
	colors  [][]int  // text characters colors

	screen Screen // screen for drawing

	lang         string // current file language
	config       Config // config, lsp, tabs, comments, etc
	langConf     Lang   // current lang conf
	langTabWidth int    // current lang tabs indentation  '\t' -> "    "

	selection Selection // selection

	undo []EditOperation // stack for undo operations
	redo []EditOperation // stack for redo operations

	inputFile        string // exact user input
	filename         string // current file name
	absoluteFilePath string // current file name and directory
	isContentChanged bool   // shows * if file is changed
	isColorize       bool   // colorize text is true by default
	update           bool   // for screen updates,  if false it will not draw
	isOverlay        bool   // true if overlay is active (completion, hover, errors...)

	filesPanelWidth     int        // current width for files panel
	files               []FileInfo // current dir files
	isFileSelection     bool       // true if in files selection menu
	fileScrollingOffset int        // for vertical scrolling  in selection menu
	fileSelectedIndex   int        // selected file index
	isFilesSearch       bool       // true if in files search mode
	isFilesPanelMoving  bool       // true if in files panel moving mode

	searchPattern []rune // pattern for search in a buffer

	//filesInfo []FileInfo

}

func (e *Editor) start() {
	logger.info("starting edgo")

	e.initScreen()

	// reading file from cmd args
	if len(os.Args) == 1 {
		e.onFiles()
	} else {
		e.filename = os.Args[1]
		e.inputFile = e.filename
		err := e.openFile(e.filename)
		if err != nil {
			fmt.Println(err)
			os.Exit(130)
		}
	}

	// main draw cycle
	for {
		if e.update && e.filename != "" {
			e.drawEverything()
			e.screen.Show()
		}
		e.handleEvents()
	}
}

func (e *Editor) handleEvents() {
	e.update = true
	ev := e.screen.PollEvent()
	switch ev := ev.(type) {
	case *EventResize:
		e.COLUMNS, e.ROWS = e.screen.Size()

	case *EventMouse:
		mx, my := ev.Position()
		buttons := ev.Buttons()
		modifiers := ev.Modifiers()

		e.handleMouse(mx, my, buttons, modifiers)

	case *EventKey:
		key := ev.Key()
		modifiers := ev.Modifiers()

		e.handleKeyboard(key, ev, modifiers)
	}
}

func (e *Editor) handleMouse(mx int, my int, buttons ButtonMask, modifiers ModMask) {
	if !e.isFilesPanelMoving && buttons & Button1 == 1 &&
		math.Abs(float64(e.filesPanelWidth - mx)) <= 2 &&
		len(e.selection.getSelectedLines(e.content)) == 0 {

		e.filesPanelWidth = mx - 1
		e.isFilesPanelMoving = true
		return
	}

	if e.isFilesPanelMoving && buttons & Button1 == 1 { e.filesPanelWidth = mx; return }
	if e.isFilesPanelMoving && buttons & Button1 == 0 { e.isFilesPanelMoving = false; return }
	if mx < e.filesPanelWidth - 2 && buttons & Button1 == 0 { e.onFiles(); return }

	if e.filename == "" { return }

	mx -= e.LINES_WIDTH + e.filesPanelWidth

	if mx < 0 { return }
	if my > e.ROWS { return }

	// if click with control or option, lookup for definition or references
	if buttons & Button1 == 1 && (modifiers & ModAlt != 0 || modifiers & ModCtrl != 0) {
		e.r = my + e.y
		if e.r > len(e.content)-1 { e.r = len(e.content) - 1 } // fit cursor to e.content

		e.c = e.findCursorXPosition(mx)
		if modifiers & ModAlt != 0 { e.onReferences() }
		if modifiers & ModCtrl != 0 { e.onDefinition() }
		return
	}

	if e.selection.isSelected && buttons & Button1 == 1 {
		e.r = my + e.y
		if e.r > len(e.content)-1 { e.r = len(e.content) - 1 } // fit cursor to e.content

		xPosition := e.findCursorXPosition(mx)

		isTripleClick := e.selection.isUnderSelection(xPosition, e.r) &&
			len(e.selection.getSelectedLines(e.content)) == 1

		if isTripleClick {
			e.r = my + e.y
			e.c = xPosition
			if e.r > len(e.content)-1 { e.r = len(e.content) - 1 } // fit cursor to e.content
			if e.c > len(e.content[e.r]) { e.c = len(e.content[e.r]) }
			//if e.c < 0 { sex = len(e.content[r]) }

			e.selection.ssx = 0
			e.selection.sex = len(e.content[e.r])
			e.selection.ssy = e.r
			e.selection.sey = e.r

			return
		} else {
			e.selection.cleanSelection()
		}
	}

	if buttons & WheelDown != 0 { e.onScrollDown(); return }
	if buttons & WheelUp != 0 { e.onScrollUp(); return }
	if buttons & Button1 == 0 && e.selection.ssx == -1 { e.update = false; return }

	if buttons & Button1 == 1 {
		e.r = my + e.y
		if e.r > len(e.content)-1 { e.r = len(e.content) - 1 } // fit cursor to e.content

		xPosition := e.findCursorXPosition(mx)

		if e.c == xPosition && len(e.selection.getSelectedLines(e.content)) == 0 {
			// double click
			prw := findPrevWord(e.content[e.r], e.c)
			nxw := findNextWord(e.content[e.r], e.c)
			e.selection.ssx, e.selection.ssy = prw, e.r
			e.selection.sex, e.selection.sey = nxw, e.r
			e.c = nxw
			return
		}
		e.c = xPosition
		if e.c < 0 { e.c = 0 }
		if e.selection.ssx < 0 { e.selection.ssx, e.selection.ssy = e.c, e.r }
		if e.selection.ssx >= 0 { e.selection.sex, e.selection.sey = e.c, e.r }
	}

	if buttons&Button1 == 0 {
		if e.selection.ssx != -1 && e.selection.sex != -1 {
			e.selection.isSelected = true
		}
	}
	return
}
func (e *Editor) handleKeyboard(key Key, ev *EventKey,  modifiers ModMask) {
	if e.filename == "" && key != KeyCtrlQ { return }

	if ev.Rune() == '/' && modifiers&ModAlt != 0 || int(ev.Rune()) == '÷' {
		// '÷' on Mac is option + '/'
		e.onCommentLine(); return
	}
	if key == KeyUp && modifiers == 3 { e.onSwapLinesUp(); return } // control + shift + up
	if key == KeyDown && modifiers == 3 { e.onSwapLinesDown(); return } // control + shift + down
	if key == KeyBacktab { e.onBackTab(); return }
	if key == KeyTab { e.onTab(); return }
	if key == KeyCtrlH { e.onHover(); return }
	if key == KeyCtrlR { e.onReferences(); return }
	if key == KeyCtrlW { e.onCodeAction(); return }
	if key == KeyCtrlP { e.onSignatureHelp(); return }
	if key == KeyCtrlG { e.onDefinition(); return }
	if key == KeyCtrlE { e.onErrors(); return }
	if key == KeyCtrlC { e.onCopy(); return }
	if key == KeyCtrlV { e.onPaste(); return }
	if key == KeyEscape { e.selection.cleanSelection(); return }
	if key == KeyCtrlA { e.onSelectAll(); return }
	if key == KeyCtrlX { e.cut() }
	if key == KeyCtrlD { e.duplicate() }

	if modifiers & ModShift != 0 && (
		key == KeyRight ||
			key == KeyLeft ||
			key == KeyUp ||
			key == KeyDown) {

		if e.selection.ssx < 0 { e.selection.ssx, e.selection.ssy = e.c, e.r }
		if key == KeyRight { e.onRight() }
		if key == KeyLeft { e.onLeft() }
		if key == KeyUp { e.onUp() }
		if key == KeyDown { e.onDown() }
		if e.selection.ssx >= 0 {
			e.selection.sex, e.selection.sey = e.c, e.r
			e.selection.isSelected = true
		}
		return
	}

	if key == KeyRune && modifiers & ModAlt != 0 && len(e.content) > 0 { e.handleSmartMove(ev.Rune()); return }
	if key == KeyDown && modifiers & ModAlt != 0 { e.handleSmartMoveDown(); return }
	if key == KeyUp && modifiers & ModAlt != 0 { e.handleSmartMoveUp(); return }

	if key == KeyRune {
		e.addChar(ev.Rune())
		if ev.Rune() == '.' {
			e.drawEverything(); e.screen.Show()
			e.onCompletion()
		}
		//if ev.Rune() == '(' { e.drawEverything(); e.screen.Show(); e.onSignatureHelp(); e.screen.Clear() }
	}

	if /*key == tcell.KeyEscape ||*/ key == KeyCtrlQ { e.screen.Fini(); os.Exit(1) }
	if key == KeyCtrlS { e.writeFile() }
	if key == KeyEnter { e.onEnter() }
	if key == KeyBackspace || key == KeyBackspace2 { e.onDelete() }
	if key == KeyDown { e.onDown(); e.selection.cleanSelection() }
	if key == KeyUp { e.onUp(); e.selection.cleanSelection() }
	if key == KeyLeft { e.onLeft(); e.selection.cleanSelection() }
	if key == KeyRight { e.onRight(); e.selection.cleanSelection() }
	if key == KeyCtrlT { e.onFiles() }
	if key == KeyCtrlF { e.onSearch() }
	if key == KeyF18 { e.onRename() }
	if key == KeyCtrlU { e.onUndo() }
	if key == KeyCtrlR { e.onRedo() } // todo: fix it
	if key == KeyCtrlSpace { e.onCompletion() }

}
func (e *Editor) openFile(fname string) error {

	absoluteDir, err := filepath.Abs(path.Dir(fname))
	if err != nil { return err }
	//directory := absoluteDir;
	e.filename = filepath.Base(fname)
	e.absoluteFilePath = path.Join(absoluteDir, e.filename)

	logger.info("open", e.absoluteFilePath)

	newLang := detectLang(e.filename)
	logger.info("new lang is", newLang)

	if newLang != "" && newLang != e.lang {
		e.lang = newLang
		lsp.lang = newLang
		_, found := lsp.lang2stdin[e.lang]
		if !found { go e.init_lsp(e.lang) }
	}

	conf, found := e.config.Langs[e.lang]
	if !found { conf = DefaultLangConfig }
	e.langConf = conf
	e.langTabWidth = conf.TabWidth

	code := e.readFile(e.absoluteFilePath)
	e.colors = highlighter.colorize(code, e.filename)

	e.undo = []EditOperation{}
	e.redo = []EditOperation{}

	e.updateFilesOpenStats(fname)

    e.r = 0; e.c = 0; e.y = 0; e.x = 0
	e.selection = Selection{ -1,-1,-1,-1,false }

	return nil
}

func (e *Editor) initScreen() {
	encoding.Register()
	screen, err := NewScreen()
	if err != nil { fmt.Fprintf(os.Stderr, "%v\n", err); os.Exit(1) }
	e.screen = screen

	err2 := e.screen.Init()
	if err2 != nil { fmt.Fprintf(os.Stderr, "%v\n", err2); os.Exit(1) }

	e.screen.EnableMouse()
	e.screen.Clear()

	e.COLUMNS, e.ROWS = e.screen.Size()
	//ROWS -= 1
	
	e.LINES_WIDTH = 6
	e.update = true
	e.isColorize = true
	e.fileSelectedIndex = -1

	return
}

func (e *Editor) drawEverything() {
	if len(e.content) == 0 { return }
	e.screen.Clear()

	if e.filesPanelWidth != 0 { e.drawFiles("", e.files, 0) }

	
	//tabs := countTabsTo(e.content[r], e.c)
	//correction := tabs*(e.langTabWidth)

	if e.c < e.x { e.x = e.c }
	if e.c + e.LINES_WIDTH+ e.filesPanelWidth >= e.x + e.COLUMNS  {
		e.x = e.c - e.COLUMNS + 1 + e.LINES_WIDTH + e.filesPanelWidth
	}

	// draw line number and chars according to scrolling offsets
	for row := 0; row < e.ROWS; row++ {
		ry := row + e.y  // index to get right row in characters buffer by scrolling offset y
		//e.cleanLineAfter(0, row)
		if row >= len(e.content) || ry >= len(e.content) { break }
		e.drawLineNumber(ry, row)

		tabsOffset := 0
		for col := 0; col <= e.COLUMNS; col++ {
			cx := col + e.x // index to get right column in characters buffer by scrolling offset x
			if cx >= len(e.content[ry]) { break }
			ch := e.content[ry][cx]
			style := e.getStyle(ry, cx)
			if ch == '\t'  {
				//draw big cursor with next symbol color
				if ry == e.r && cx == e.c {
					var color = Color(AccentColor)
					if cx+1 < len(e.colors[ry]) { color = Color(e.colors[ry][cx+1]) }
					if color == -1 { color = Color(AccentColor)}
					style = StyleDefault.Background(color)
				}
				for i := 0; i < e.langTabWidth; i++ {
					e.screen.SetContent(col + e.LINES_WIDTH + tabsOffset + e.filesPanelWidth, row, ' ', nil, style)
					if i != e.langTabWidth-1 { tabsOffset++ }
				}
			} else {
				e.screen.SetContent(col + e.LINES_WIDTH+ tabsOffset + e.filesPanelWidth, row , ch, nil, style)
			}
		}
	}

	e.drawDiagnostic()
	//e.drawTabs()

	var changes = ""; if e.isContentChanged { changes = "*" }
	status := fmt.Sprintf(" %s %d %d %s%s ", e.lang, e.r+1, e.c+1, e.filename, changes)
	e.drawStatus(status)

	// if tab under cursor, hide cursor because it has already drawn
	if e.r < len(e.content) && e.c < len(e.content[e.r]) && e.content[e.r][e.c] == '\t' {
		e.screen.HideCursor()
	} else  {
		tabs := countTabsTo(e.content[e.r], e.c) * (e.langTabWidth -1)
		e.screen.ShowCursor(e.c - e.x + e.LINES_WIDTH+tabs + e.filesPanelWidth, e.r - e.y) // show cursor
	}

}

func (e *Editor) getStyle(ry int, cx int) Style {
	var style = StyleDefault
	if !e.isColorize { return style }
	if ry >= len(e.colors) || cx >= len(e.colors[ry])  { return style }
	color := e.colors[ry][cx]
	if color > 0 { style = StyleDefault.Foreground(Color(color)) }
	if e.selection.isUnderSelection(cx, ry) {
		style = style.Background(Color(SelectionColor))
	}
	return style
}

func (e *Editor) drawDiagnostic() {
	//lsp.someMapMutex2.Lock()
	maybeDiagnostic, found := lsp.file2diagnostic["file://" + e.absoluteFilePath]
	//lsp.someMapMutex2.Unlock()

	if found {
		//style := tcell.StyleDefault.Background(tcell.ColorIndianRed).Foreground(tcell.ColorWhite)
		style := StyleDefault.Foreground(Color(AccentColor))
		//textStyle := tcell.StyleDefault.Foreground(tcell.ColorIndianRed)

		for _, diagnostic := range maybeDiagnostic.Diagnostics {
			dline := int(diagnostic.Range.Start.Line)
			if dline >= len(e.content) { continue } // sometimes it out of e.content
			if dline - e.y > e.ROWS { continue } // sometimes it out of e.content

			// iterate over error range and, todo::fix
			//for i := dline; i <= int(diagnostic.Range.End.Line); i++ {
			//	if i >= len(e.content) { continue }
			//	tabs := countTabs(e.content[i], dline)
			//	for j := int(diagnostic.Range.Start.Character); j <= int(diagnostic.Range.End.Character); j++ {
			//		if j >= len(e.content[i]) { continue }
			//
			//		ch := e.content[dline][j]
			//		e.screen.SetContent(j+LINES_WIDTH + tabs*e.langTabWidth + x, i-y, ch, nil, textStyle)
			//	}
			//}


			tabs := countTabs(e.content[dline], len(e.content[dline]))
			var shifty = 0
			errorMessage := "error: " + diagnostic.Message
			errorMessage = PadLeft(errorMessage, e.COLUMNS - len(e.content[dline]) - tabs*e.langTabWidth- 5 - e.LINES_WIDTH- e.filesPanelWidth)

			// iterate over message characters and draw it
			for i, m := range errorMessage {
				ypos :=  dline - e.y
				if ypos < 0 || ypos >= len(e.content) { break }

				tabs = countTabs(e.content[dline], len(e.content[dline]))
				xpos := i + e.LINES_WIDTH + e.filesPanelWidth + len(e.content[dline+shifty]) + tabs*e.langTabWidth + 5

				//for { // draw ch on the next line if not fit to e.screen
				//	if xpos >= COLUMNS {
				//		shifty++
				//		tabs = countTabs(e.content[dline+shifty], len(e.content[dline+shifty]))
				//		ypos +=  (i / COLUMNS) + 1
				//		if ypos >= len(e.content) { break}
				//		xpos = len(e.content[dline+shifty]) + 5 + (xpos % COLUMNS) % COLUMNS
				//	} else { break }
				//}

				e.screen.SetContent(xpos,  ypos, m, nil,  style)
			}
		}

	}
}

func (e *Editor) drawLineNumber(brw int, row int) {
	var style = StyleDefault.Foreground(ColorDimGray)
	if brw == e.r { style = StyleDefault}
	lineNumber := centerNumber(brw + 1, e.LINES_WIDTH)
	for index, char := range lineNumber {
		e.screen.SetContent(index + e.filesPanelWidth, row, char, nil, style)
	}
}

func (e *Editor) drawStatus(text string) {
	var style = StyleDefault
	e.drawText(e.ROWS-1, e.COLUMNS - len(text), text, style)
}

func (e *Editor) drawText(row, col int, text string, style Style) {
	e.screen.SetContent(col-1, row, ' ', nil, style)
	for _, ch := range []rune(text) {
		if col > e.COLUMNS { break }
		e.screen.SetContent(col, row, ch, nil, style)
		col++
	}
}

func (e *Editor) findCursorXPosition(mx int) int {
	count := 0; realCount := 0  // searching x position
	for _, ch := range e.content[e.r] {
		if count >= mx + e.x { break }
		if ch == '\t' {
			count += e.langTabWidth; realCount++
		} else {
			count++; realCount++
		}
	}
	return realCount
}

func (e *Editor) init_lsp(lang string) {
	//start := time.Now()

	// Getting the lsp command with args for a language:
	conf, ok := e.config.Langs[strings.ToLower(lang)]
	if !ok || len(conf.Lsp) == 0 { return }  // lang is not supported.

	started := lsp.start(lang, strings.Split(conf.Lsp, " "))
	if !started { return }

	var diagnosticUpdateChan = make(chan string)
	go lsp.receiveLoop(diagnosticUpdateChan, lang)

	currentDir, _ := os.Getwd()

	lsp.init(currentDir)
	lsp.didOpen(e.absoluteFilePath, lang)

	//e.drawEverything()
	//
	//lspStatus := "lsp started, elapsed " + time.Since(start).String()
	//if !lsp.isReady { lspStatus = "lsp is not ready yet" }
	//logger.info("lsp status", lspStatus)
	//status := fmt.Sprintf(" %e.screen %e.screen %d %d %e.screen ", lspStatus,  lang, r+1, c+1, inputFile)
	//e.drawText(COLUMNS- len(status), ROWS-1, COLUMNS, ROWS-1, status)
	//e.screen.Show()

	go func() {
		for range diagnosticUpdateChan {
			if e.isOverlay { continue }
			e.drawEverything()
			e.screen.Show()
		}
	}()
}

func (e *Editor) onErrors() {
	if !lsp.IsLangReady(e.lang) { return }

	maybeDiagnostics, found := lsp.file2diagnostic["file://" + e.absoluteFilePath]

	if !found || len(maybeDiagnostics.Diagnostics) == 0 { return }

	e.isOverlay = true
	defer e.overlayFalse()

	var end = false

	// loop until escape or enter pressed
	for !end {

		var options = []string{}
		for _, diagnistic := range maybeDiagnostics.Diagnostics {
			text := fmt.Sprintf("[%d:%d] %s ",
				int(diagnistic.Range.Start.Line) + 1, int(diagnistic.Range.Start.Character + 1),
				diagnistic.Message,
			)
			options = append(options, text)
		}


		width := max(50, maxString(options))                   // width depends on max option len or 30 at min
		height := minMany(10, len(options) + 1)                // depends on min option len or 5 at min or how many rows to the end of e.screen
		atx := 0 + e.LINES_WIDTH + e.filesPanelWidth; aty := 0 // Define the window  position and dimensions
		style := StyleDefault.Foreground(ColorWhite)

		var selectionEnd = false; var selected = 0; var selectedOffset = 0

		for !selectionEnd {
			if selected < selectedOffset { selectedOffset = selected }  // calculate offsets for scrolling completion
			if selected >= selectedOffset+height { selectedOffset = selected - height + 1 }

			shifty := e.drawErrors(atx, width, aty, height, options, selectedOffset, selected, style)

			e.screen.Show()

			switch ev := e.screen.PollEvent().(type) { // poll and handle event
			case *EventResize:
				e.COLUMNS, e.ROWS = e.screen.Size()
				//ROWS -= 1
				e.screen.Sync()
				e.screen.Clear(); e.drawEverything(); e.screen.Show()

			case *EventKey:
				key := ev.Key()
				if key == KeyEscape || key == KeyEnter ||
					key == KeyBackspace || key == KeyBackspace2 ||
					key == KeyCtrlE { e.screen.Clear(); selectionEnd = true; end = true }
				if key == KeyDown { selected = min(len(options)-1, selected+1) }
				if key == KeyUp { selected = max(0, selected-1) }
				if key == KeyCtrlC {
					diagnostic := maybeDiagnostics.Diagnostics[selected]
					clipboard.WriteAll(diagnostic.Message)
				}
				//if key == tcell.KeyRight { e.onRight(); e.screen.Clear(); e.drawEverything(); selectionEnd = true }
				if key == KeyRight {
					diagnostic := maybeDiagnostics.Diagnostics[selected]
					e.r = int(diagnostic.Range.Start.Line)
					e.c = int(diagnostic.Range.Start.Character)
					e.focus();
					// add space for errors panel
					if e.r - e.y  < shifty + height { e.y -= shifty + height + 1}
					if e.y < 0 { e.y = 0 }
					e.drawEverything(); e.screen.Show()
				}
				//if key == tcell.KeyRune { e.addChar(ev.Rune()); e.writeFile(); e.screen.Clear(); e.drawEverything(); selectionEnd = true  }
				if key == KeyEnter {
					selectionEnd = true; end = true
					diagnostic := maybeDiagnostics.Diagnostics[selected]
					e.r = int(diagnostic.Range.Start.Line)
					e.c = int(diagnostic.Range.Start.Character)

					e.selection.ssx = e.c; e.selection.ssy = e.r;
					e.selection.sey = int(diagnostic.Range.End.Line)
					e.selection.sex = int(diagnostic.Range.End.Character)
					e.r = e.selection.sey; e.c = e.selection.sex
					e.selection.isSelected = true
					e.focus()
					// add space for errors panel
					if e.r - e.y  < shifty + height { e.y -= shifty + height + 1}
					if e.y < 0 { e.y = 0 }
					e.drawEverything(); e.screen.Show()
				}
			}
		}
	}

}

func (e *Editor) drawErrors(atx int, width int, aty int, height int, options []string,
	selectedOffset int, selected int, style Style) int {

	var shifty = 0
	for row := 0; row < aty+height; row++ {
		if row >= len(options) || row >= height {
			break
		}
		var option = options[row+selectedOffset]

		isRowSelected := selected == row+selectedOffset
		if isRowSelected {
			style = style.Background(Color(AccentColor))
		} else {
			//style = tcell.StyleDefault.Background(tcell.ColorIndianRed)
			style = StyleDefault.Background(Color(OverlayColor))
		}

		shiftx := 0
		runes := []rune(option)
		for j := 0; j < len(runes); j++ {
			ch := runes[j]
			nextWord := findNextWord(runes, j)
			if shiftx == 0 {
				e.screen.SetContent(atx, row+aty+shifty, ' ', nil, style)
			}
			if shiftx+atx+(nextWord-j) >= e.COLUMNS {
				for k := shiftx; k <= e.COLUMNS; k++ { // Fill the remaining space
					e.screen.SetContent(k+atx, row+aty+shifty, ' ', nil, style)
				}
				shifty++
				shiftx = 0
			}
			e.screen.SetContent(atx+shiftx, row+aty+shifty, ch, nil, style)
			shiftx++
		}

		for col := shiftx; col < e.COLUMNS; col++ { // Fill the remaining space
			e.screen.SetContent(col+atx, row+aty+shifty, ' ', nil, style)
		}
	}

	for col := 0; col < width; col++ { // Fill empty line below
		e.screen.SetContent(col+atx, height+aty+shifty-1, ' ', nil,
			StyleDefault.Background(Color(OverlayColor)))
	}

	return shifty
}

func (e *Editor) onSearch() {
	var end = false
	if e.searchPattern == nil { e.searchPattern = []rune{} }
	var patternx = len(e.searchPattern)
	var startline = e.y
	var isChanged = true
	var isDownSearch = true
	var prefix = []rune("search: ")

	// loop until escape or enter pressed
	for !end {

		e.drawSearch(prefix, e.searchPattern, patternx)
		e.screen.Show()

		if isChanged {
			var sy, sx = -1, -1
			if isDownSearch {
				sy, sx = searchDown(e.content, string(e.searchPattern), startline)
			} else {
				sy, sx = searchUp(e.content, string(e.searchPattern), startline)
			}

			if sx != -1 && sy != -1 {
				e.r = sy; e.c = sx; e.focus()
				startline = sy;
				e.selection.ssx = sx; e.selection.ssy = sy;
				e.selection.sex = sx + len(e.searchPattern); e.selection.sey = sy; e.selection.isSelected = true
				e.drawEverything()
				e.drawSearch(prefix, e.searchPattern, patternx)
				e.screen.ShowCursor(len(prefix) + patternx + e.LINES_WIDTH+ e.filesPanelWidth, e.ROWS-1)
				e.screen.Show()
			}else {
				e.selection.cleanSelection()
				if isDownSearch { startline = 0 } else  { startline = len(e.content)}
				e.drawEverything()
				e.drawSearch(prefix, e.searchPattern, patternx)
				e.screen.ShowCursor(len(prefix) + patternx + e.LINES_WIDTH+ e.filesPanelWidth, e.ROWS-1)
				e.screen.Show()
			}
		}

		switch ev := e.screen.PollEvent().(type) { // poll and handle event
		case *EventResize:
			e.COLUMNS, e.ROWS = e.screen.Size()
			//ROWS -= 1

		case *EventKey:
			isChanged = false
			key := ev.Key()

			if key == KeyRune {
				e.searchPattern = insert(e.searchPattern, patternx, ev.Rune())
				patternx++
				isChanged = true
			}
			if key == KeyBackspace2 && patternx > 0 && len(e.searchPattern) > 0 {
				patternx--
				e.searchPattern = remove(e.searchPattern, patternx)
				isChanged = true
			}
			if key == KeyLeft && patternx > 0 { patternx-- }
			if key == KeyRight && patternx < len(e.searchPattern) { patternx++ }
			if key == KeyDown  {
				isDownSearch = true
				if startline < len(e.content) {
					startline++
					isChanged = true
				} else {
					startline = 0
					isChanged = true
				}
			}
			if key == KeyUp {
				isDownSearch = false
				isChanged = true
				if startline == 0 { startline = len(e.content) } else { startline-- }
			}
			if key == KeyESC || key == KeyEnter || key == KeyCtrlF { end = true }
		}
	}
}
func (e *Editor) drawSearch(prefix []rune, pattern []rune, patternx int) {
	for i := 0; i < len(prefix); i++ {
		e.screen.SetContent(i + e.LINES_WIDTH+ e.filesPanelWidth, e.ROWS-1, prefix[i], nil, StyleDefault)
		//e.screen.Show()
	}

	e.screen.SetContent(len(prefix) + e.LINES_WIDTH+ e.filesPanelWidth, e.ROWS-1, ' ', nil, StyleDefault)
	//e.screen.Show()

	for i := 0; i < len(pattern); i++ {
		e.screen.SetContent(len(prefix) + i + e.LINES_WIDTH+ e.filesPanelWidth, e.ROWS-1, pattern[i], nil, StyleDefault)
		//e.screen.Show()
	}

	e.screen.ShowCursor(len(prefix) + patternx + e.LINES_WIDTH+ e.filesPanelWidth, e.ROWS-1)
	//e.screen.Show()

	for i := len(prefix) + len(pattern) + e.LINES_WIDTH + e.filesPanelWidth; i < e.COLUMNS; i++ {
		e.screen.SetContent(i, e.ROWS-1, ' ', nil, StyleDefault)
		//e.screen.Show()
	}
}

func (e *Editor) onFiles() {
	e.isFileSelection = true

	if e.filesPanelWidth == 0 {
		e.readUpdateFiles()
		if len(e.files) == 0 { return }
		e.filesPanelWidth = 28
	}

	if e.filename != "" { e.drawEverything() }

	var end = false
	var filterPattern = []rune{}
	var patternx = 0
	var isChanged = true

	// loop until escape or enter pressed
	for !end {

		var selectionEnd = false;

		for !selectionEnd {
			if e.fileSelectedIndex != -1 && e.fileSelectedIndex < e.fileScrollingOffset {
				e.fileScrollingOffset = e.fileSelectedIndex
			}
			if e.fileSelectedIndex >= e.fileScrollingOffset + e.ROWS {
				e.fileScrollingOffset = e.fileSelectedIndex - e.ROWS + 1
			}

			filteredFiles := e.files
			if e.isFilesSearch && len(filterPattern) > 0 {
				pattern := string(filterPattern)
				filteredFiles = []FileInfo{}

				for _, f := range e.files {
					var foundMatch = false
					foundMatch = strings.Contains(f.filename, pattern)
					if foundMatch { filteredFiles = append(filteredFiles, f) } else {
						matches, err := filepath.Match(pattern, f.filename)
						if err != nil { continue }
						if matches { filteredFiles = append(filteredFiles, f) }
					}
				}

				if isChanged { e.drawFiles(string(filterPattern), filteredFiles, patternx) }
			} else {
				if isChanged { e.drawFiles(string(filterPattern), e.files, patternx) }
			}

			e.screen.Show()

			switch ev := e.screen.PollEvent().(type) { // poll and handle event
			case *EventMouse:
				mx, my := ev.Position()
				buttons := ev.Buttons()
				modifiers := ev.Modifiers()

				//if buttons & Button1 == 1 && math.Abs(float64(filesPanelWidth - mx)) <= 5  {
				//	filesPanelWidth = mx
				//	return
				//}

				if mx > e.filesPanelWidth {
					selectionEnd = true; end = true; e.isFilesSearch = false;
				} else {
					if buttons & WheelDown != 0  && modifiers & ModCtrl != 0 && e.filesPanelWidth < e.COLUMNS  {
						e.filesPanelWidth++
						if e.filename != "" { e.drawEverything(); e.screen.Show() }
						continue
					}
					if buttons & WheelUp != 0  && modifiers & ModCtrl != 0  && e.filesPanelWidth > 0 {
						e.filesPanelWidth--
						if e.filename != "" { e.drawEverything(); e.screen.Show() }
						continue
					}

					if buttons & WheelDown != 0 &&  len(filteredFiles) > e.ROWS {
						if len(filteredFiles) > e.ROWS {
							if !e.isFilesSearch && e.fileScrollingOffset <  len(filteredFiles) - e.ROWS {
								e.fileScrollingOffset++
							}
							if e.isFilesSearch && e.fileScrollingOffset <  len(filteredFiles) - e.ROWS +1 {
								e.fileScrollingOffset++
							}

						}
					}
					if buttons & WheelUp != 0 && e.fileScrollingOffset > 0 {
						e.fileScrollingOffset--
					}

					if my < len(filteredFiles) { e.fileSelectedIndex = my + e.fileScrollingOffset }
					if buttons & Button1 == 1 {
						e.readUpdateFiles()
						e.fileSelectedIndex = my + e.fileScrollingOffset
						if e.fileSelectedIndex < 0  { continue }
						if e.fileSelectedIndex >= len(filteredFiles) { continue }
						if mx > len(filteredFiles[e.fileSelectedIndex].filename) { continue}
						selectionEnd = true; end = true
						selectedFile := filteredFiles[e.fileSelectedIndex]
						e.inputFile = selectedFile.fullfilename
						e.openFile(e.inputFile)
						e.isFilesSearch = false
					}
				}

			case *EventResize:
				e.COLUMNS, e.ROWS = e.screen.Size()
				if e.filename != "" { e.drawEverything(); e.screen.Show() }

			case *EventKey:
				key := ev.Key()

				if key == KeyCtrlF { e.isFilesSearch = !e.isFilesSearch }
				if key == KeyEscape && !e.isFilesSearch { selectionEnd = true; end = true; e.filesPanelWidth =  0 }
				if key == KeyEscape  && e.isFilesSearch {  e.isFilesSearch = false}
				if key == KeyDown { e.fileSelectedIndex = min(len(filteredFiles)-1, e.fileSelectedIndex+1) }
				if key == KeyUp { e.fileSelectedIndex = max(0, e.fileSelectedIndex-1) }
				if key == KeyRune {
					e.isFilesSearch = true
					filterPattern = insert(filterPattern, patternx, ev.Rune())
					patternx++
					isChanged = true
					e.fileSelectedIndex = 0
				}
				if key == KeyBackspace2  && e.isFilesSearch && patternx > 0 && len(filterPattern) > 0 {
					patternx--
					filterPattern = remove(filterPattern, patternx)
					isChanged = true
				}
				if key == KeyLeft && e.isFilesSearch && patternx > 0 {patternx--; isChanged = true }
				if key == KeyRight && e.isFilesSearch && patternx < len(filterPattern) { patternx++; isChanged = true }
				if key == KeyRight && !e.isFilesSearch  {
					e.filesPanelWidth++
					if e.filename != "" { e.drawEverything(); e.screen.Show() }
				}
				if key == KeyLeft && !e.isFilesSearch  && e.filesPanelWidth > 0  {
					e.filesPanelWidth--
					if e.filename != "" { e.drawEverything(); e.screen.Show() }
				}
				if key == KeyCtrlT {
					selectionEnd = true; end = true
					e.isFilesSearch = false
					e.filesPanelWidth = 0
				}
				if key == KeyEnter && e.fileSelectedIndex < len(filteredFiles) {
					selectionEnd = true; end = true
					e.isFilesSearch = false
					selectedFile := filteredFiles[e.fileSelectedIndex]
					e.inputFile = selectedFile.fullfilename
					e.openFile(e.inputFile)
				}
			}
		}
	}

	e.isFileSelection = false
}

func (e *Editor) drawFiles(pattern string, files []FileInfo, patternx int) {

	for row := 0; row < e.ROWS; row++ {
		for col := 0; col < e.filesPanelWidth ; col++ { // clean
			e.screen.SetContent(col, row, ' ', nil, StyleDefault)
		}
	}

	var offsety = 0

	for fileIndex := 0; fileIndex < len(files); fileIndex++ {
		if e.isFilesSearch && offsety == e.ROWS-1 { continue }
		style := StyleDefault

		//e.screen.SetContent(filesPanelWidth, offsety, '│', nil, style)

		if fileIndex >= len(files) || fileIndex >= e.ROWS { break }
		if fileIndex + max(e.fileScrollingOffset,0) >= len(files) { break }
		file := files[fileIndex + max(e.fileScrollingOffset,0)]


		isSelectedFile := e.isFileSelection && e.fileSelectedIndex != -1 && fileIndex + e.fileScrollingOffset == e.fileSelectedIndex
		if isSelectedFile {
			style = style.Foreground(Color(AccentColor))
		}
		if e.inputFile != "" && e.inputFile == file.fullfilename {
			style = style.Background(Color(AccentColor)).Foreground(ColorWhite)
		}

		for j, ch := range file.filename {
			if j+1 > e.filesPanelWidth-1 { break }
			e.screen.SetContent(j + 1, offsety, ch, nil, style)
		}

		offsety++
	}

	for row := 0; row <= e.ROWS; row++ {
		if row >= len(files) {
			for col := 0; col < e.filesPanelWidth ; col++ { // clean
				e.screen.SetContent(col, row, ' ', nil, StyleDefault)
			}
		}
		//e.screen.SetContent(filesPanelWidth, row, '│', nil, StyleDefault.Foreground(Color(AccentColor)))
	}

	e.screen.HideCursor()

	if e.isFilesSearch {
		pref := " search: "
		e.screen.ShowCursor(len(pref) + patternx, e.ROWS-1)
		for i, ch := range pref { // draw prefix
			e.screen.SetContent(i, e.ROWS-1, ch, nil, StyleDefault)
		}

		for i, ch := range pattern { // draw pattern
			e.screen.SetContent(i+len(pref), e.ROWS-1, ch, nil, StyleDefault)
		}
		for col := len(pref) + len(pattern); col < e.filesPanelWidth - 1; col++ { // clean
			e.screen.SetContent(col, e.ROWS-1, ' ', nil, StyleDefault)
		}
	}

}

//func (e *Editor) addTab() {
//	if e.filesInfo == nil || len(e.filesInfo) == 0 {
//		e.filesInfo = append(e.filesInfo, FileInfo{e.filename, e.absoluteFilePath, 1})
//	} else {
//		var tabExists = false
//
//		for i := 0; i < len(e.filesInfo); i++ {
//			ti := e.filesInfo[i]
//			if e.absoluteFilePath == ti.fullfilename {
//				ti.openCount += 1
//				e.filesInfo[i] = ti
//				tabExists = true
//			}
//		}
//
//		if !tabExists {
//			e.filesInfo = append(e.filesInfo, FileInfo{e.filename, e.absoluteFilePath, 1})
//		}
//
//		sort.SliceStable(e.filesInfo, func(i, j int) bool {
//			return e.filesInfo[i].openCount < e.filesInfo[j].openCount
//		})
//	}
//}
//
//func (e *Editor) drawTabs() {
//	e.COLUMNS, e.ROWS = e.screen.Size()
//
//	if len(e.filesInfo) == 0 { return }
//	if e.filesPanelWidth == 0 { return }
//	if e.filesPanelWidth == 0 { return }
//
//	e.ROWS -= 1
//	at := e.ROWS
//	fromx := 1
//	styleDefault := StyleDefault
//
//	for i := fromx; i < e.COLUMNS; i++ {
//		e.screen.SetContent(0, at, ' ', nil, styleDefault)
//	}
//
//	xpos := 0
//	for _, info := range e.filesInfo {
//		if xpos > e.COLUMNS { break }
//		for _, ch := range info.filename {
//			e.screen.SetContent(xpos + fromx, at, ch, nil, styleDefault)
//			xpos++
//		}
//
//		e.screen.SetContent(xpos + fromx, at, ' ', nil, styleDefault)
//		xpos++
//		e.screen.SetContent(xpos + fromx, at, ' ', nil, styleDefault)
//	}
//}

func (e *Editor) overlayFalse() {
	e.isOverlay = false
}

func (e *Editor) updateColors() {
	if !e.isColorize { return }
	if e.lang == "" { return }
	if len(e.content) >= 10000 {
		line := string(e.content[e.r])
		linecolors := highlighter.colorize(line, e.filename)
		e.colors[e.r] = linecolors[0]
	} else {
		code := convertToString(e.content)
		e.colors = highlighter.colorize(code, e.filename)
	}
}

func (e *Editor) updateColorsFull() {
	if !e.isColorize { return }
	if e.lang == "" { return }

	code := convertToString(e.content)
	e.colors = highlighter.colorize(code, e.filename)
}

func (e *Editor) updateColorsAtLine(at int) {
	if !e.isColorize { return }
	if e.lang == "" { return }
	if at >= len(e.colors) { return }

	line := string(e.content[at])
	if line == "" { e.colors[at] = []int{}; return }
	linecolors := highlighter.colorize(line, e.filename)
	e.colors[at] = linecolors[0]
}

// todo, get rid of this function, cause updateColors is slow for big files
func (e *Editor) updateNeeded() {
	e.update = true
	e.isContentChanged = true
	if len(e.content) <= 10000 { go e.writeFile() }
	e.updateColors()
}