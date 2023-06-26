package editor

import (
	. "edgo/internal/config"
	. "edgo/internal/highlighter"
	. "edgo/internal/logger"
	. "edgo/internal/lsp"
	. "edgo/internal/operations"
	. "edgo/internal/search"
	. "edgo/internal/selection"
	. "edgo/internal/utils"

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

var EditorGlobal = Editor{ }

type Editor struct {
	COLUMNS     int // terminal size columns
	ROWS        int // terminal size rows
	LINES_WIDTH int // draw file lines number

	Row int // cursor position row
	Col int // cursor position column
	Y   int // row offset for scrolling
	X   int // col offset for scrolling

	Content [][]rune // text characters
	Colors  [][]int  // text characters colors

	Screen Screen // Screen for drawing

	Lang         string // current file language
	Config       Config // config, lsp, tabs, comments, etc
	langConf     Lang   // current lang conf
	langTabWidth int    // current lang tabs indentation  '\t' -> "    "

	Selection Selection // selection

	Undo []EditOperation // stack for undo operations
	Redo []EditOperation // stack for redo operations

	InputFile        string // exact user input
	Filename         string // current file name
	AbsoluteFilePath string // current file name and directory
	IsContentChanged bool   // shows * if file is changed
	IsColorize       bool   // colorize text is true by default
	Update           bool   // for Screen updates,  if false it will not draw
	IsOverlay        bool   // true if overlay is active (completion, hover, errors...)

	FilesPanelWidth     int        // current width for files panel
	Files               []FileInfo // current dir files
	IsFileSelection     bool       // true if in files selection menu
	FileScrollingOffset int        // for vertical scrolling  in selection menu
	FileSelectedIndex   int        // selected file index
	IsFilesSearch       bool       // true if in files search mode
	IsFilesPanelMoving  bool       // true if in files panel moving mode

	SearchPattern []rune // pattern for search in a buffer

	//filesInfo []FileInfo

	CursorHistory     []CursorMove
	CursorHistoryUndo []CursorMove

}

func (e *Editor) Start() {
	Log.Info("starting edgo")

	e.InitScreen()

	// reading file from cmd args
	if len(os.Args) == 1 {
		e.OnFiles()
	} else {
		e.Filename = os.Args[1]
		e.InputFile = e.Filename
		err := e.OpenFile(e.Filename)
		if err != nil {
			fmt.Println(err)
			os.Exit(130)
		}
	}

	// main draw cycle
	for {
		if e.Update && e.Filename != "" {
			e.DrawEverything()
			e.Screen.Show()
		}
		e.HandleEvents()
	}
}

func (e *Editor) HandleEvents() {
	e.Update = true
	ev := e.Screen.PollEvent()
	switch ev := ev.(type) {
	case *EventResize:
		e.COLUMNS, e.ROWS = e.Screen.Size()

	case *EventMouse:
		mx, my := ev.Position()
		buttons := ev.Buttons()
		modifiers := ev.Modifiers()

		e.HandleMouse(mx, my, buttons, modifiers)

	case *EventKey:
		key := ev.Key()
		modifiers := ev.Modifiers()

		e.HandleKeyboard(key, ev, modifiers)
	}
}

func (e *Editor) HandleMouse(mx int, my int, buttons ButtonMask, modifiers ModMask) {
	if !e.IsFilesPanelMoving && buttons & Button1 == 1 &&
		math.Abs(float64(e.FilesPanelWidth- mx)) <= 2 &&
		len(e.Selection.GetSelectedLines(e.Content)) == 0 {

		e.FilesPanelWidth = mx - 1
		e.IsFilesPanelMoving = true
		return
	}

	if e.IsFilesPanelMoving && buttons & Button1 == 1 { e.FilesPanelWidth = mx; return }
	if e.IsFilesPanelMoving && buttons & Button1 == 0 { e.IsFilesPanelMoving = false; return }
	if mx < e.FilesPanelWidth- 2 && buttons & Button1 == 0 { e.OnFiles(); return }

	if e.Filename == "" { return }

	mx -= e.LINES_WIDTH + e.FilesPanelWidth

	if mx < 0 { return }
	if my > e.ROWS { return }

	// if click with control or option, lookup for definition or references
	if buttons & Button1 == 1 && (modifiers & ModAlt != 0 || modifiers & ModCtrl != 0) {
		e.Row = my + e.Y
		if e.Row > len(e.Content)-1 { e.Row = len(e.Content) - 1 } // fit cursor to e.Content

		e.Col = e.FindCursorXPosition(mx)
		if modifiers & ModAlt != 0 { e.OnReferences() }
		if modifiers & ModCtrl != 0 { e.OnDefinition() }
		return
	}

	if e.Selection.IsSelected && buttons & Button1 == 1 {
		e.Row = my + e.Y
		if e.Row > len(e.Content)-1 { e.Row = len(e.Content) - 1 } // fit cursor to e.Content

		xPosition := e.FindCursorXPosition(mx)

		isTripleClick := e.Selection.IsUnderSelection(xPosition, e.Row) &&
			len(e.Selection.GetSelectedLines(e.Content)) == 1

		if isTripleClick {
			e.Row = my + e.Y
			e.Col = xPosition
			if e.Row > len(e.Content)-1 { e.Row = len(e.Content) - 1 } // fit cursor to e.Content
			if e.Col > len(e.Content[e.Row]) { e.Col = len(e.Content[e.Row]) }
			//if e.Col < 0 { Sex = len(e.Content[Row]) }

			e.Selection.Ssx = 0
			e.Selection.Sex = len(e.Content[e.Row])
			e.Selection.Ssy = e.Row
			e.Selection.Sey = e.Row

			return
		} else {
			e.Selection.CleanSelection()
		}
	}

	if buttons & WheelDown != 0 { e.OnScrollDown(); return }
	if buttons & WheelUp != 0 { e.OnScrollUp(); return }
	if buttons & Button1 == 0 && e.Selection.Ssx == -1 { e.Update = false; return }

	if buttons & Button1 == 1 {
		e.Row = my + e.Y
		if e.Row > len(e.Content)-1 { e.Row = len(e.Content) - 1 } // fit cursor to e.Content

		xPosition := e.FindCursorXPosition(mx)

		if e.Col == xPosition && len(e.Selection.GetSelectedLines(e.Content)) == 0 {
			// double click
			prw := FindPrevWord(e.Content[e.Row], e.Col)
			nxw := FindNextWord(e.Content[e.Row], e.Col)
			e.Selection.Ssx, e.Selection.Ssy = prw, e.Row
			e.Selection.Sex, e.Selection.Sey = nxw, e.Row
			e.Col = nxw
			return
		}
		e.Col = xPosition
		e.CursorHistory = append(e.CursorHistory, CursorMove{e.AbsoluteFilePath, e.Row, e.Col})

		if e.Col < 0 { e.Col = 0 }
		if e.Selection.Ssx < 0 { e.Selection.Ssx, e.Selection.Ssy = e.Col, e.Row }
		if e.Selection.Ssx >= 0 { e.Selection.Sex, e.Selection.Sey = e.Col, e.Row }
	}

	if buttons&Button1 == 0 {
		if e.Selection.Ssx != -1 && e.Selection.Sex != -1 {
			e.Selection.IsSelected = true
		}
	}
	return
}
func (e *Editor) HandleKeyboard(key Key, ev *EventKey,  modifiers ModMask) {
	if e.Filename == "" && key != KeyCtrlQ { return }

	if ev.Rune() == '/' && modifiers&ModAlt != 0 || int(ev.Rune()) == '÷' {
		// '÷' on Mac is option + '/'
		e.OnCommentLine(); return
	}
	if key == KeyUp && modifiers == 3 { e.OnSwapLinesUp(); return } // control + shift + up
	if key == KeyDown && modifiers == 3 { e.OnSwapLinesDown(); return } // control + shift + down
	if key == KeyBacktab { e.OnBackTab(); return }
	if key == KeyTab { e.OnTab(); return }
	if key == KeyCtrlH { e.OnHover(); return }
	if key == KeyCtrlR { e.OnReferences(); return }
	if key == KeyCtrlW { e.OnCodeAction(); return }
	if key == KeyCtrlP { e.OnSignatureHelp(); return }
	if key == KeyCtrlG { e.OnDefinition(); return }
	if key == KeyCtrlE { e.OnErrors(); return }
	if key == KeyCtrlC { e.OnCopy(); return }
	if key == KeyCtrlV { e.OnPaste(); return }
	if key == KeyEscape { e.Selection.CleanSelection(); return }
	if key == KeyCtrlA { e.OnSelectAll(); return }
	if key == KeyCtrlX { e.Cut() }
	if key == KeyCtrlD { e.Duplicate() }

	if modifiers & ModShift != 0 && (
		key == KeyRight ||
			key == KeyLeft ||
			key == KeyUp ||
			key == KeyDown) {

		if e.Selection.Ssx < 0 { e.Selection.Ssx, e.Selection.Ssy = e.Col, e.Row
		}
		if key == KeyRight { e.OnRight() }
		if key == KeyLeft { e.OnLeft() }
		if key == KeyUp { e.OnUp() }
		if key == KeyDown { e.OnDown() }
		if e.Selection.Ssx >= 0 {
			e.Selection.Sex, e.Selection.Sey = e.Col, e.Row
			e.Selection.IsSelected = true
		}
		return
	}

	if key == KeyRune && modifiers & ModAlt != 0 && len(e.Content) > 0 { e.HandleSmartMove(ev.Rune()); return }
	if key == KeyDown && modifiers & ModAlt != 0 { e.HandleSmartMoveDown(); return }
	if key == KeyUp && modifiers & ModAlt != 0 { e.HandleSmartMoveUp(); return }

	if key == KeyRune {
		e.AddChar(ev.Rune())
		if ev.Rune() == '.' {
			e.DrawEverything(); e.Screen.Show()
			e.OnCompletion()
		}
		//if ev.Rune() == '(' { e.DrawEverything(); e.Screen.Show(); e.OnSignatureHelp(); e.Screen.Clear() }
	}

	if /*key == tcell.KeyEscape ||*/ key == KeyCtrlQ { e.Screen.Fini(); os.Exit(1) }
	if key == KeyCtrlS { e.WriteFile() }
	if key == KeyEnter { e.OnEnter() }
	if key == KeyBackspace || key == KeyBackspace2 { e.OnDelete() }
	if key == KeyDown { e.OnDown(); e.Selection.CleanSelection() }
	if key == KeyUp { e.OnUp(); e.Selection.CleanSelection() }
	if key == KeyLeft { e.OnLeft(); e.Selection.CleanSelection() }
	if key == KeyRight { e.OnRight(); e.Selection.CleanSelection() }
	if key == KeyCtrlT { e.OnFiles() }
	if key == KeyCtrlF { e.OnSearch() }
	if key == KeyF18 { e.OnRename() }
	if key == KeyCtrlU { e.OnUndo() }
	//if key == KeyCtrlR { e.OnRedo() } // todo: fix i
	if key == KeyCtrlO { e.OnCursorBack() }
	if key == KeyCtrlRightSq { e.OnCursorBackUndo() }

	if key == KeyCtrlSpace { e.OnCompletion() }

}

func (e *Editor) OpenFile(fname string) error {

	absoluteDir, err := filepath.Abs(path.Dir(fname))
	if err != nil { return err }
	//directory := absoluteDir;
	e.Filename = filepath.Base(fname)
	e.AbsoluteFilePath = path.Join(absoluteDir, e.Filename)

	Log.Info("open", e.AbsoluteFilePath)

	newLang := DetectLang(e.Filename)
	Log.Info("new lang is", newLang)

	if newLang != "" && newLang != e.Lang {
		e.Lang = newLang
		Lsp.Lang = newLang
		ready := Lsp.IsLangReady(e.Lang)
		if !ready { go e.InitLsp(e.Lang) }
	}

	conf, found := e.Config.Langs[e.Lang]
	if !found { conf = DefaultLangConfig }
	e.langConf = conf
	e.langTabWidth = conf.TabWidth

	code := e.ReadFile(e.AbsoluteFilePath)
	e.Colors = HighlighterGlobal.Colorize(code, e.Filename)

	e.Undo = []EditOperation{}
	e.Redo = []EditOperation{}

	e.UpdateFilesOpenStats(fname)

    e.Row = 0; e.Col = 0; e.Y = 0; e.X = 0
	e.Selection = Selection{-1,-1,-1,-1,false }

	return nil
}

func (e *Editor) InitScreen() {
	encoding.Register()
	screen, err := NewScreen()
	if err != nil { fmt.Fprintf(os.Stderr, "%v\n", err); os.Exit(1) }
	e.Screen = screen

	err2 := e.Screen.Init()
	if err2 != nil { fmt.Fprintf(os.Stderr, "%v\n", err2); os.Exit(1) }

	e.Screen.EnableMouse()
	e.Screen.Clear()

	e.COLUMNS, e.ROWS = e.Screen.Size()
	//ROWS -= 1
	
	e.LINES_WIDTH = 6
	e.Update = true
	e.IsColorize = true
	e.FileSelectedIndex = -1
	e.CursorHistory = []CursorMove{}

	return
}

func (e *Editor) DrawEverything() {
	if len(e.Content) == 0 { return }
	e.Screen.Clear()

	if e.FilesPanelWidth != 0 { e.DrawFiles("", e.Files, 0, 0) }

	
	//tabs := CountTabsTo(e.Content[e.Row], e.Col)
	//correction := tabs*(e.langTabWidth - 1)
	countTabsTo := CountTabsTo(e.Content[e.Row], e.Col)
	tabcor := countTabsTo *(e.langTabWidth - 1)
	// todo: fix horizontal scrolling
	if e.Col < e.X { e.X = e.Col }
	if e.Col + e.LINES_WIDTH + e.FilesPanelWidth + tabcor >= e.X + e.COLUMNS  {
		e.X = e.Col - e.COLUMNS + 1 + e.LINES_WIDTH + e.FilesPanelWidth + tabcor
	}

	// draw Line number and chars according to scrolling offsets
	for row := 0; row < e.ROWS; row++ {
		ry := row + e.Y // index to get right row in characters buffer by scrolling offset Y
		//e.cleanLineAfter(0, row)
		if row >= len(e.Content) || ry >= len(e.Content) { break }
		e.DrawLineNumber(ry, row)

		tabsOffset := 0
		for col := 0; col <= e.COLUMNS; col++ {
			cx := col + e.X // index to get right column in characters buffer by scrolling offset x
			if cx < 0 { break }
			if cx >= len(e.Content[ry]) { break }
			ch := e.Content[ry][cx]
			style := e.GetStyle(ry, cx)
			if ch == '\t' && e.X == 0  {
				//draw big cursor with next symbol color
				if ry == e.Row && cx == e.Col {
					var color = Color(AccentColor)
					if cx+1 < len(e.Colors[ry]) { color = Color(e.Colors[ry][cx+1]) }
					if color == -1 { color = Color(AccentColor)}
					style = StyleDefault.Background(color)
				}
				for i := 0; i < e.langTabWidth; i++ {
					e.Screen.SetContent(col + e.LINES_WIDTH + tabsOffset + e.FilesPanelWidth, row, ' ', nil, style)
					if i != e.langTabWidth-1 { tabsOffset++ }
				}
			} else {
				e.Screen.SetContent(col + e.LINES_WIDTH+ tabsOffset + e.FilesPanelWidth, row , ch, nil, style)
			}
		}
	}

	e.DrawDiagnostic()
	//e.drawTabs()

	var changes = ""; if e.IsContentChanged { changes = "*" }
	status := fmt.Sprintf(" %s %d %d %s%s ", e.Lang, e.Row+1, e.Col+1, e.Filename, changes)
	e.DrawStatus(status)

	// if tab under cursor, hide cursor because it has already drawn
	if e.Row < len(e.Content) && e.Col < len(e.Content[e.Row]) && e.Content[e.Row][e.Col] == '\t' {
		e.Screen.HideCursor()
	} else  {
		tabs := CountTabsTo(e.Content[e.Row], e.Col) * (e.langTabWidth - 1)
		e.Screen.ShowCursor(e.Col - e.X + e.LINES_WIDTH+tabs + e.FilesPanelWidth, e.Row - e.Y) // show cursor
		if e.X != 0 {
			e.Screen.ShowCursor(e.Col - e.X + e.LINES_WIDTH + e.FilesPanelWidth, e.Row - e.Y) // show cursor
		}
	}

}

func (e *Editor) GetStyle(ry int, cx int) Style {
	var style = StyleDefault
	if !e.IsColorize { return style }
	if ry >= len(e.Colors) || cx >= len(e.Colors[ry])  { return style }
	color := e.Colors[ry][cx]
	if color > 0 { style = StyleDefault.Foreground(Color(color)) }
	if e.Selection.IsUnderSelection(cx, ry) {
		style = style.Background(Color(SelectionColor))
	}
	return style
}

func (e *Editor) DrawDiagnostic() {
	//lsp.someMapMutex2.Lock()
	maybeDiagnostic, found := Lsp.GetDiagnostic("file://" + e.AbsoluteFilePath)
	//lsp.someMapMutex2.Unlock()

	if found {
		//style := tcell.StyleDefault.Background(tcell.ColorIndianRed).Foreground(tcell.ColorWhite)
		style := StyleDefault.Foreground(Color(AccentColor))
		//textStyle := tcell.StyleDefault.Foreground(tcell.ColorIndianRed)

		for _, diagnostic := range maybeDiagnostic.Diagnostics {
			dline := int(diagnostic.Range.Start.Line)
			if dline >= len(e.Content) { continue } // sometimes it out of e.Content
			if dline - e.Y > e.ROWS { continue } // sometimes it out of e.Content

			// iterate over error range and, todo::fix
			//for i := dline; i <= int(diagnostic.Range.End.Line); i++ {
			//	if i >= len(e.Content) { continue }
			//	tabs := CountTabs(e.Content[i], dline)
			//	for j := int(diagnostic.Range.Start.Character); j <= int(diagnostic.Range.End.Character); j++ {
			//		if j >= len(e.Content[i]) { continue }
			//
			//		ch := e.Content[dline][j]
			//		e.Screen.SetContent(j+LINES_WIDTH + tabs*e.langTabWidth + X, i-Y, ch, nil, textStyle)
			//	}
			//}


			tabs := CountTabs(e.Content[dline], len(e.Content[dline]))
			var shifty = 0
			errorMessage := "error: " + diagnostic.Message
			errorMessage = PadLeft(errorMessage, e.COLUMNS - len(e.Content[dline]) - tabs*e.langTabWidth- 5 - e.LINES_WIDTH- e.FilesPanelWidth)

			// iterate over message characters and draw it
			for i, m := range errorMessage {
				ypos :=  dline - e.Y
				if ypos < 0 || ypos >= len(e.Content) { break }

				tabs = CountTabs(e.Content[dline], len(e.Content[dline]))
				xpos := i + e.LINES_WIDTH + e.FilesPanelWidth + len(e.Content[dline+shifty]) + tabs*e.langTabWidth + 5

				//for { // draw ch on the next Line if not fit to e.Screen
				//	if xpos >= COLUMNS {
				//		shifty++
				//		tabs = CountTabs(e.Content[dline+shifty], len(e.Content[dline+shifty]))
				//		ypos +=  (i / COLUMNS) + 1
				//		if ypos >= len(e.Content) { break}
				//		xpos = len(e.Content[dline+shifty]) + 5 + (xpos % COLUMNS) % COLUMNS
				//	} else { break }
				//}

				e.Screen.SetContent(xpos,  ypos, m, nil,  style)
			}
		}

	}
}

func (e *Editor) DrawLineNumber(brw int, row int) {
	var style = StyleDefault.Foreground(ColorDimGray)
	if brw == e.Row { style = StyleDefault}
	lineNumber := CenterNumber(brw + 1, e.LINES_WIDTH)
	for index, char := range lineNumber {
		e.Screen.SetContent(index + e.FilesPanelWidth, row, char, nil, style)
	}
}

func (e *Editor) DrawStatus(text string) {
	var style = StyleDefault
	e.DrawText(e.ROWS-1, e.COLUMNS - len(text), text, style)
}

func (e *Editor) DrawText(row, col int, text string, style Style) {
	e.Screen.SetContent(col-1, row, ' ', nil, style)
	for _, ch := range []rune(text) {
		if col > e.COLUMNS { break }
		e.Screen.SetContent(col, row, ch, nil, style)
		col++
	}
}

func (e *Editor) FindCursorXPosition(mx int) int {
	count := 0; realCount := 0  // searching x position
	for _, ch := range e.Content[e.Row] {
		if count >= mx + e.X { break }
		if ch == '\t'  && e.X == 0 {
			count += e.langTabWidth; realCount++
		} else {
			count++; realCount++
		}
	}
	return realCount
}

func (e *Editor) InitLsp(lang string) {
	//Start := time.Now()

	// Getting the lsp command with args for a language:
	conf, ok := e.Config.Langs[strings.ToLower(lang)]
	if !ok || len(conf.Lsp) == 0 { return }  // lang is not supported.

	started := Lsp.Start(lang, strings.Split(conf.Lsp, " "))
	if !started { return }

	var diagnosticUpdateChan = make(chan string)
	go Lsp.ReceiveLoop(diagnosticUpdateChan, lang)

	currentDir, _ := os.Getwd()

	Lsp.Init(currentDir)
	Lsp.DidOpen(e.AbsoluteFilePath, lang)

	//e.DrawEverything()
	//
	//lspStatus := "lsp started, elapsed " + time.Since(Start).String()
	//if !lsp.isReady { lspStatus = "lsp is not ready yet" }
	//Log.Info("lsp status", lspStatus)
	//status := fmt.Sprintf(" %e.Screen %e.Screen %d %d %e.Screen ", lspStatus,  lang, Row+1, Col+1, InputFile)
	//e.drawText(COLUMNS- len(status), ROWS-1, COLUMNS, ROWS-1, status)
	//e.Screen.Show()

	go func() {
		for range diagnosticUpdateChan {
			if e.IsOverlay { continue }
			e.DrawEverything()
			e.Screen.Show()
		}
	}()
}

func (e *Editor) OnErrors() {
	if !Lsp.IsLangReady(e.Lang) { return }

	maybeDiagnostics, found := Lsp.GetDiagnostic("file://" + e.AbsoluteFilePath)

	if !found || len(maybeDiagnostics.Diagnostics) == 0 { return }

	e.IsOverlay = true
	defer e.OverlayFalse()

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


		width := Max(50, MaxString(options))                   // width depends on max option len or 30 at min
		height := MinMany(10, len(options) + 1)                // depends on min option len or 5 at min or how many rows to the end of e.Screen
		atx := 0 + e.LINES_WIDTH + e.FilesPanelWidth; aty := 0 // Define the window  position and dimensions
		style := StyleDefault.Foreground(ColorWhite)

		var selectionEnd = false; var selected = 0; var selectedOffset = 0

		for !selectionEnd {
			if selected < selectedOffset { selectedOffset = selected }  // calculate offsets for scrolling completion
			if selected >= selectedOffset+height { selectedOffset = selected - height + 1 }

			shifty := e.DrawErrors(atx, width, aty, height, options, selectedOffset, selected, style)

			e.Screen.Show()

			switch ev := e.Screen.PollEvent().(type) { // poll and handle event
			case *EventResize:
				e.COLUMNS, e.ROWS = e.Screen.Size()
				//ROWS -= 1
				e.Screen.Sync()
				e.Screen.Clear(); e.DrawEverything(); e.Screen.Show()

			case *EventKey:
				key := ev.Key()
				if key == KeyEscape || key == KeyEnter ||
					key == KeyBackspace || key == KeyBackspace2 ||
					key == KeyCtrlE { e.Screen.Clear(); selectionEnd = true; end = true }
				if key == KeyDown { selected = Min(len(options)-1, selected+1) }
				if key == KeyUp { selected = Max(0, selected-1) }
				if key == KeyCtrlC {
					diagnostic := maybeDiagnostics.Diagnostics[selected]
					clipboard.WriteAll(diagnostic.Message)
				}
				//if key == tcell.KeyRight { e.OnRight(); e.Screen.Clear(); e.DrawEverything(); selectionEnd = true }
				if key == KeyRight {
					diagnostic := maybeDiagnostics.Diagnostics[selected]
					e.Row = int(diagnostic.Range.Start.Line)
					e.Col = int(diagnostic.Range.Start.Character)
					e.Focus();
					// add space for errors panel
					if e.Row- e.Y < shifty + height { e.Y -= shifty + height + 1}
					if e.Y < 0 { e.Y = 0 }
					e.DrawEverything(); e.Screen.Show()
				}
				//if key == tcell.KeyRune { e.AddChar(ev.Rune()); e.WriteFile(); e.Screen.Clear(); e.DrawEverything(); selectionEnd = true  }
				if key == KeyEnter {
					selectionEnd = true; end = true
					diagnostic := maybeDiagnostics.Diagnostics[selected]
					e.Row = int(diagnostic.Range.Start.Line)
					e.Col = int(diagnostic.Range.Start.Character)

					e.Selection.Ssx = e.Col; e.Selection.Ssy = e.Row;
					e.Selection.Sey = int(diagnostic.Range.End.Line)
					e.Selection.Sex = int(diagnostic.Range.End.Character)
					e.Row = e.Selection.Sey; e.Col = e.Selection.Sex
					e.Selection.IsSelected = true
					e.Focus()
					// add space for errors panel
					if e.Row- e.Y < shifty + height { e.Y -= shifty + height + 1}
					if e.Y < 0 { e.Y = 0 }
					e.DrawEverything(); e.Screen.Show()
				}
			}
		}
	}

}

