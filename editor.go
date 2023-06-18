package main

import (
	"bufio"
	"fmt"
	"github.com/atotto/clipboard"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strings"
	"time"

	. "github.com/gdamore/tcell"
	"github.com/gdamore/tcell/encoding"
)

var content [][]rune      // characters
var COLUMNS, ROWS = 0, 0  //  term size
var r, c = 0, 0           // cursor position, row and column
var y, x = 0, 0           // offset for scrolling for row and column
var LS = 6                // left shift for line number
var ssx, ssy, sex, sey = -1, -1, -1, -1    // left shift for line number
var isSelected = false
var inputFile = ""  	  // exact user input
var filename = "" 	    // file name to show
var directory = ""        // directory
var absoluteFilePath = "" // file name and directory
var isFileChanged = false
var colorize = true
var colors [][]int        // characters colors
var highlighter = Highlighter{}
var lsp = LspClient{}
var lang = ""
var update = true
var isOverlay = false
var s Screen
var SelectionColor = 8
var OverlayColor = -1
var AccentColor = 303

type Editor struct {
	undo      []EditOperation
	redoStack []EditOperation
	logger Logger
	config Config
	langConf Lang
	tabWidth int
}

func (e *Editor) start() {
	e.config = GetConfig()

	e.logger = Logger{ }
	e.logger.start()
	highlighter.logger = e.logger
	lsp.logger = e.logger

	e.logger.info("starting edgo")

	s = e.initScreen()

	// reading file from cmd args
	if len(os.Args) > 1 {
		filename = os.Args[1]
		inputFile = filename
		err := e.openFile(filename)
		if err != nil {
			fmt.Println(err)
			os.Exit(130)
		}
	} else {
		fmt.Println("filename not found. usage: edgo [filename]")
		// in the future it will open current directory
		os.Exit(130)
	}

	// async initialize lsp
	go e.init_lsp()

	// main draw cycle
	for {
		if update {
			e.drawEverything()
			s.Show()
		}
		e.handleEvents()
	}
}

func (e *Editor) openFile(fname string) error {

	absoluteDir, err := filepath.Abs(path.Dir(fname))
	if err != nil { return err }
	directory = absoluteDir; filename = filepath.Base(fname)
	absoluteFilePath = path.Join(directory, filename)

	e.logger.info("open", directory, filename)

	lang = detectLang(filename)
	e.logger.info("lang is", lang)

	conf, found := e.config.Langs[lang]
	if !found { conf = UnknownLang }
	e.langConf = conf
	e.tabWidth = conf.TabWidth

	code := e.readFile()
	colors = highlighter.colorize(code, filename)

	e.undo = []EditOperation{}
	e.redoStack = []EditOperation{}

	return nil
}

func (e *Editor) initScreen() Screen {
	encoding.Register()
	screen, err := NewScreen()
	if err != nil { fmt.Fprintf(os.Stderr, "%v\n", err); os.Exit(1) }

	err2 := screen.Init()
	if err2 != nil { fmt.Fprintf(os.Stderr, "%v\n", err2); os.Exit(1) }

	screen.EnableMouse()
	screen.Clear()

	COLUMNS, ROWS = screen.Size()
	ROWS -= 2  // for status line

	return screen
}

func (e *Editor) drawEverything() {
	// calculate scrolling offsets for scrolling vertically and horizontally
	if r < y { y = r }
	if r >= y + ROWS { y = r - ROWS }
	if c < x { x = c }
	if c >= x + COLUMNS { x = c - COLUMNS + 1 }

	// draw line number and chars according to scrolling offsets
	for row := 0; row <= ROWS; row++ {
		ry := row + y  // index to get right row in characters buffer by scrolling offset y
		e.cleanLineAfter(0, row)
		if row >= len(content) || ry >= len(content) { break }
		e.drawLineNumber(ry, row)

		tabsOffset := 0
		for col := 0; col <= COLUMNS; col++ {
			cx := col + x // index to get right column in characters buffer by scrolling offset x
			if cx >= len(content[ry]) { break }
			ch := content[ry][cx]
			style := e.getStyle(ry, cx)
			if ch == '\t'  { //draw big cursor with next symbol color
				if ry == r && cx == c { style = StyleDefault.Background(Color(colors[ry][cx+1])) }
				for i := 0; i < e.tabWidth ; i++ {
					s.SetContent(col + LS + tabsOffset, row, ' ', nil, style)
					if i != e.tabWidth-1 { tabsOffset++ }
				}
			} else {
				s.SetContent(col + LS + tabsOffset, row, ch, nil, style)
			}
		}
	}

	e.drawDiagnostic()

	var changes = ""; if isFileChanged { changes = "*" }
	status := fmt.Sprintf(" %s %d %d %s%s", lang, r+1, c+1, inputFile, changes)
	e.drawText(0, ROWS+1, COLUMNS, ROWS+1, status)
	e.cleanLineAfter(len(status), ROWS+1)

	if r < len(content) && c < len(content[r]) && content[r][c] == '\t' { s.HideCursor()} else  {
		tabs := countTabs(content[r], c) * (e.tabWidth -1)
		s.ShowCursor(c-x+LS+tabs, r-y) // show cursor
	}

}