func (e *Editor) DrawErrors(atx int, width int, aty int, height int, options []string,
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
			nextWord := FindNextWord(runes, j)
			if shiftx == 0 {
				e.Screen.SetContent(atx, row+aty+shifty, ' ', nil, style)
			}
			if shiftx+atx+(nextWord-j) >= e.COLUMNS {
				for k := shiftx; k <= e.COLUMNS; k++ { // Fill the remaining space
					e.Screen.SetContent(k+atx, row+aty+shifty, ' ', nil, style)
				}
				shifty++
				shiftx = 0
			}
			e.Screen.SetContent(atx+shiftx, row+aty+shifty, ch, nil, style)
			shiftx++
		}

		for col := shiftx; col < e.COLUMNS; col++ { // Fill the remaining space
			e.Screen.SetContent(col+atx, row+aty+shifty, ' ', nil, style)
		}
	}

	for col := 0; col < width; col++ { // Fill empty Line below
		e.Screen.SetContent(col+atx, height+aty+shifty-1, ' ', nil,
			StyleDefault.Background(Color(OverlayColor)))
	}

	return shifty
}

func (e *Editor) OnSearch() {
	var end = false
	if e.SearchPattern == nil { e.SearchPattern = []rune{} }
	var patternx = len(e.SearchPattern)
	var startline = e.Y
	var isChanged = true
	var isDownSearch = true
	var prefix = []rune("search: ")

	// loop until escape or enter pressed
	for !end {

		e.DrawSearch(prefix, e.SearchPattern, patternx)
		e.Screen.Show()

		if isChanged {
			var sy, sx = -1, -1
			e.X = 0
			if isDownSearch {
				sy, sx = SearchDown(e.Content, string(e.SearchPattern), startline)
			} else {
				sy, sx = SearchUp(e.Content, string(e.SearchPattern), startline)
			}

			if sx != -1 && sy != -1 {
				e.Row = sy; e.Col = sx; e.Focus()
				startline = sy;
				e.Selection.Ssx = sx; e.Selection.Ssy = sy;
				e.Selection.Sex = sx + len(e.SearchPattern); e.Selection.Sey = sy; e.Selection.IsSelected = true
				e.DrawEverything()
				e.DrawSearch(prefix, e.SearchPattern, patternx)
				e.Screen.ShowCursor(len(prefix) + patternx + e.LINES_WIDTH+ e.FilesPanelWidth, e.ROWS-1)
				e.Screen.Show()
			}else {
				e.Selection.CleanSelection()
				if isDownSearch { startline = 0 } else  { startline = len(e.Content)}
				e.DrawEverything()
				e.DrawSearch(prefix, e.SearchPattern, patternx)
				e.Screen.ShowCursor(len(prefix) + patternx + e.LINES_WIDTH+ e.FilesPanelWidth, e.ROWS-1)
				e.Screen.Show()
			}
		}

		switch ev := e.Screen.PollEvent().(type) { // poll and handle event
		case *EventResize:
			e.COLUMNS, e.ROWS = e.Screen.Size()
			//ROWS -= 1

		case *EventKey:
			isChanged = false
			key := ev.Key()

			if key == KeyRune {
				e.SearchPattern = InsertTo(e.SearchPattern, patternx, ev.Rune())
				patternx++
				isChanged = true
			}
			if key == KeyBackspace2 && patternx > 0 && len(e.SearchPattern) > 0 {
				patternx--
				e.SearchPattern = Remove(e.SearchPattern, patternx)
				isChanged = true
			}
			if key == KeyLeft && patternx > 0 { patternx-- }
			if key == KeyRight && patternx < len(e.SearchPattern) { patternx++ }
			if key == KeyDown  {
				isDownSearch = true
				if startline < len(e.Content) {
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
				if startline == 0 { startline = len(e.Content) } else { startline-- }
			}
			if key == KeyESC || key == KeyEnter || key == KeyCtrlF { end = true }
		}
	}
}
func (e *Editor) DrawSearch(prefix []rune, pattern []rune, patternx int) {
	for i := 0; i < len(prefix); i++ {
		e.Screen.SetContent(i + e.LINES_WIDTH+ e.FilesPanelWidth, e.ROWS-1, prefix[i], nil, StyleDefault)
		//e.Screen.Show()
	}

	e.Screen.SetContent(len(prefix) + e.LINES_WIDTH+ e.FilesPanelWidth, e.ROWS-1, ' ', nil, StyleDefault)
	//e.Screen.Show()

	for i := 0; i < len(pattern); i++ {
		e.Screen.SetContent(len(prefix) + i + e.LINES_WIDTH+ e.FilesPanelWidth, e.ROWS-1, pattern[i], nil, StyleDefault)
		//e.Screen.Show()
	}

	e.Screen.ShowCursor(len(prefix) + patternx + e.LINES_WIDTH+ e.FilesPanelWidth, e.ROWS-1)
	//e.Screen.Show()

	for i := len(prefix) + len(pattern) + e.LINES_WIDTH + e.FilesPanelWidth; i < e.COLUMNS; i++ {
		e.Screen.SetContent(i, e.ROWS-1, ' ', nil, StyleDefault)
		//e.Screen.Show()
	}
}

func (e *Editor) OnFiles() {
	e.IsFileSelection = true

	if e.FilesPanelWidth == 0 {
		e.ReadFilesUpdate()
		if len(e.Files) == 0 { return }
		e.FilesPanelWidth = 28
	}

	if e.Filename == "" { e.FilesPanelWidth = findMaxByFilenameLength(e.Files) + 1 }
	if e.Filename != "" { e.DrawEverything() }

	var end = false
	var filterPattern = []rune{}
	var patternx = 0
	var isChanged = true
	var shiftx = 0

	// loop until escape or enter pressed
	for !end {

		var selectionEnd = false;

		for !selectionEnd {
			if e.FileSelectedIndex != -1 && e.FileSelectedIndex < e.FileScrollingOffset {
				e.FileScrollingOffset = e.FileSelectedIndex
			}
			if e.FileSelectedIndex >= e.FileScrollingOffset+ e.ROWS {
				e.FileScrollingOffset = e.FileSelectedIndex - e.ROWS + 1
			}

			filteredFiles := e.Files
			if e.IsFilesSearch && len(filterPattern) > 0 {
				pattern := string(filterPattern)
				filteredFiles = []FileInfo{}

				for _, f := range e.Files {
					var foundMatch = false
					foundMatch = strings.Contains(f.Filename, pattern)
					if foundMatch { filteredFiles = append(filteredFiles, f) } else {
						matches, err := filepath.Match(pattern, f.Filename)
						if err != nil { continue }
						if matches { filteredFiles = append(filteredFiles, f) }
					}
				}

				if isChanged { e.DrawFiles(string(filterPattern), filteredFiles, patternx, shiftx) }
			} else {
				if isChanged { e.DrawFiles(string(filterPattern), e.Files, patternx, shiftx) }
			}

			e.Screen.Show()

			switch ev := e.Screen.PollEvent().(type) { // poll and handle event
			case *EventMouse:
				mx, my := ev.Position()
				buttons := ev.Buttons()
				modifiers := ev.Modifiers()


				if mx > e.FilesPanelWidth {
					selectionEnd = true; end = true; e.IsFilesSearch = false;
				} else {
					if buttons & WheelDown != 0  && modifiers & ModAlt != 0 && shiftx > 0  {
						shiftx--
						continue
					}
					if buttons & WheelUp != 0  && modifiers & ModAlt != 0   {
						shiftx++
						continue
					}
					if buttons & WheelDown != 0  && modifiers & ModCtrl != 0 && e.FilesPanelWidth < e.COLUMNS  {
						if e.FilesPanelWidth > findMaxByFilenameLength(e.Files) { continue }
						e.FilesPanelWidth++
						if e.Filename != "" { e.DrawEverything(); e.Screen.Show() }
						continue
					}
					if buttons & WheelUp != 0  && modifiers & ModCtrl != 0  && e.FilesPanelWidth > 0 {
						e.FilesPanelWidth--
						if e.Filename != "" { e.DrawEverything(); e.Screen.Show() }
						continue
					}

					if buttons & WheelDown != 0 &&  len(filteredFiles) > e.ROWS {
						if len(filteredFiles) > e.ROWS {
							if !e.IsFilesSearch && e.FileScrollingOffset <  len(filteredFiles) - e.ROWS {
								e.FileScrollingOffset++
							}
							if e.IsFilesSearch && e.FileScrollingOffset <  len(filteredFiles) - e.ROWS +1 {
								e.FileScrollingOffset++
							}

						}
					}
					if buttons & WheelUp != 0 && e.FileScrollingOffset > 0 {
						e.FileScrollingOffset--
					}

					if my < len(filteredFiles) { e.FileSelectedIndex = my + e.FileScrollingOffset }
					if buttons & Button1 == 1 {
						e.ReadFilesUpdate()
						e.FileSelectedIndex = my + e.FileScrollingOffset
						if e.FileSelectedIndex < 0  { continue }
						if e.FileSelectedIndex >= len(filteredFiles) { continue }
						if mx > len(filteredFiles[e.FileSelectedIndex].Filename) { continue}
						selectionEnd = true; end = true
						selectedFile := filteredFiles[e.FileSelectedIndex]
						e.InputFile = selectedFile.FullFilename
						e.OpenFile(e.InputFile)
						e.IsFilesSearch = false
					}
				}

			case *EventResize:
				e.COLUMNS, e.ROWS = e.Screen.Size()
				if e.Filename != "" { e.DrawEverything(); e.Screen.Show() }

			case *EventKey:
				key := ev.Key()

				if key == KeyCtrlF { e.IsFilesSearch = !e.IsFilesSearch }
				if key == KeyEscape && !e.IsFilesSearch { selectionEnd = true; end = true; e.FilesPanelWidth =  0 }
				if key == KeyEscape  && e.IsFilesSearch {  e.IsFilesSearch = false}
				if key == KeyDown { e.FileSelectedIndex = Min(len(filteredFiles)-1, e.FileSelectedIndex+1) }
				if key == KeyUp { e.FileSelectedIndex = Max(0, e.FileSelectedIndex-1) }
				if key == KeyRune {
					e.IsFilesSearch = true
					filterPattern = InsertTo(filterPattern, patternx, ev.Rune())
					patternx++
					isChanged = true
					e.FileSelectedIndex = 0
				}
				if key == KeyBackspace2  && e.IsFilesSearch && patternx > 0 && len(filterPattern) > 0 {
					patternx--
					filterPattern = Remove(filterPattern, patternx)
					isChanged = true
				}
				if key == KeyLeft && e.IsFilesSearch && patternx > 0 { patternx--; isChanged = true }
				if key == KeyRight && e.IsFilesSearch && patternx < len(filterPattern) { patternx++; isChanged = true }
				if key == KeyRight && !e.IsFilesSearch {
					isChanged = true
					e.FilesPanelWidth++
					if e.Filename != "" { e.DrawEverything(); e.Screen.Show() }
				}
				if key == KeyLeft && !e.IsFilesSearch && e.FilesPanelWidth > 0  {
					isChanged = true

					e.FilesPanelWidth--
					//if e.Filename != "" { e.DrawEverything(); e.Screen.Show() }
					e.Screen.Clear(); e.DrawEverything(); e.Screen.Show()
				}
				if key == KeyCtrlT {
					selectionEnd = true; end = true
					e.IsFilesSearch = false
					e.FilesPanelWidth = 0
				}
				if key == KeyEnter && e.FileSelectedIndex < len(filteredFiles) {
					selectionEnd = true; end = true
					e.IsFilesSearch = false
					selectedFile := filteredFiles[e.FileSelectedIndex]
					e.InputFile = selectedFile.FullFilename
					e.OpenFile(e.InputFile)
				}
			}
		}
	}

	e.IsFileSelection = false
}

func (e *Editor) DrawFiles(pattern string, files []FileInfo, patternx int, shiftx int) {

	for row := 0; row < e.ROWS; row++ {
		for col := 0; col < e.FilesPanelWidth; col++ { // clean
			e.Screen.SetContent(col, row, ' ', nil, StyleDefault)
		}
	}

	var offsety = 0

	for fileIndex := 0; fileIndex < len(files); fileIndex++ {
		if e.IsFilesSearch && offsety == e.ROWS-1 { continue }
		style := StyleDefault

		e.Screen.SetContent(e.FilesPanelWidth, offsety, ' ', nil, style)

		if fileIndex >= len(files) || fileIndex >= e.ROWS { break }
		if fileIndex + Max(e.FileScrollingOffset,0) >= len(files) { break }
		file := files[fileIndex + Max(e.FileScrollingOffset,0)]


		isSelectedFile := e.IsFileSelection && e.FileSelectedIndex != -1 && fileIndex + e.FileScrollingOffset == e.FileSelectedIndex
		if isSelectedFile {
			style = style.Foreground(Color(AccentColor))
		}
		if e.InputFile != "" && e.InputFile == file.FullFilename {
			style = style.Background(Color(AccentColor)).Foreground(ColorWhite)
		}

		//filename := file.Filename
		//dir, f := filepath.Split(filename)
		//if strings.HasSuffix(dir, "/") { dir = dir[:len(dir)-1] }

		for i := 0; i < len(file.Filename); i++ {
			if i+1 > e.FilesPanelWidth-1 { break }
			if i+shiftx +1 > len(file.Filename) { break }
			e.Screen.SetContent(i + 1, offsety, rune(file.Filename[i+shiftx]), nil, style)
		}
		//for j, ch := range file.Filename {
		//	if shiftx >= j { continue }
		//	if j+1 > e.FilesPanelWidth-1 { break }
		//	e.Screen.SetContent(j + 1, offsety, ch, nil, style)
		//}

		offsety++
	}

	for row := 0; row <= e.ROWS; row++ {
		if row >= len(files) {
			for col := 0; col < e.FilesPanelWidth; col++ { // clean
				e.Screen.SetContent(col, row, ' ', nil, StyleDefault)
			}
		}
		//e.Screen.SetContent(FilesPanelWidth, row, '│', nil, StyleDefault.Foreground(Color(AccentColor)))
	}

	e.Screen.HideCursor()

	if e.IsFilesSearch {
		pref := " search: "
		e.Screen.ShowCursor(len(pref) + patternx, e.ROWS-1)
		for i, ch := range pref { // draw prefix
			e.Screen.SetContent(i, e.ROWS-1, ch, nil, StyleDefault)
		}

		for i, ch := range pattern { // draw pattern
			e.Screen.SetContent(i+len(pref), e.ROWS-1, ch, nil, StyleDefault)
		}
		for col := len(pref) + len(pattern); col < e.FilesPanelWidth- 1; col++ { // clean
			e.Screen.SetContent(col, e.ROWS-1, ' ', nil, StyleDefault)
		}
	}

}

//func (e *Editor) addTab() {
//	if e.filesInfo == nil || len(e.filesInfo) == 0 {
//		e.filesInfo = append(e.filesInfo, FileInfo{e.Filename, e.AbsoluteFilePath, 1})
//	} else {
//		var tabExists = false
//
//		for i := 0; i < len(e.filesInfo); i++ {
//			ti := e.filesInfo[i]
//			if e.AbsoluteFilePath == ti.FullFilename {
//				ti.OpenCount += 1
//				e.filesInfo[i] = ti
//				tabExists = true
//			}
//		}
//
//		if !tabExists {
//			e.filesInfo = append(e.filesInfo, FileInfo{e.Filename, e.AbsoluteFilePath, 1})
//		}
//
//		sort.SliceStable(e.filesInfo, func(i, j int) bool {
//			return e.filesInfo[i].OpenCount < e.filesInfo[j].OpenCount
//		})
//	}
//}
//
//func (e *Editor) drawTabs() {
//	e.COLUMNS, e.ROWS = e.Screen.Size()
//
//	if len(e.filesInfo) == 0 { return }
//	if e.FilesPanelWidth == 0 { return }
//	if e.FilesPanelWidth == 0 { return }
//
//	e.ROWS -= 1
//	at := e.ROWS
//	fromx := 1
//	styleDefault := StyleDefault
//
//	for i := fromx; i < e.COLUMNS; i++ {
//		e.Screen.SetContent(0, at, ' ', nil, styleDefault)
//	}
//
//	xpos := 0
//	for _, info := range e.filesInfo {
//		if xpos > e.COLUMNS { break }
//		for _, ch := range info.Filename {
//			e.Screen.SetContent(xpos + fromx, at, ch, nil, styleDefault)
//			xpos++
//		}
//
//		e.Screen.SetContent(xpos + fromx, at, ' ', nil, styleDefault)
//		xpos++
//		e.Screen.SetContent(xpos + fromx, at, ' ', nil, styleDefault)
//	}
//}

func (e *Editor) OverlayFalse() {
	e.IsOverlay = false
}

func (e *Editor) UpdateColors() {
	if !e.IsColorize { return }
	if e.Lang == "" { return }
	if len(e.Content) >= 10000 {
		line := string(e.Content[e.Row])
		linecolors := HighlighterGlobal.Colorize(line, e.Filename)
		e.Colors[e.Row] = linecolors[0]
	} else {
		code := ConvertContentToString(e.Content)
		e.Colors = HighlighterGlobal.Colorize(code, e.Filename)
	}
}

func (e *Editor) UpdateColorsFull() {
	if !e.IsColorize { return }
	if e.Lang == "" { return }

	code := ConvertContentToString(e.Content)
	e.Colors = HighlighterGlobal.Colorize(code, e.Filename)
}

func (e *Editor) UpdateColorsAtLine(at int) {
	if !e.IsColorize { return }
	if e.Lang == "" { return }
	if at >= len(e.Colors) { return }

	line := string(e.Content[at])
	if line == "" { e.Colors[at] = []int{}; return }
	linecolors := HighlighterGlobal.Colorize(line, e.Filename)
	e.Colors[at] = linecolors[0]
}

// todo, get rid of this function, cause UpdateColors is slow for big files
func (e *Editor) UpdateNeeded() {
	e.Update = true
	e.IsContentChanged = true
	if len(e.Content) <= 10000 { go e.WriteFile() }
	e.UpdateColors()
}