func (e *Editor) drawDiagnostic() {
	//lsp.someMapMutex2.Lock()
	maybeDiagnostic, found := lsp.file2diagnostic["file://" + absoluteFilePath]
	//lsp.someMapMutex2.Unlock()

	if found {
		//style := tcell.StyleDefault.Background(tcell.ColorIndianRed).Foreground(tcell.ColorWhite)
		style := StyleDefault.Foreground(Color(AccentColor))
		//textStyle := tcell.StyleDefault.Foreground(tcell.ColorIndianRed)

		for _, diagnostic := range maybeDiagnostic.Diagnostics {
			dline := int(diagnostic.Range.Start.Line)
			if dline >= len(content) { continue } // sometimes it out of content
			if dline - y > ROWS { continue } // sometimes it out of content

			// iterate over error range and, todo::fix
			//for i := dline; i <= int(diagnostic.Range.End.Line); i++ {
			//	if i >= len(content) { continue }
			//	tabs := countTabs(content[i], dline)
			//	for j := int(diagnostic.Range.Start.Character); j <= int(diagnostic.Range.End.Character); j++ {
			//		if j >= len(content[i]) { continue }
			//
			//		ch := content[dline][j]
			//		s.SetContent(j+LS + tabs*e.tabWidth + x, i-y, ch, nil, textStyle)
			//	}
			//}


			tabs := countTabs(content[dline], len(content[dline]))
			var shifty = 0
			errorMessage := "error: " + diagnostic.Message
			errorMessage = PadLeft(errorMessage, COLUMNS - len(content[dline]) - tabs*e.tabWidth - 5 - LS)

			// iterate over message characters and draw it
			for i, m := range errorMessage {
				ypos :=  dline - y
				if ypos < 0 || ypos >= len(content) { break }

				tabs = countTabs(content[dline], len(content[dline]))
				xpos := i + LS + len(content[dline+shifty]) + tabs*e.tabWidth + 5

				//for { // draw ch on the next line if not fit to screen
				//	if xpos >= COLUMNS {
				//		shifty++
				//		tabs = countTabs(content[dline+shifty], len(content[dline+shifty]))
				//		ypos +=  (i / COLUMNS) + 1
				//		if ypos >= len(content) { break}
				//		xpos = len(content[dline+shifty]) + 5 + (xpos % COLUMNS) % COLUMNS
				//	} else { break }
				//}

				s.SetContent( xpos,  ypos, m, nil,  style)
			}
		}

	}
}

func (e *Editor) drawLineNumber(brw int, row int) {
	var style = StyleDefault.Foreground(ColorDimGray)
	if brw == r { style = StyleDefault
	}
	lineNumber := centerNumber(brw + 1, LS)
	for index, char := range lineNumber {
		s.SetContent(index, row, char, nil, style)
	}
}

func (e *Editor) cleanLineAfter(x, y int) {
	for i := x; i < COLUMNS; i++ {
		s.SetContent(i, y, ' ', nil, StyleDefault)
	}
}

func (e *Editor) handleEvents() {
	update = true
	ev := s.PollEvent()      // Poll event
	switch ev := ev.(type) { // Process event
	case *EventMouse:
		mx, my := ev.Position()
		buttons := ev.Buttons()
		modifiers := ev.Modifiers()
		mx -= LS

		if mx < 0  { return }
		if my == ROWS  { return }

		// if click with control or option, lookup for definition or references

		if buttons & Button1 == 1 && ( modifiers & ModAlt != 0  || modifiers & ModCtrl != 0 )  {
			r = my + y
			if r > len(content)-1 { r = len(content) - 1 } // fit cursor to content

			count := 0; realCount := 0  // searching x position
			for _, ch := range content[r] {
				if count >= mx + x { break }
				if ch == '\t' {
					count += e.tabWidth; realCount++
				} else {
					count++; realCount++
				}
			}
			c = realCount

			if modifiers & ModAlt != 0 { e.onReferences() }
			if modifiers & ModCtrl != 0 { e.onDefinition() }

			return
		}

		if isSelected && buttons & Button1 == 1 {
			r = my + y
			if r > len(content)-1 { r = len(content) - 1 } // fit cursor to content

			count := 0; realCount := 0  // searching x position
			for _, ch := range content[r] {
				if count >= mx + x { break }
				if ch == '\t' {
					count += e.tabWidth; realCount++
				} else {
					count++; realCount++
				}
			}

			isTripleClick := isUnderSelection(realCount, r) &&
				len(getSelectedLines(content, ssx, ssy, sex, sey)) == 1

			if isTripleClick {
				r = my + y
				c = realCount
				if r > len(content)-1 { r = len(content) - 1 } // fit cursor to content
				if c > len(content[r]) { c = len(content[r]) }
				//if c < 0 { sex = len(content[r]) }

				ssx = 0; sex = len(content[r])
				ssy = r; sey = r
				return
			} else {
				e.cleanSelection()
			}
		}

		//fmt.Printf("Left button: %v\n", buttons&tcell.Button1)

		if buttons & WheelDown != 0 { e.onDown(); return }
		if buttons & WheelUp != 0 { e.onUp(); return }
		if buttons & Button1 == 0 && ssx == -1 { update = false; return }

		if buttons & Button1 == 1 {
			r = my + y
			if r > len(content)-1 { r = len(content) - 1 } // fit cursor to content

			count := 0; realCount := 0  // searching x position
			for _, ch := range content[r] {
				if count >= mx + x { break }
				if ch == '\t' {
					count += e.tabWidth; realCount++
				} else {
					count++; realCount++
				}
			}

			if c == realCount && len(getSelectedLines(content, ssx, ssy, sex, sey)) == 0 {
				// double click
				prw := findPrevWord(content[r], c)
				nxw := findNextWord(content[r], c)
				ssx, ssy = prw, r
				sex, sey = nxw, r
				c = nxw
				return
			}

			c = realCount


			if c < 0 { c = 0 }

			if ssx < 0  { ssx, ssy = c, r }
			if ssx >= 0  { sex, sey = c, r }
		}

		if buttons & Button1 == 0 {
			if ssx != -1 && sex != -1 {
				isSelected = true
			}
		}

	case *EventResize:
		COLUMNS, ROWS = s.Size()
		ROWS -= 2
		s.Sync()
		s.Clear()

	case *EventKey:
		key := ev.Key()
		modifiers := ev.Modifiers()

		if key == KeyUp   && modifiers == 3 { e.onSwapLinesUp(); return }   // control + shift + up
		if key == KeyDown && modifiers == 3 { e.onSwapLinesDown(); return } // control + shift + down
		if key == KeyBacktab { e.onBackTab(); e.writeFile(); return	 }
		if key == KeyTab { e.onTab(); e.writeFile(); return	 }
		if key == KeyCtrlH { e.onHover();  return }
		if key == KeyCtrlR { e.onReferences();  return }
		if key == KeyCtrlP { e.onSignatureHelp();  return }
		if key == KeyCtrlG { e.onDefinition();  return }
		if key == KeyCtrlE { e.onErrors();  return }
		if key == KeyCtrlC { e.onCopy(); return; }
		if key == KeyCtrlV { e.paste(); return }
		if key == KeyCtrlX { e.cut(); s.Clear() }
		if key == KeyCtrlD { e.duplicate() }

		if modifiers &ModShift != 0 && (
				key == KeyRight ||
				key == KeyLeft ||
				key == KeyUp ||
				key == KeyDown) {

			if ssx < 0 { ssx, ssy = c, r }
			if key == KeyRight { e.onRight() }
			if key == KeyLeft { e.onLeft() }
			if key == KeyUp { e.onUp() }
			if key == KeyDown { e.onDown() }
			if ssx >= 0 { sex, sey = c, r; isSelected = true }
			return
		}

		if key == KeyRune && modifiers &ModAlt != 0 {
			if len(content) > 0 { e.handleSmartMove(ev.Rune()) }
			return
		}
		if key == KeyDown && modifiers &ModAlt != 0 { e.handleSmartMoveDown(); return }
		if key == KeyUp && modifiers &ModAlt != 0 { e.handleSmartMoveUp(); return }

		if key == KeyRune {
			e.addChar(ev.Rune());
			if ev.Rune() == '.' { e.drawEverything(); s.Show(); e.onCompletion(); s.Clear();}
			if ev.Rune() == '(' { e.drawEverything(); s.Show(); e.onSignatureHelp(); s.Clear();}
		}

		if /*key == tcell.KeyEscape ||*/ key == KeyCtrlQ { s.Fini(); os.Exit(1) }
		if key == KeyCtrlS { e.writeFile() }
		if key == KeyEnter { e.onEnter(); s.Clear() }
		//if ev.Modifiers()&tcell.ModAlt != 0 && key == tcell.KeyBackspace || key == tcell.KeyBackspace2 {
		//	e.cut(); return
		//}
		if key == KeyBackspace || key == KeyBackspace2 { e. onDelete(); s.Clear() }

		if key == KeyDown { e.onDown(); e.cleanSelection() }
		if key == KeyUp { e.onUp(); e.cleanSelection() }
		if key == KeyLeft { e.onLeft(); e.cleanSelection() }
		if key == KeyRight { e.onRight(); e.cleanSelection() }
		if key == KeyCtrlT { } // TODO: tree
		if key == KeyCtrlF { } // TODO: find
		if key == KeyCtrlU { e.onUndo() }
		//if key == tcell.KeyCtrlR { e.redo() } // todo: fix it
		if key == KeyCtrlSpace { e.onCompletion(); s.Clear(); }

	}
}

func (e *Editor) readFile() string {
	/// if file is big, read only first 1000 lines and read rest async
	fileSize := getFileSize(absoluteFilePath)
	fileSizeMB := fileSize / (1024 * 1024) // Convert size to megabytes

	var code string
	if fileSizeMB >= 1 {
		//colorize = false
		code = e.readFileAndBuildContent(absoluteFilePath, 1000)

		go func() { // sync?? no need
			code = e.readFileAndBuildContent(absoluteFilePath, 1000000)
			if colorize { colors = highlighter.colorize(code, filename); e.drawEverything();s.Show() }
		}()

	} else {
		code = e.readFileAndBuildContent(absoluteFilePath, 1000000)
	}
	return code
}

func (e *Editor) init_lsp() {
	start := time.Now()

	// Getting the lsp command with args for a language:
	conf, ok := e.config.Langs[strings.ToLower(lang)]
	if !ok || len(conf.Lsp) == 0 { return }  // lang is not supported.

	started := lsp.start(lang, strings.Split(conf.Lsp, " "))
	if !started { return }

	var diagnosticUpdateChan = make(chan string)
	go lsp.receiveLoop(diagnosticUpdateChan)

	currentDir, _ := os.Getwd()

	lsp.init(currentDir)
	lsp.didOpen(absoluteFilePath, lang)

	e.drawEverything()

	lspStatus := "lsp started, elapsed " + time.Since(start).String()
	if !lsp.isReady { lspStatus = "lsp is not ready yet" }
	status := fmt.Sprintf(" %s %d %d %s %s", lang, r+1, c+1, inputFile, lspStatus)
	e.drawText(0, ROWS+1, COLUMNS, ROWS+1, status)
	e.cleanLineAfter(len(status), ROWS+1)
	s.Show()
	e.logger.info("lsp status", lspStatus)

	go func() {
		for range diagnosticUpdateChan {
			if isOverlay { continue}
			e.drawEverything()
			s.Show()
		}
	}()
}

func markOverlayFalse() {
	isOverlay = false
}

func (e *Editor) onCompletion() {
	if !lsp.isReady { return }
    isOverlay = true
	defer markOverlayFalse()

	var completionEnd = false

	// loop until escape or enter pressed
	for !completionEnd {

		start := time.Now()
		completion, err := lsp.completion(absoluteFilePath, r, c)
		elapsed := time.Since(start)

		lspStatus := "lsp completion, elapsed " + elapsed.String()
		status := fmt.Sprintf(" %s %d %d %s %s", lang, r+1, c+1, inputFile, lspStatus)
		e.drawText(0, ROWS+1, COLUMNS, ROWS+1, status)
		e.cleanLineAfter(len(status), ROWS+1)

		options := e.buildCompletionOptions(completion)
		if err != nil || len(options) == 0 { return }

		tabs := countTabs(content[r], c)
		atx := c + LS + tabs*e.tabWidth; aty := r + 1 - y // Define the window  position and dimensions
		width := max(30, maxString(options)) // width depends on max option len or 30 at min
		height := minMany(5, len(options), ROWS - (r - y)) // depends on min option len or 5 at min or how many rows to the end of screen
		style := StyleDefault.Foreground(ColorWhite)
		// if completion on last two rows of the screen - move window up
		if r - y  >= ROWS - 1 { aty -= min(5, len(options)); aty--; height = min(5, len(options)) }

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
						e.updateColors(); s.Show(); e.writeFile()
					}
			}
		}
	}
}

func (e *Editor) onHover() {
	if !lsp.isReady { return }

	isOverlay = true
	defer markOverlayFalse()

	var hoverEnd = false

	// loop until escape or enter pressed
	for !hoverEnd {

		start := time.Now()
		hover, err := lsp.hover(absoluteFilePath, r, c)
		elapsed := time.Since(start)

		lspStatus := "lsp hover, elapsed " + elapsed.String()
		status := fmt.Sprintf(" %s %d %d %s %s", lang, r+1, c+1, inputFile, lspStatus)
		e.drawText(0, ROWS+1, COLUMNS, ROWS+1, status)
		e.cleanLineAfter(len(status), ROWS+1)

		if err != nil || len(hover.Result.Contents.Value) == 0 { return }
		options := strings.Split(hover.Result.Contents.Value, "\n")
		if len(options) == 0 { return }

		tabs := countTabs(content[r], c)
		width := max(30, maxString(options)) // width depends on max option len or 30 at min
		height := minMany(10, len(options)) // depends on min option len or 5 at min or how many rows to the end of screen
		atx := c + LS + tabs*e.tabWidth; aty := r - height - y // Define the window  position and dimensions
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
	if !lsp.isReady { return }

	isOverlay = true
	defer markOverlayFalse()

	var end = false

	// loop until escape or enter pressed
	for !end {

		start := time.Now()
		signatureHelpResponse, err := lsp.signatureHelp(absoluteFilePath, r, c)
		elapsed := time.Since(start)

		lspStatus := "lsp signature help, elapsed " + elapsed.String()
		status := fmt.Sprintf(" %s %d %d %s %s", lang, r+1, c+1, inputFile, lspStatus)
		e.drawText(0, ROWS+1, COLUMNS, ROWS+1, status)
		e.cleanLineAfter(len(status), ROWS+1)

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

		tabs := countTabs(content[r], c)
		width := max(30, maxString(options)) // width depends on max option len or 30 at min
		height := minMany(10, len(options)) // depends on min option len or 5 at min or how many rows to the end of screen
		atx := c + LS + tabs*e.tabWidth; aty := r - height - y // Define the window  position and dimensions
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
	if !lsp.isReady { return }

	isOverlay = true
	defer markOverlayFalse()

	var end = false

	// loop until escape or enter pressed
	for !end {

		start := time.Now()
		referencesResponse, err := lsp.references(absoluteFilePath, r, c)
		elapsed := time.Since(start)

		lspStatus := "lsp references, elapsed " + elapsed.String()
		status := fmt.Sprintf(" %s %d %d %s %s", lang, r+1, c+1, inputFile, lspStatus)
		e.drawText(0, ROWS+1, COLUMNS, ROWS+1, status)
		e.cleanLineAfter(len(status), ROWS+1)

		if err != nil || len(referencesResponse.Result) == 0 { return }

		var options = []string{}
		for _, ref := range referencesResponse.Result {
			text := fmt.Sprintf(" %s %d %d ", ref.URI, ref.Range.Start.Line + 1, ref.Range.Start.Character + 1)
			options = append(options, text)
		}

		if len(options) == 0 { return }
		if len(options) == 1 {
			// if only one option, no need to draw options
			r = referencesResponse.Result[0].Range.Start.Line
			c = referencesResponse.Result[0].Range.Start.Character
			ssx = c; ssy = r;
			sey = referencesResponse.Result[0].Range.End.Line
			sex = referencesResponse.Result[0].Range.End.Character
			isSelected = true
			r = sey; c = sex
			s.Clear(); e.drawEverything()
			return
		}

		tabs := countTabs(content[r], c)
		width := max(30, maxString(options)) // width depends on max option len or 30 at min
		height := minMany(10, len(options)) // depends on min option len or 5 at min or how many rows to the end of screen
		atx := c + LS + tabs*e.tabWidth; aty := r - height - y // Define the window  position and dimensions
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
					if referencesResponse.Result[selected].URI != "file://"+ absoluteFilePath {  // if another file
						// do nothing
					} else {
						r = referencesResponse.Result[selected].Range.Start.Line
						c = referencesResponse.Result[selected].Range.Start.Character
						ssx = c; ssy = r;
						sey = referencesResponse.Result[selected].Range.End.Line
						sex = referencesResponse.Result[selected].Range.End.Character
						isSelected = true
						r = sey; c = sex
						s.Clear(); e.drawEverything();
					}
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
		style = style.Foreground(ColorWhite)

		s.SetContent(atx-1, row+aty, ' ', nil, style)
		for col, char := range option { s.SetContent(col+atx, row+aty, char, nil, style) }
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

	// add newText to chars
	for _, char := range newText {
		e.insertCharacter(r,c,char)
		c++
	}

}

func (e *Editor) onDefinition() {
	definition, err := lsp.definition(absoluteFilePath, r, c )

	if err != nil || len(definition.Result) == 0 ||
		definition.Result[0].URI != "file://" + absoluteFilePath || // same file
		int(definition.Result[0].Range.Start.Line) > len(content) ||  // not out of content
		int(definition.Result[0].Range.Start.Character) > len(content[int(definition.Result[0].Range.Start.Line)]) {
		return
	}


	r = int(definition.Result[0].Range.Start.Line)
	c = int(definition.Result[0].Range.Start.Character)
	ssx = c; ssy = r;
	sey = int(definition.Result[0].Range.End.Line)
	sex = int(definition.Result[0].Range.End.Character)
	r = sey; c = sex
	isSelected = true
}

func (e *Editor) onErrors() {
	if !lsp.isReady { return }

	maybeDiagnostics, found := lsp.file2diagnostic["file://" + absoluteFilePath]

	if !found || len(maybeDiagnostics.Diagnostics) == 0 { return }

	isOverlay = true
	defer markOverlayFalse()

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


		width := max(50, maxString(options)) // width depends on max option len or 30 at min
		height := minMany(len(options) + 1) // depends on min option len or 5 at min or how many rows to the end of screen
		atx := 0 + LS; aty := 0 // Define the window  position and dimensions
		style := StyleDefault.Foreground(ColorWhite)

		var selectionEnd = false; var selected = 0; var selectedOffset = 0

		for !selectionEnd {
			if selected < selectedOffset { selectedOffset = selected }  // calculate offsets for scrolling completion
			if selected >= selectedOffset+height { selectedOffset = selected - height + 1 }

			//draw errors
			var shifty = 0
			for row := 0; row < aty+height; row++ {
				if row >= len(options) || row >= height { break }
				var option = options[row+selectedOffset]

				isRowSelected := selected == row+selectedOffset
				if isRowSelected { style = style.Background(Color(AccentColor)) } else {
					//style = tcell.StyleDefault.Background(tcell.ColorIndianRed)
					style = StyleDefault.Background(Color(OverlayColor))
				}
				style = style.Foreground(ColorWhite)

				shiftx :=0
				runes := []rune(option)
				for j :=0;  j < len(option); j++ {
					ch := runes[j]
					nextWord := findNextWord(runes, j)
					if shiftx == 0 { s.SetContent(atx, row+aty+shifty, ' ', nil, style) }
					if shiftx+atx + (nextWord - j) >= COLUMNS {
						for k := shiftx; k <= COLUMNS; k++ { // Fill the remaining space
							s.SetContent(k+atx, row+aty+shifty, ' ', nil, style)
						}
						shifty++; shiftx = 0

					}
					s.SetContent(atx + shiftx, row+aty+shifty, ch, nil, style)
					shiftx++
				}

				for col := shiftx; col < COLUMNS; col++ { // Fill the remaining space
					s.SetContent(col+atx, row+aty+shifty, ' ', nil, style)
				}
			}

			for col := 0; col < width; col++ { // Fill empty line below
				s.SetContent(col+atx, height + aty + shifty-1, ' ', nil,
					StyleDefault.Background(Color(OverlayColor)))
			}

			s.Show()

			switch ev := s.PollEvent().(type) { // poll and handle event
			case *EventResize:
				COLUMNS, ROWS = s.Size()
				ROWS -= 2
				s.Sync()
				s.Clear(); e.drawEverything(); s.Show()

			case *EventKey:
				key := ev.Key()
				if key == KeyEscape || key == KeyEnter ||
					key == KeyBackspace || key == KeyBackspace2 ||
					key == KeyCtrlE { s.Clear(); selectionEnd = true; end = true }
				if key == KeyDown { selected = min(len(options)-1, selected+1) }
				if key == KeyUp { selected = max(0, selected-1) }
				if key == KeyCtrlC {
					diagnostic := maybeDiagnostics.Diagnostics[selected]
					clipboard.WriteAll(diagnostic.Message)
				}
				//if key == tcell.KeyRight { e.onRight(); s.Clear(); e.drawEverything(); selectionEnd = true }
				if key == KeyRight {
					diagnostic := maybeDiagnostics.Diagnostics[selected]
					r = int(diagnostic.Range.Start.Line)
					c = int(diagnostic.Range.Start.Character)
					s.Clear(); e.drawEverything(); s.Show()
				}
				//if key == tcell.KeyRune { e.addChar(ev.Rune()); e.writeFile(); s.Clear(); e.drawEverything(); selectionEnd = true  }
				if key == KeyEnter {
					selectionEnd = true; end = true
					diagnostic := maybeDiagnostics.Diagnostics[selected]
					r = int(diagnostic.Range.Start.Line)
					c = int(diagnostic.Range.Start.Character)

					ssx = c; ssy = r;
					sey = int(diagnostic.Range.End.Line)
					sex = int(diagnostic.Range.End.Character)
					r = sey; c = sex
					isSelected = true

					s.Clear(); e.drawEverything(); s.Show()
				}
			}
		}
	}

}

func (e *Editor) getSelectedStyle(isSelected bool, style Style) Style {
	if isSelected {
		style = style.Background(Color(AccentColor))
	} else {
		style = StyleDefault.Background(Color(SelectionColor))
	}
	return style.Foreground(ColorWhite)
}

func (e *Editor) cleanSelection() {
	isSelected = false
	ssx, ssy, sex, sey = -1, -1, -1, -1
}

func (e *Editor) getStyle(ry int, cx int) Style {
	var style = StyleDefault
	if !colorize { return style }
	if ry >= len(colors) || cx >= len(colors[ry])  { return style }
	color := colors[ry][cx]
	if color > 0 { style = StyleDefault.Foreground(Color(color)) }
	if isUnderSelection(cx, ry) {
		style = style.Background(Color(SelectionColor))
	}
	return style
}

func (e *Editor) addChar(ch rune) {
	if ssx != -1 && sex != -1 && isSelected  && ssx != sex { e.cut() }

	e.insertCharacter(r, c, ch)
	c++

	e.maybeAddPair(ch)

	if len(e.redoStack) > 0 { e.redoStack = []EditOperation{} }
	e.updateNeeded()
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
    r++
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
func (e *Editor) onDelete() {
	if len(getSelectionString(content, ssx, ssy, sex, sey)) > 0 { e.cut(); return }

	if c > 0 {
		c--
		e.deleteCharacter(r,c)

	} else if r > 0 { // delete line
		e.undo = append(e.undo, EditOperation{{DeleteLine, ' ', r-1, len(content[r-1])}})
		l := content[r][c:]
		content = remove(content, r)
		r--
		c = len(content[r])
		content[r] = append(content[r], l...)
	}

	if len(e.redoStack) > 0 { e.redoStack = []EditOperation{} }
	e.updateNeeded()
}

func (e *Editor) onDown() {
	if len(content) == 0 { return }
	if r+1 >= len(content) { return }
	r++
	if c > len(content[r]) { c = len(content[r]) } // fit to content
}

func (e *Editor) onUp() {
	if len(content) == 0 { return }
	if r == 0 { return }
	r--
	if c > len(content[r]) { c = len(content[r]) } // fit to content
}

func (e *Editor) onLeft() {
	if len(content) == 0 { return }

	if c > 0 {
		c--
	} else if r > 0 {
		r -= 1
		c = len(content[r]) // fit to content
	}
}
func (e *Editor) onRight() {
	if len(content) == 0 { return }

	if c < len(content[r]) {
		c++
	} else if r < len(content)-1 {
		r += 1 // to newline
		c = 0
	}
}

func (e *Editor) onEnter() {
	var ops = EditOperation{{Enter, '\n', r, c}}

	tabs := countTabs(content[r], c)
	spaces := countSpaces(content[r], c)

	after := content[r][c:]
	before := content[r][:c]
	content[r] = before
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

	e.undo = append(e.undo, ops)

	if len(e.redoStack) > 0 { e.redoStack = []EditOperation{} }
	e.updateNeeded()
}


func (e *Editor) writeFile() {

	// Create a new file, or open it if it exists
	f, err := os.Create(absoluteFilePath)
	if err != nil { panic(err) }

	// Create a buffered writer from the file
	w := bufio.NewWriter(f)

	for i, row := range content {
		for j := 0; j < len(row); {
			if _, err := w.WriteRune(row[j]); err != nil { panic(err) }
			j++
		}

		if i != len(content) - 1 { // do not write \n at the end
			if _, err := w.WriteRune('\n'); err != nil { panic(err) }
		}

	}

	// Don't forget to flush the buffered writer to ensure all data is written
	if err := w.Flush(); err != nil { panic(err) }
	if err := f.Close(); err != nil { panic(err) }

	isFileChanged = false

	if lsp.isReady {
		go lsp.didOpen(absoluteFilePath, lang) // todo remove it in future
		//go lsp.didChange(absoluteFilePath)
		//go lsp.didSave(absoluteFilePath)
	}

}

func (e *Editor) readFileAndBuildContent(filename string, limit int) string {
	file, err := os.Open(filename)
	if err != nil {
		filec, err2 := os.Create(filename)
		if err2 != nil {fmt.Printf("Failed to create file: %v\n", err2)}
		defer filec.Close()
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)

	lines := ""
	for scanner.Scan() {
		var line = scanner.Text()
		lines += line + "\n"
		var lineChars = []rune{}
		for _, char := range line {
			lineChars = append(lineChars, char)
		}
		content = append(content, lineChars)
		if len(content) >= limit { break }
	}

	// if no content, consider it like one line for next editing
	if content == nil {
		content = make([][]rune, 1)
		colors = make([][]int, 1)
	}

	return lines
}

func (e *Editor) drawText(x1, y1, x2, y2 int, text string) {
	row := y1
	col := x1
	var style = StyleDefault.Foreground(ColorGray)
	for _, ch := range []rune(text) {
		s.SetContent(col, row, ch, nil, style)
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
	if !colorize { return }
	if len(content) >= 10000 {
		line := string(content[r])
		linecolors := highlighter.colorize(line, filename)
		colors[r] = linecolors[0]
	} else {
		code := convertToString(content)
		colors = highlighter.colorize(code, filename)
	}
}

func (e *Editor) onTab() {
	selectedLines := getSelectedLines(content, ssx,ssy,sex,sey)

	if len(selectedLines) == 0 {
		ch := '\t'
		e.insertCharacter(r, c, ch);
		c++
	} else  {
		var ops = EditOperation{}
		ssx = 0
		for _, linenumber := range selectedLines {
			r = linenumber
			content[r] = insert(content[r], 0, '\t')
			ops = append(ops, Operation{Insert, '\t', r, 0})
			c = len(content[r])
		}
		sex = c
		e.undo = append(e.undo, ops)
	}

	if len(e.redoStack) > 0 { e.redoStack = []EditOperation{} }
	e.updateNeeded()
}

func (e *Editor) onBackTab() {
	selectedLines := getSelectedLines(content, ssx,ssy,sex,sey)

	// deleting tabs from begining
	if len(selectedLines) == 0 {
		if content[r][0] == '\t'  {
			e.deleteCharacter(r,0)
			//content[r] = content[r][1:] // delete first
			c--
		}
	} else {
		ssx = 0
		for _, linenumber := range selectedLines {
			r = linenumber
			if content[r][0] == '\t'  {
				e.deleteCharacter(r,0)
				//content[r] = content[r][1:] // delete first
				c = len(content[r])
			}
		}
	}

	if len(e.redoStack) > 0 { e.redoStack = []EditOperation{} }
	e.updateNeeded()
}

func (e *Editor) onCopy() {
	selectionString := getSelectionString(content, ssx, ssy, sex, sey)
	clipboard.WriteAll(selectionString)
}

func (e *Editor) paste() {
	if len(getSelectionString(content, ssx, ssy, sex, sey)) > 0 { e.cut() }

	text, _ := clipboard.ReadAll()
	lines := strings.Split(text, "\n")

	if len(lines) == 0 { return }

	if len(lines) == 1 { // single line paste
		e.insertString(r,c, lines[0])
	}

	if len(lines) > 1 { // multiple line paste
		e.insertLines(r,c, lines)
	}

	update = true
	e.updateNeeded()
}

func (e *Editor) cut() {
	if len(content) <= 1 {
		content[0] = []rune{};
		r, c = 0, 0
		return
	}
	var ops = EditOperation{}

	if len(getSelectionString(content, ssx, ssy, sex, sey)) == 0 { // cut single line
		ops = append(ops, Operation{MoveCursor, ' ', r, c})

		for i := len(content[r])-1; i >= 0; i-- {
			ops = append(ops, Operation{Delete, content[r][i], r, i})
		}

		if r == 0 {
			ops = append(ops, Operation{DeleteLine, '\n', 0, 0})
			c = 0
		} else {
			ops = append(ops, Operation{DeleteLine, '\n', r-1, len(content[r-1])})
			c = len(content[r-1])
		}

		content = append(content[:r], content[r+1:]...)
		if r > 0 { r-- }

	} else { // cut selection
		ops = append(ops, Operation{MoveCursor, ' ', r, c})

		selectedIndices := getSelectedIndices(content, ssx, ssy, sex, sey)

		// Sort selectedIndices in reverse order to delete characters from the end
		for i := len(selectedIndices) - 1; i >= 0; i-- {
			indices := selectedIndices[i]
			xd := indices[0]
			yd := indices[1]
			c, r = xd, yd

			// Delete the character at index (x, j)
			ops = append(ops, Operation{Delete, content[yd][xd], yd, xd})
			content[yd] = append(content[yd][:xd], content[yd][xd+1:]...)
			if len(content[yd]) == 0 { // delete line
				if r == 0 {
					ops = append(ops, Operation{DeleteLine, '\n', 0, 0})
				} else {
					ops = append(ops, Operation{DeleteLine, '\n', r-1, len(content[r-1])})
				}

				content = append(content[:yd], content[yd+1:]...)
				colors = append(colors[:yd], colors[yd+1:]...)
			}
		}
		if r >= len(content) {
			r = len(content) - 1
			c = len(content[r])
		}
		e.cleanSelection()
		e.undo = append(e.undo, ops)
	}

	e.undo = append(e.undo, ops)
	e.updateNeeded()
}

func (e *Editor) duplicate() {
	if len(content) == 0 { return }

	if ssx == -1 && ssy == -1 || ssx == sex && ssy == sey  {
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
		e.undo = append(e.undo, ops)
	} else {
		selection := getSelectionString(content, ssx,ssy,sex,sey)
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
		e.cleanSelection()
	}

	e.updateNeeded()
}

func (e *Editor) onSwapLinesUp() {
	if r == 0 { return }
	var ops = EditOperation{}
	ops = append(ops, Operation{MoveCursor, ' ', r, c})

	line1 := content[r]; line2 := content[r-1]

	for i := len(line1)-1; i >= 0; i-- { ops = append(ops, Operation{Delete, line1[i], r, i}) }
	for i := len(line2)-1; i >= 0; i-- { ops = append(ops, Operation{Delete, line2[i], r-1, i}) }
	for i, ch := range line1 { ops = append(ops, Operation{Insert, ch, r-1, i}) }
	for i, ch := range line2 { ops = append(ops, Operation{Insert, ch, r, i}) }

	content[r] = line2; content[r-1] = line1 // swap
	r--

	e.undo = append(e.undo, ops)
	e.cleanSelection()
	e.updateNeeded()
}

func (e *Editor) onSwapLinesDown() {
	if r+1 == len(content) { return }

	var ops = EditOperation{}
	ops = append(ops, Operation{MoveCursor, ' ', r, c})

	line1 := content[r]
	line2 := content[r+1]

	for i := len(line1)-1; i >= 0; i-- { ops = append(ops, Operation{Delete, line1[i], r, i}) }
	for i := len(line2)-1; i >= 0; i-- { ops = append(ops, Operation{Delete, line2[i], r+1, i}) }
	for i, ch := range line1 { ops = append(ops, Operation{Insert, ch, r+1, i}) }
	for i, ch := range line2 { ops = append(ops, Operation{Insert, ch, r, i}) }

	content[r] = line2; content[r+1] = line1 // swap
	r++

	e.undo = append(e.undo, ops)
	e.cleanSelection()
	e.updateNeeded()
}

func (e *Editor) updateNeeded() {
	update = true
	isFileChanged = true
	if len(content) <= 10000 { e.writeFile() }
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

	e.undo = append(e.undo, ops)
	e.updateNeeded()
}
func (e *Editor) handleSmartMoveUp() {
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

	e.undo = append(e.undo, ops)
	e.updateNeeded()
}

func (e *Editor) onUndo() {
	if len(e.undo) == 0 { return }

	lastOperation := e.undo[len(e.undo)-1]
	e.undo = e.undo[:len(e.undo)-1]

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

func isUnderSelection(x, y int) bool {
	// Check if there is an active selection
	if ssx == -1 || ssy == -1  || sex == -1 || sey == -1{ return false }

	var startx, starty = ssx, ssy
	var endx, endy = sex, sey

	if GreaterThan(startx, starty, endx, endy) {
		startx, endx = endx, startx
		starty, endy = endy, starty
	}

	return GreaterEqual(x, y, startx, starty) && LessThan(x, y, endx, endy)
}