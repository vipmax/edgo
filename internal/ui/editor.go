package ui
// 

import (
	. "edgo/internal/config"
	"edgo/internal/dap"
	. "edgo/internal/highlighter"
	. "edgo/internal/io"
	. "edgo/internal/logger"
	. "edgo/internal/lsp"
	. "edgo/internal/operations"
	. "edgo/internal/process"
	. "edgo/internal/search"
	. "edgo/internal/selection"
	. "edgo/internal/tests"
	. "edgo/internal/utils"
	"fmt"
	"github.com/atotto/clipboard"
	. "github.com/gdamore/tcell"
	"github.com/gdamore/tcell/encoding"
	"github.com/rjeczalik/notify"
	"log"
	"os"
	"os/user"
	"path"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)


type Editor struct {
	COLUMNS     int // terminal size columns
	ROWS        int // terminal size rows
	LINES_WIDTH int // draw file lines number

	TERMINAL_HEIGHT int
	TERMINAL_WIDHT  int

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

	Cwd              string // current dir
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
	Tree                FileInfo   // files Tree
	FilesSearchPattern  []rune

	IsContentSearch   bool
	SearchPattern     []rune // pattern for search in a buffer
	SearchResults     []SearchResult
	SearchResultIndex int

	//filesInfo []FileInfo
	CursorHistory     []CursorMove
	CursorHistoryUndo []CursorMove

	//LastCommitFileContent string
	//Added                 Set
	//Removed               Set

	// process panel vars
	ProcessPanelHeight    int
	ProcessPanelWidth     int
	ProcessContent        [][]rune
	ProcessPanelScroll            int
	ProcessPanelHScroll           int
	IsProcessPanelMoving          bool
	IsProcessPanelFocused         bool
	Process                       *Process
	ProcessPanelSpacing           int
	ProcessPanelCursorX           int
	ProcessPanelCursorY           int
	ProcessPanelSelection         Selection
	IsProcessPanelSearch          bool
	ProcessPanelSearchPattern     []rune // pattern for search in a buffer
	ProcessPanelSearchResults     []SearchResult
	ProcessPanelSearchResultIndex int

	lsp2lang map[string]*LspClient

	Dap       dap.DapClient
	DebugInfo DebugInfo

	treeSitterHighlighter *TreeSitterHighlighter

	FileWatcher *FileWatcher
	DirWatcher  *DirWatcher

	Tests     map[int]TestData
	TestFinder TestFinder
	Test       Test

	TreePath *Path

	HighlightElements map[int][]NodeRange
}

func (e *Editor) Start() {
	Log.Info("starting edgo")

	e.Init()

	cwd, _ := os.Getwd()
	e.Cwd = cwd

	// reading file from cmd args
	if len(os.Args) == 1 {
		// if no args, open current dir
		e.DrawLogo()
		e.OnFilesTree(false)
	} else {
		e.Filename = os.Args[1]
		e.InputFile = e.Filename

		info, err := os.Stat(e.InputFile)
		//if err != nil { log.Fatal(err); return }

		if info != nil && info.IsDir() {
			// if arg is dir, go to dir and open
			err = os.Chdir(e.InputFile)
			if err != nil { log.Fatal(err) }
			e.OnFilesTree(true)
		} else {
			// if arg is file, open file
			err := e.OpenFile(e.InputFile)
			if err != nil { log.Fatal(err) }
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

func (e *Editor) Exit() {
	e.Screen.Fini()
}

func (e *Editor) HandleEvents() {
	//e.Update = false
	e.Update = true
	ev := e.Screen.PollEvent()
	switch ev := ev.(type) {
	case *EventResize:
		e.COLUMNS, e.ROWS = e.Screen.Size()
		e.TERMINAL_HEIGHT = e.ROWS
		e.TERMINAL_WIDHT = e.COLUMNS
		e.ROWS -= e.ProcessPanelHeight
		e.DrawEverything()
		e.Screen.Show()

	case *EventMouse:
		mx, my := ev.Position()
		buttons := ev.Buttons()
		modifiers := ev.Modifiers()

		e.HandleMouse(mx, my, buttons, modifiers)

	case *EventKey:
		key := ev.Key()
		modifiers := ev.Modifiers()
		if e.Dap.IsStarted { e.OnDebugKeyHandle(key, ev, 1); return }

		e.HandleKeyboard(key, ev, modifiers)
	}
}

func (e *Editor) HandleMouse(mx int, my int, buttons ButtonMask, modifiers ModMask) {
	_, screenRows := e.Screen.Size()

	// upper play button
	if mx == e.COLUMNS - 2 && my == 0  && buttons & Button1 == 1 {
		// do not show if process active
		if e.Process == nil || (e.Process != nil && e.Process.IsStopped()){
			e.OnProcessRun(true)
		}
		return
	}

	// play button on process panel
	if mx == e.COLUMNS - 6 && my == e.ROWS  && buttons & Button1 == 1 {
		e.OnProcessStop()
		e.OnProcessRun(false)
		return
	}

	// stop button on process panel
	if mx == e.COLUMNS - 4 && my == e.ROWS  && buttons & Button1 == 1 {
		e.OnProcessStop()
		return
	}

	// close button on process panel
	if mx == e.COLUMNS - 2 && my == e.ROWS && buttons & Button1 == 1 {
		e.OnProcessStop()
		e.ROWS = screenRows
		e.ProcessPanelHeight = 0
		e.ProcessContent = [][]rune{}
		return
	}

	// detect process panel drag start event
	if !e.IsProcessPanelMoving && buttons & Button1 == 1 &&
		my == e.ROWS && e.ProcessPanelHeight > 0 &&
		len(e.ProcessPanelSelection.GetSelectedLines(e.ProcessContent)) == 0 {

		e.ROWS = my
		e.ProcessPanelHeight = screenRows - e.ROWS
		e.IsProcessPanelMoving = true
		e.Update = true
		return
	}
	// detect process panel dragging event
	if e.IsProcessPanelMoving && buttons & Button1 == 1 && screenRows >= my  {
		e.ROWS = my
		e.ProcessPanelHeight = screenRows - e.ROWS
		e.Update = true
		return
	}
	// detect process panel drag stop event
	if e.IsProcessPanelMoving && buttons & Button1 == 0 {
		e.IsProcessPanelMoving = false; return
	}

	if my >= e.ROWS {
		if mx < 2 || mx > e.COLUMNS { return }
		// in process panel
		e.IsProcessPanelFocused = true
		if buttons & WheelDown != 0 && e.ProcessPanelScroll <= len(e.ProcessContent) - e.ProcessPanelHeight {
			e.ProcessPanelScroll++
		}
		if buttons & WheelUp != 0 && e.ProcessPanelScroll > 0 {
			e.ProcessPanelScroll--
		}

		if buttons & Button1 == 1 {
			if mx < e.ProcessPanelSpacing { return }
			e.ProcessPanelCursorX = mx + e.ProcessPanelHScroll - e.ProcessPanelSpacing
			e.ProcessPanelCursorY = my + e.ProcessPanelScroll - e.ROWS -1

			if e.ProcessPanelCursorY < 0 { e.ProcessPanelCursorY = 0 }
			// fit cursor
			if e.ProcessPanelCursorY >= len(e.ProcessContent) { e.ProcessPanelCursorY = len(e.ProcessContent)-1 }
			if e.ProcessPanelCursorY < len(e.ProcessContent) && e.ProcessPanelCursorX > len(e.ProcessContent[e.ProcessPanelCursorY]) {
				e.ProcessPanelCursorX = len(e.ProcessContent[e.ProcessPanelCursorY])
			}

			if e.ProcessPanelSelection.Ssx < 0 {
				e.ProcessPanelSelection.Ssx, e.ProcessPanelSelection.Ssy =
					e.ProcessPanelCursorX, e.ProcessPanelCursorY }
			if e.ProcessPanelSelection.Ssx >= 0 {
				e.ProcessPanelSelection.Sex, e.ProcessPanelSelection.Sey =
					e.ProcessPanelCursorX, e.ProcessPanelCursorY }
			return
		}

		if buttons & Button1 == 0 {
			if e.ProcessPanelSelection.IsSelectionNonEmpty() {
				selectionString := e.ProcessPanelSelection.GetSelectionString(e.ProcessContent)
				clipboard.WriteAll(selectionString)
			}

			e.ProcessPanelSelection.CleanSelection()
		}

		return
	}

	e.IsProcessPanelFocused = false


	if !e.IsFilesPanelMoving && buttons & Button1 == 1 &&
		(mx == e.FilesPanelWidth-2 || mx == e.FilesPanelWidth-1) &&
			my < e.ROWS && len(e.Selection.GetSelectedLines(e.Content)) == 0 {
			e.IsFilesPanelMoving = true
			return
	}

	if e.IsFilesPanelMoving && buttons & Button1 == 1 { e.FilesPanelWidth = mx; return }
	if e.IsFilesPanelMoving && buttons & Button1 == 0 { e.IsFilesPanelMoving = false; return }
	if mx < e.FilesPanelWidth-3 && buttons & Button1 == 0 && !e.Dap.IsStarted { e.OnFilesTree(true); return }

	if e.Filename == "" { return }

	if buttons & Button1 == 1 && mx == e.COLUMNS - 2 { // test button
		line := my + e.Y
		if  _, found := e.Tests[line]; found {
			e.RunTest(e.Tests[line])
			return
		}

	}

	mx -= e.LINES_WIDTH + e.FilesPanelWidth

	if mx < 0 { return }
	if my > e.ROWS { return }

	if e.Content == nil { return }



	// if click with control or option, lookup for definition or references
	if buttons & Button1 == 1 && (modifiers & ModAlt != 0 || modifiers & ModCtrl != 0) {
		e.Row = my + e.Y
		if e.Row > len(e.Content)-1 { e.Row = len(e.Content) - 1 } // fit cursor to e.Content

		e.Col = e.FindCursorXPosition(mx)

		if len(e.Selection.GetSelectedLines(e.Content)) > 0 { // if text selected
			e.Selection.Sey = e.Row
			e.Selection.Sex = e.Col
			return
		}
		if modifiers & ModAlt != 0 { e.OnReferences() }
		if modifiers & ModCtrl != 0 { e.OnDefinition() }
		return
	}
	prevRow := e.Row

	if e.Selection.IsSelected && buttons & Button1 == 1 {
		e.Update = true
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
		e.Update = true

		e.Row = my + e.Y
		if e.Row > len(e.Content)-1 { e.Row = len(e.Content) - 1 } // fit cursor to e.Content

		xPosition := e.FindCursorXPosition(mx)

		if prevRow == e.Row && e.Col == xPosition && len(e.Selection.GetSelectedLines(e.Content)) == 0 {
			// double click
			lastChar := len(e.Content[e.Row]) == e.Col
			if lastChar {
				e.Selection.Ssx, e.Selection.Ssy = e.Col, e.Row
				e.Selection.Sex, e.Selection.Sey = e.Col, e.Row
				return
			}
			prw := FindPrevWord(e.Content[e.Row], e.Col)
			nxw := FindNextWord(e.Content[e.Row], e.Col)
			e.Selection.Ssx, e.Selection.Ssy = prw, e.Row
			e.Selection.Sex, e.Selection.Sey = nxw, e.Row
			e.Col = nxw
			return
		}

		e.Col = xPosition
		e.OnCursorChanged()
		e.CursorHistory = append(e.CursorHistory,
			CursorMove{e.AbsoluteFilePath, e.Row, e.Col, e.Y, e.X},
		)

		if e.Col < 0 { e.Col = 0 }
		if e.Selection.Ssx < 0 { e.Selection.Ssx, e.Selection.Ssy = e.Col, e.Row }
		if e.Selection.Ssx >= 0 { e.Selection.Sex, e.Selection.Sey = e.Col, e.Row }
	}

	if buttons & Button1 == 0 {
		if e.Selection.Ssx != -1 && e.Selection.Sex != -1 {
			e.Selection.IsSelected = true
			e.Update = true
		}
	}
	return
}

func (e *Editor) HandleKeyboard(key Key, ev *EventKey, modifiers ModMask) {
	if key == KeyCtrlF && !e.IsProcessPanelFocused { e.OnSearch() }

	if e.Filename == "" && key != KeyCtrlQ { return }

	if e.IsProcessPanelFocused {
		e.OnProcessKeyHandle(key, ev.Rune())
		return
	}

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
	if key == KeyCtrlX { e.Cut(true) }
	if key == KeyCtrlD { e.Duplicate() }
	if key == KeyCtrlB { e.Breakpoint() }
	if key == KeyCtrlK { e.GoBottom(); return }
	if key == KeyCtrlJ { e.GoTop(); return }
	if key == KeyCtrlL { e.GoToLine(); return }

	if modifiers & ModShift != 0 && (
		key == KeyRight ||
			key == KeyLeft ||
			key == KeyUp ||
			key == KeyDown) {

		if e.Selection.Ssx < 0 { e.Selection.Ssx, e.Selection.Ssy = e.Col, e.Row }
		if key == KeyRight { e.OnRight() }
		if key == KeyLeft { e.OnLeft() }
		if key == KeyUp { e.OnUp() }
		if key == KeyDown { e.OnDown() }
		if e.Selection.Ssx >= 0 {
			e.Selection.Sex, e.Selection.Sey = e.Col, e.Row; e.Selection.IsSelected = true
		}
		return
	}

	if key == KeyRune && modifiers & ModAlt != 0 && len(e.Content) > 0 {
		e.HandleSmartMove(ev.Rune());
		return
	}
	if key == KeyDown && modifiers & ModAlt != 0 {
		e.OnSelectLessAtCursor()
		//e.HandleSmartMoveDown();
		return
	}
	if key == KeyUp && modifiers & ModAlt != 0 {
		e.OnSelectMoreAtCursor()
		//e.HandleSmartMoveUp();
		return
	}

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
	if key == KeyEnter { e.OnEnter(); return }
	if key == KeyBackspace || key == KeyBackspace2 { e.OnDelete() }
	if key == KeyDown { e.OnDown(); e.Selection.CleanSelection() }
	if key == KeyUp { e.OnUp(); e.Selection.CleanSelection() }
	if key == KeyLeft { e.OnLeft(); e.Selection.CleanSelection() }
	if key == KeyRight { e.OnRight(); e.Selection.CleanSelection() }
	if key == KeyCtrlT { e.OnFilesTree(true) }
	if key == KeyF18 { e.OnRename() }
	if key == KeyF22 { e.OnProcessRun(true) }
	if key == KeyF23 { e.OnDebug() }
	if key == KeyCtrlU { e.OnUndo() }
	//if key == KeyCtrlR { e.OnRedo() } // todo: fix it
	if key == KeyCtrlO { e.OnCursorBack() }
	if key == KeyCtrlRightSq { e.OnCursorBackUndo() }

	if key == KeyCtrlSpace { e.OnCompletion() }

}

func (e *Editor) OpenFile(fname string) error {
	if strings.HasPrefix(fname, "~/") {
		currentUser, _ := user.Current()
		homeDir := currentUser.HomeDir
		fname = filepath.Join(homeDir, fname[2:])
	}

	absoluteDir, err := filepath.Abs(path.Dir(fname))
	if err != nil { return err }
	//directory := absoluteDir;
	e.Filename = filepath.Base(fname)
	e.AbsoluteFilePath = path.Join(absoluteDir, e.Filename)

	Log.Info("open", e.AbsoluteFilePath)

	newLang := DetectLang(e.AbsoluteFilePath)
	Log.Info("new lang is", newLang)

	if newLang != "" && newLang != e.Lang {
		e.Lang = newLang

		_, found := e.lsp2lang[newLang]
		if !found {
			lsp := LspClient{Lang: newLang}
			e.lsp2lang[newLang] = &lsp
			go e.InitLsp(e.Lang)
		}

		if e.Dap.Port > 0 {
			e.Dap = dap.DapClient{Lang: newLang, Conntype: "tcp", Port: e.Dap.Port +1}
		} else {
			e.Dap = dap.DapClient{Lang: newLang, Conntype: "tcp", Port: 54752}
		}

		e.DebugInfo = DebugInfo{stopline: -1}
	}

	conf, found := e.Config.Langs[e.Lang]
	if !found { conf = DefaultLangConfig }
	e.langConf = conf
	e.langTabWidth = conf.TabWidth

	code := e.ReadFile(e.AbsoluteFilePath)
	//e.Colors = HighlighterGlobal.Colorize(code, e.Filename)
	e.treeSitterHighlighter = NewTreeSitter()
	e.treeSitterHighlighter.SetTheme(e.Config.Theme)
	e.treeSitterHighlighter.SetLang(e.Lang)
	e.Colors = e.treeSitterHighlighter.Colorize(code)
	clear(e.HighlightElements)

	//cwd, _ := os.Getwd()
	//relativePath, _ := filepath.Rel(cwd, e.AbsoluteFilePath)
	//
	//lastCommitFileContent, err := GetLastCommitFileContent(relativePath)
	//if err != nil { e.LastCommitFileContent = "" } else  {
	//	e.LastCommitFileContent = lastCommitFileContent
	//	added, removed := Diff(lastCommitFileContent, ConvertContentToString(e.Content))
	//	e.Added = added
	//	e.Removed = removed
	//}


	e.Undo = []EditOperation{}
	e.Redo = []EditOperation{}

	e.UpdateFilesOpenStats(fname)

    e.Row = 0; e.Col = 0; e.Y = 0; e.X = 0
	e.Selection = Selection{-1,-1,-1,-1,false }
	e.SearchResults = []SearchResult{}

	e.FileWatcher.UpdateFile(e.AbsoluteFilePath)
	e.FileWatcher.UpdateStats()

	e.FindTests()

	return nil
}

func (e *Editor) Init() {
	encoding.Register()
	screen, err := NewScreen()
	if err != nil { fmt.Fprintf(os.Stderr, "%v\n", err); os.Exit(1) }
	e.Screen = screen

	err2 := e.Screen.Init()
	if err2 != nil { fmt.Fprintf(os.Stderr, "%v\n", err2); os.Exit(1) }

	e.Screen.EnableMouse()
	e.Screen.Clear()

	e.COLUMNS, e.ROWS = e.Screen.Size()
	e.TERMINAL_HEIGHT = e.ROWS
	e.TERMINAL_WIDHT = e.COLUMNS
	e.LINES_WIDTH = 6
	e.Update = true
	e.IsColorize = true
	e.FileSelectedIndex = -1
	e.CursorHistory = []CursorMove{}
	e.lsp2lang = map[string]*LspClient{}
	e.DebugInfo = DebugInfo{}

	e.treeSitterHighlighter = NewTreeSitter()
	e.treeSitterHighlighter.SetTheme(e.Config.Theme)

	e.FileWatcher = NewFileWatcher(1000)
	e.FileWatcher.StartWatch(e.OnFileUpdate)

	e.DirWatcher = NewDirWatcher(".")
	e.DirWatcher.StartWatch(e.OnFilesTreeUpdate)

	e.Tests = make(map[int]TestData)
	e.TestFinder = TestFinder{}

	return
}



func (e *Editor) DrawEverything() {
	e.Screen.Clear()

	if e.FilesPanelWidth != 0 {
		// clean  files panel and draw separator
		//_, screenRows := e.Screen.Size()
		for row := 0; row < e.ROWS; row++ {
			for col := 0; col < e.FilesPanelWidth; col++ { // clean
				e.Screen.SetContent(col, row, ' ', nil, StyleDefault)
			}
			e.Screen.SetContent(e.FilesPanelWidth-2, row, '▕', nil, DimmedStyle)
		}

		var aty = 0
		var fileindex = 0
		e.DrawTree(e.Tree, 0, &fileindex, &aty)
	}

	if len(e.Content) == 0 {
		e.DrawLogo()
		return
	}

	countTabsTo := CountTabsTo(e.Content[e.Row], e.Col)
	tabcor := countTabsTo *(e.langTabWidth - 1)

	if e.Col < e.X { e.X = e.Col }
	if e.Col + e.LINES_WIDTH + e.FilesPanelWidth + tabcor >= e.X + e.COLUMNS  {
		e.X = e.Col - e.COLUMNS + 1 + e.LINES_WIDTH + e.FilesPanelWidth + tabcor
	}

	// draw Line number and chars according to scrolling offsets
	for row := 0; row < e.ROWS; row++ {
		ry := row + e.Y // index to get right row in characters buffer by scrolling offset Y
		if row >= len(e.Content) || ry >= len(e.Content) { break }
		e.DrawLineNumber(ry, row)

		if _, found := e.Tests[ry]; found {
			e.DrawTest(ry, row)
		}

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

		if helements, found := e.HighlightElements[ry]; found {
			for _, helement := range helements {
				tabs := CountTabsTo(e.Content[helement.Ssy], helement.Ssx)
				tabcorrection := tabs * (e.langTabWidth - 1)
				skip := false
				for i := helement.Ssx; !skip && i < helement.Sex; i++ {
					x := i + e.LINES_WIDTH + e.FilesPanelWidth + tabcorrection
					mainc, _, stylec, _ := e.Screen.GetContent(x, row)
					if e.Selection.IsUnderSelection(i, ry) { skip = true } else {
						e.Screen.SetContent(x, row , mainc, nil, stylec.Background(Color(HighlightColor)))
					}
				}
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

	if e.Row - e.Y >= e.ROWS { e.Screen.HideCursor() }

	e.DrawProcessPanel()

	if e.IsContentSearch {
		e.DrawSearch(e.SearchPattern, len(e.SearchPattern))
	}
	if e.IsFilesSearch && !e.Dap.IsStarted {
		e.DrawTreeSearch(e.FilesSearchPattern, len(e.FilesSearchPattern))
	}
	if e.Dap.IsStarted {
		e.DrawDebugPanel()
	}
}

func (e *Editor) CleanProcessPanel() {
	for j := e.ROWS; j < e.TERMINAL_HEIGHT; j++ {
		for i := 0; i < e.COLUMNS; i++ {
			e.Screen.SetContent(i, j, ' ',nil, StyleDefault)
		}
	}
}

func (e *Editor) DrawProcessPanel() {
	if e.langConf.Cmd != "" && (e.Process == nil || e.Process != nil && e.Process.IsStopped()) {
		e.Screen.SetContent(e.COLUMNS-2, 0, '▶', nil, StyleDefault.Foreground(Color(HighlighterGlobal.GetRunButtonStyle())))
	}

	for i := 0; i < e.COLUMNS-7; i++ {
		e.Screen.SetContent(i, e.ROWS, '─', nil, SeparatorStyle)
	}

	e.Screen.SetContent(e.COLUMNS-7, e.ROWS, ' ',nil, StyleDefault)

	e.Screen.SetContent(e.COLUMNS-6, e.ROWS, '▶', nil, StyleDefault.Foreground(Color(HighlighterGlobal.GetRunButtonStyle())))

	e.Screen.SetContent(e.COLUMNS-5, e.ROWS, ' ',nil, StyleDefault)

	if e.Process != nil && e.Process.IsStopped() {
		e.Screen.SetContent(e.COLUMNS-4, e.ROWS, ' ',nil, StyleDefault)
	} else {
		e.Screen.SetContent(e.COLUMNS-4, e.ROWS, '■',nil, StyleDefault.Foreground(Color(AccentColor)))
	}
	e.Screen.SetContent(e.COLUMNS-3, e.ROWS, ' ',nil, StyleDefault)
	e.Screen.SetContent(e.COLUMNS-2, e.ROWS, 'x',nil, StyleDefault)

	screenCols, screenRows := e.Screen.Size()

	if e.ProcessPanelCursorX < e.ProcessPanelHScroll {
		e.ProcessPanelHScroll = e.ProcessPanelCursorX
	}
	if e.ProcessPanelCursorX >= e.ProcessPanelHScroll + screenCols - e.ProcessPanelSpacing  {
		e.ProcessPanelHScroll = e.ProcessPanelCursorX - screenCols + e.ProcessPanelSpacing +1
	}

	for index := 0; index < len(e.ProcessContent); index++ {
		if index + e.ProcessPanelScroll > len(e.ProcessContent) - 1 { break }
		line := e.ProcessContent[index + e.ProcessPanelScroll]
		y := e.ROWS + index + 1
		if y > screenRows { break }

		for col := 0; col <= e.COLUMNS; col++ {
			cx := col + e.ProcessPanelHScroll // index to get right column in characters buffer by scrolling offset x
			if cx >= len(line) { break }
			ch := line[cx]

			style := StyleDefault
			style = style.Foreground(Color(AccentColor3))
			if e.ProcessPanelSelection.IsUnderSelection(cx, index + e.ProcessPanelScroll ) {
				style = style.Background(Color(SelectionColor))
			}

			e.Screen.SetContent(col + e.ProcessPanelSpacing, y, ch,nil, style)
		}

		for i := len(line); i < e.COLUMNS; i++ {
			e.Screen.SetContent(i + e.ProcessPanelSpacing, y, ' ',nil, StyleDefault)
		}
	}

	if e.IsProcessPanelFocused {
		if e.ProcessPanelCursorY - e.ProcessPanelScroll + e.ROWS +1 <= e.ROWS {
			e.Screen.HideCursor()
		} else {
			e.Screen.ShowCursor(e.ProcessPanelCursorX + e.ProcessPanelSpacing - e.ProcessPanelHScroll, e.ProcessPanelCursorY - e.ProcessPanelScroll + e.ROWS + 1)
		}

	} else {
		//e.Screen.HideCursor()
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
	if e.DebugInfo.stopline == ry {
		style = style.Background(Color(SelectionColor))
	}
	return style
}

func (e *Editor) DrawDiagnostic() {
	if e.Lang == "" { return }
	lsp := e.lsp2lang[e.Lang]
	if !lsp.IsReady { return }

	maybeDiagnostic, found := lsp.GetDiagnostic("file://" + e.AbsoluteFilePath)

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
	//if e.Added.Contains(brw+1) {
	//	style = StyleDefault.Foreground(Color(AccentColor))
	//}

	bps, found := e.Dap.Breakpoints[e.AbsoluteFilePath]
	if found && Contains(bps, brw+1){
		style = StyleDefault.Foreground(Color(AccentColor))
		e.Screen.SetContent(e.FilesPanelWidth + e.LINES_WIDTH/2 -1, row, '●', nil, style)
		return
	}


	lineNumber := CenterNumber(brw + 1, e.LINES_WIDTH)
	for index, char := range lineNumber {
		e.Screen.SetContent(index + e.FilesPanelWidth, row, char, nil, style)
	}
}

func (e *Editor) DrawStatus(text string) {
	//var style = StyleDefault
	var style = StyleDefault.Foreground(ColorDimGray)
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

	lsp := e.lsp2lang[lang]
	split := strings.Split(conf.Lsp, " ")
	started := lsp.Start(split[0], split[1:]...)
	if !started { return }

	//var diagnosticUpdateChan = make(chan string)
	//go lsp.ReceiveLoop(diagnosticUpdateChan, lang)

	currentDir, _ := os.Getwd()

	lsp.Init(currentDir)
	lsp.DidOpen(e.AbsoluteFilePath, ConvertContentToString(e.Content))

	//e.DrawEverything()
	//
	//lspStatus := "lsp started, elapsed " + time.Since(Start).String()
	//if !lsp.isReady { lspStatus = "lsp is not ready yet" }
	//Log.Info("lsp status", lspStatus)
	//status := fmt.Sprintf(" %e.Screen %e.Screen %d %d %e.Screen ", lspStatus,  lang, Row+1, Col+1, InputFile)
	//e.drawText(COLUMNS- len(status), ROWS-1, COLUMNS, ROWS-1, status)
	//e.Screen.Show()

	go func() {
		for range lsp.DiagnosticsChannel {
			if e.IsOverlay { continue }
			e.DrawEverything()
			e.Screen.Show()
		}
	}()
}

func (e *Editor) OnErrors() {
	e.Screen.HideCursor()
	defer e.Screen.ShowCursor(e.Col, e.Row)

	if e.Lang == "" { return }
	lsp := e.lsp2lang[e.Lang]
	if !lsp.IsReady { return }

	maybeDiagnostics, found := lsp.GetDiagnostic("file://" + e.AbsoluteFilePath)

	if !found || len(maybeDiagnostics.Diagnostics) == 0 { return }

	e.IsOverlay = true
	defer e.OverlayFalse()

	var end = false

	// loop until escape or enter pressed
	for !end {

		var options = []string{}
		for i, diagnistic := range maybeDiagnostics.Diagnostics {
			text := fmt.Sprintf("%d/%d [%d:%d] %s ", i + 1, len(maybeDiagnostics.Diagnostics),
				int(diagnistic.Range.Start.Line) + 1, int(diagnistic.Range.Start.Character + 1),
				diagnistic.Message,
			)
			options = append(options, text)
		}


		width := Max(50, MaxString(options))                   // width depends on max option len or 30 at min
		height := MinMany(10, len(options))                // depends on min option len or 5 at min or how many rows to the end of e.Screen
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

	for col := 0; col < e.COLUMNS; col++ { // Fill empty Line below
		e.Screen.SetContent(col+atx, height+aty+shifty,
			//' ', nil, StyleDefault.Background(Color(OverlayColor)))
			'─', nil, SeparatorStyle)
	}

	return shifty
}

func (e *Editor) OnSearch() {
	clear(e.HighlightElements)

	e.IsContentSearch = true

	var end = false
	if e.SearchPattern == nil { e.SearchPattern = []rune{} }
	if e.Selection.IsSelectionNonEmpty() {
		e.SearchPattern = []rune(e.Selection.GetSelectionString(e.Content))
		e.SearchResults = Search(e.Content, string(e.SearchPattern))
		e.SearchResultIndex = 0
	}

	var patternx = len(e.SearchPattern)
	var isChanged = true

	// loop until escape or enter pressed
	for !end {

		e.DrawSearch(e.SearchPattern, patternx)
		e.Screen.Show()

		if isChanged && len(e.SearchPattern) > 0 && len(e.SearchResults) > 0 {

			var sy, sx = -1, -1
			e.X = 0

			result := e.SearchResults[e.SearchResultIndex]
			sy = result.Line; sx = result.Position

			if sx != -1 && sy != -1 {
				e.Row = sy; e.Col = sx; e.Focus()
				e.Selection.Ssx = sx;
				e.Selection.Ssy = sy;
				e.Selection.Sex = sx + len(e.SearchPattern);
				e.Selection.Sey = sy;
				e.Selection.IsSelected = true
				e.FocusCenter()
				e.DrawEverything()
				e.DrawSearch(e.SearchPattern, patternx)
				e.Screen.Show()
			} else {
				e.Selection.CleanSelection()
				e.DrawEverything()
				e.DrawSearch(e.SearchPattern, patternx)
				e.Screen.Show()
			}
			isChanged = false
		}

		switch ev := e.Screen.PollEvent().(type) { // poll and handle event
		case *EventResize:
			e.COLUMNS, e.ROWS = e.Screen.Size()

		case *EventKey:
			isChanged = false
			key := ev.Key()

			if key == KeyRune {
				e.SearchPattern = InsertTo(e.SearchPattern, patternx, ev.Rune())
				patternx++
				isChanged = true
				e.SearchResults = Search(e.Content, string(e.SearchPattern))
				e.SearchResultIndex = 0
			}
			if key == KeyBackspace2 && patternx > 0 && len(e.SearchPattern) > 0 {
				patternx--
				e.SearchPattern = Remove(e.SearchPattern, patternx)
				isChanged = true
				e.SearchResults = Search(e.Content, string(e.SearchPattern))
				e.SearchResultIndex = 0
			}
			if key == KeyLeft && patternx > 0 { patternx-- }
			if key == KeyRight && patternx < len(e.SearchPattern) { patternx++ }
			if key == KeyRight && patternx < len(e.SearchPattern) { patternx++ }
			if key == KeyDown {
				e.SearchResultIndex++
				if e.SearchResultIndex >= len(e.SearchResults) { e.SearchResultIndex = 0 }
				isChanged = true
			}
			if key == KeyUp {
				e.SearchResultIndex--
				if e.SearchResultIndex < 0 { e.SearchResultIndex = len(e.SearchResults) - 1}
				isChanged = true
			}
			if key == KeyCtrlX {
				e.SearchPattern = []rune{}
				patternx = 0
			}
			if key == KeyCtrlG {
				end = e.OnGlobalSearch()
				e.FocusCenter()
				e.DrawEverything()
				e.DrawSearch(e.SearchPattern, patternx)
				if end { e.CleanContentSearch() }
				e.Screen.Show()
			}
			if key == KeyESC || key == KeyCtrlF {
				end = true
				e.CleanContentSearch()
				e.Screen.Show()
			}
			if key == KeyEnter {
				if len(e.Content) == 0 { // global search if no content and enter
					end = e.OnGlobalSearch()
					e.FocusCenter()
					e.DrawEverything()
					e.DrawSearch(e.SearchPattern, patternx)
					if end { e.CleanContentSearch() }
					e.Screen.Show()
				} else {
					end = true
					e.FocusCenter()
					e.CleanContentSearch()
					e.Screen.Show()
				}
			}
		}
	}

	e.IsContentSearch = false
}

func (e *Editor) DrawSearch(pattern []rune, patternx int) {
	var prefix = []rune("search: ")

	for i := 0; i < len(prefix); i++ {
		e.Screen.SetContent(i + e.LINES_WIDTH+ e.FilesPanelWidth, e.ROWS-1, prefix[i], nil, StyleDefault)
	}

	e.Screen.SetContent(len(prefix) + e.LINES_WIDTH+ e.FilesPanelWidth, e.ROWS-1, ' ', nil, StyleDefault)

	for i := 0; i < len(pattern); i++ {
		e.Screen.SetContent(len(prefix) + i + e.LINES_WIDTH+ e.FilesPanelWidth, e.ROWS-1, pattern[i], nil, StyleDefault)
	}

	e.Screen.ShowCursor(len(prefix) + patternx + e.LINES_WIDTH+ e.FilesPanelWidth, e.ROWS-1)

	for i := len(prefix) + len(pattern) + e.LINES_WIDTH + e.FilesPanelWidth; i < e.COLUMNS; i++ {
		e.Screen.SetContent(i, e.ROWS-1, ' ', nil, StyleDefault)
	}

	if len(e.SearchResults) > 0 {
		status := fmt.Sprintf("  %d/%d", e.SearchResultIndex+1, len(e.SearchResults))

		for i := 0; i < len(status); i++ {
			e.Screen.SetContent(e.FilesPanelWidth + e.LINES_WIDTH + len(prefix) + len(pattern) + i , e.ROWS-1,
				rune(status[i]), nil, StyleDefault)
		}
	}


	e.Screen.ShowCursor(len(prefix) + patternx + e.LINES_WIDTH + e.FilesPanelWidth, e.ROWS-1)
}


func (e *Editor) CleanContentSearch() {
	for i := e.LINES_WIDTH + e.FilesPanelWidth; i < e.COLUMNS; i++ {
		e.Screen.SetContent(i, e.ROWS-1, ' ', nil, StyleDefault)
	}

	if len(e.Content) == 0 {
		e.Screen.HideCursor()
	}
}

func (e *Editor) OnGlobalSearch() bool {
	clear(e.HighlightElements)

	dir, _ := os.Getwd()

	start := time.Now()
	searchResults, filesProcessedCount, totalRowsProcessed := SearchOnDirParallel(dir, string(e.SearchPattern))
	elapsed := time.Since(start).String()

	e.IsOverlay = true
	defer e.OverlayFalse()

	var end = false
	var isChanged = true

	// loop until escape or enter pressed
	cwd, _ := os.Getwd()

	initialLang := e.treeSitterHighlighter.GetLangStr()

	for !end {
		var resultsCount = 0
		for _, searchResult := range searchResults { resultsCount += len(searchResult.Results) }

		var options = []string{}
		for _, searchResult := range searchResults {
			for _, result := range searchResult.Results {
				relativePath, _ := filepath.Rel(cwd, searchResult.File)

				text := fmt.Sprintf("%d/%d [%d:%d] %s ", len(options)+1, resultsCount, result.Line, result.Position, relativePath)
				options = append(options, text)
			}

		}

		height := MinMany(5, len(options) + 1)                // depends on min option len or 5 at min or how many rows to the end of e.Screen
		atx := 0 + e.FilesPanelWidth; aty := 0 // Define the window  position and dimensions
		style := StyleDefault

		var selectionEnd = false; var selected = 0; var selectedOffset = 0

		for !selectionEnd {
			if selected < selectedOffset { selectedOffset = selected }  // calculate offsets for scrolling completion
			if selected >= selectedOffset+height { selectedOffset = selected - height + 1 }

			if isChanged && resultsCount > 0 {
				isChanged = false
				e.DrawCodePreview(atx, aty, height, options, selectedOffset, selected, style, searchResults,
					fmt.Sprintf("global search: '%s', %d rows found, processed %d rows, %d files, elapsed %s",
						string(e.SearchPattern), resultsCount, totalRowsProcessed, filesProcessedCount, elapsed),
				)

				e.Screen.HideCursor()
				e.Screen.Show()
			}

			switch ev := e.Screen.PollEvent().(type) { // poll and handle event
			case *EventResize:
				e.COLUMNS, e.ROWS = e.Screen.Size()
				e.Screen.Sync()
				e.Screen.Clear()
				e.DrawEverything()
				e.Screen.Show()
				isChanged = true

			case *EventKey:
				key := ev.Key()
				if key == KeyEscape || key == KeyBackspace || key == KeyBackspace2 {
					e.Screen.Clear()
					selectionEnd = true
					end = true
					if e.treeSitterHighlighter.GetLangStr() != initialLang  {
						e.treeSitterHighlighter.SetLang(initialLang)
						e.UpdateColors()
					}

					return true
				}

				if key == KeyDown && selected < len(options)-1 { selected++; isChanged = true }
				if key == KeyUp && selected > 0 { selected--; isChanged = true }

				if key == KeyEnter {
					end = true
					file, searchResult, found := e.findSearchGlobalOption(searchResults, selected)
					if found  {
						if e.AbsoluteFilePath != file { e.OpenFile(file) }
						searchPattern, _ := ParsePattern(string(e.SearchPattern))
						e.Selection.CleanSelection()
						e.Row = searchResult.Line - 1
						e.Col = searchResult.Position + len(searchPattern)
						e.Selection.Ssy = e.Row
						e.Selection.Sey = e.Row
						e.Selection.Ssx = searchResult.Position
						e.Selection.Sex = searchResult.Position + len(searchPattern)
						e.Selection.IsSelected = true
						e.Focus()

						return true
					}

				}
			}
		}
	}

	if e.treeSitterHighlighter.GetLangStr() != initialLang  {
		e.treeSitterHighlighter.SetLang(initialLang)
	}

	return false
}

func (e *Editor) DrawCodePreview(atx int, aty int, height int, options []string, selectedOffset int, selected int,
	style Style, searchResults []FileSearchResult, status string)  {

	//searchPattern, _ := ParsePattern(string(e.SearchPattern))

	// draw options
	for row := aty; row < aty+height; row++ {
		if row >= len(options) || row >= height { break }

		var option = options[row+selectedOffset]

		isRowSelected := selected == row + selectedOffset
		if isRowSelected { style = style.Background(Color(AccentColor)) } else {
			style = StyleDefault.Background(Color(OverlayColor))
		}

		for i, ch := range option { e.Screen.SetContent(atx+i, row, ch, nil, style) }

		for i := atx + len(option); i < e.COLUMNS; i++ { e.Screen.SetContent(i, row, ' ', nil, style) }
	}

	for i := atx; i < e.COLUMNS; i++ {
		e.Screen.SetContent(i, height, ' ', nil, StyleDefault)
	}

	file, searchResult, found := e.findSearchGlobalOption(searchResults, selected)
	if found {
		rowsToShow := e.ROWS - height
		previewContent := e.ReadContent(file, searchResult.Line-rowsToShow/2, searchResult.Line+rowsToShow/2)
		text := ConvertContentToString(previewContent)
		lang := DetectLang(file)
		if e.treeSitterHighlighter.GetLangStr() != lang  { e.treeSitterHighlighter.SetLang(lang) }
		previewContentColors := e.treeSitterHighlighter.Colorize(text)

		// clear
		for j := height + 1; j < e.ROWS; j++ {
			for i := atx; i < e.COLUMNS; i++ {
				e.Screen.SetContent(i, j, ' ', nil, StyleDefault)
			}
		}

		linenumber := searchResult.Line - rowsToShow/2
		if linenumber < 0 { linenumber = 0 }

		// draw preview
		for row := 0; row < len(previewContent); row++ {
			y := row + height + 1
			if y >= e.ROWS { break }
			var shiftTabs = 0

			var lineNumberStyle = StyleDefault.Foreground(ColorDimGray)
			for index, char := range CenterNumber(linenumber+1, e.LINES_WIDTH) {
				e.Screen.SetContent(index + e.FilesPanelWidth, y, char, nil, lineNumberStyle)
			}

			for col := 0; col < len(previewContent[row]); col++ {

				chstyle := StyleDefault
				if row < len(previewContentColors) && col < len(previewContentColors[row]) {
					color := previewContentColors[row][col]
					if color > 0 { chstyle = StyleDefault.Foreground(Color(color)) }
				}

				//if linenumber == searchResult.Line-1 &&  // color match
				//	col >= searchResult.Position && col < searchResult.Position + len(searchPattern) {
				//	chstyle = chstyle.Background(Color(SelectionColor))
				//}

				if linenumber == searchResult.Line-1 {
					chstyle = chstyle.Background(Color(SelectionColor))
				}

				if previewContent[row][col] == '\n' { continue }
				if previewContent[row][col] == '\t' {
					for i := 0; i < e.langTabWidth; i++ {
						e.Screen.SetContent(atx+e.LINES_WIDTH+col+shiftTabs, y, ' ', nil, chstyle)
						if i != e.langTabWidth-1 { shiftTabs++ }
					}
				} else {
					e.Screen.SetContent(atx+e.LINES_WIDTH+col+shiftTabs, y, previewContent[row][col], nil, chstyle)
				}

				if atx+e.LINES_WIDTH+col+shiftTabs >= e.COLUMNS { break }
			}

			for i := atx + len(previewContent[row]) + e.LINES_WIDTH + shiftTabs; i < e.COLUMNS; i++ {
				e.Screen.SetContent(i, y, ' ', nil, StyleDefault)
			}

			linenumber++
		}

		label := append([]rune(status), []rune(strings.Repeat(" ", e.COLUMNS - atx))...)

		for i := 0; i < len(label); i++ {
			e.Screen.SetContent(atx + i, e.ROWS-1, label[i], nil, StyleDefault)
		}
	}

}

func (e *Editor) findSearchGlobalOption(searchResults []FileSearchResult, selected int) (string, SearchResult, bool) {
	var i = 0
	for _, searchResult := range searchResults {
		for _, result := range searchResult.Results {
			if i == selected {
				return searchResult.File, result, true
			}
			i++
		}
	}
	return "", SearchResult{}, false
}

func (e *Editor) OnFilesTree(forceOpen bool) {
	e.IsFileSelection = true

	if e.FilesPanelWidth == 0 {
		tree, _ := ReadDirTree(e.Cwd, "", false, 0)
		e.Tree = tree
		if len(tree.Childs) == 0 { return }
		e.FilesPanelWidth = 28
		// root is always opened
		e.Tree.IsDirOpen = true
	}
	
	if e.Filename != "" { e.DrawEverything() }

	var end = false
	var patternx = len(e.FilesSearchPattern)

	// loop until escape or enter pressed
	for !end {
		//_, screenRows := e.Screen.Size()
		_, screenRows := 0, e.ROWS

		if e.FileSelectedIndex != -1 && e.FileSelectedIndex < e.FileScrollingOffset {
			e.FileScrollingOffset = e.FileSelectedIndex
		}
		if e.FileSelectedIndex >= e.FileScrollingOffset + screenRows {
			e.FileScrollingOffset = e.FileSelectedIndex - screenRows + 1
		}

		treeSize := TreeSize(e.Tree, 0)
		var aty = 0
		var fileindex = 0

		for row := 0; row < screenRows; row++ {
			for col := 0; col < e.FilesPanelWidth-2; col++ { // clean
				e.Screen.SetContent(col, row, ' ', nil, StyleDefault)
			}
			//e.Screen.SetContent(e.FilesPanelWidth-2, row, '▕', nil, SeparatorStyle)
		}

		e.DrawTree(e.Tree, 0, &fileindex, &aty)
		e.DrawTreeSearch(e.FilesSearchPattern, patternx)
		e.Screen.Show()

		switch ev := e.Screen.PollEvent().(type) { // poll and handle event
		case *EventMouse:
			mx, my := ev.Position()
			buttons := ev.Buttons()
			//modifiers := ev.Modifiers()

			if mx > e.FilesPanelWidth - 3  { end = true; continue }
			if my >= screenRows { end = true; continue }

			if buttons & WheelDown != 0 && treeSize > screenRows {
				if e.FileScrollingOffset < treeSize - screenRows {
					e.FileScrollingOffset++
				}
			}
			if buttons & WheelUp != 0 && e.FileScrollingOffset > 0 {
				e.FileScrollingOffset--
			}

			if my < treeSize { e.FileSelectedIndex = my + e.FileScrollingOffset }
			if buttons & Button1 == 1 {
				e.FileSelectedIndex = my + e.FileScrollingOffset
				if e.FileSelectedIndex < 0  { continue }
				if e.FileSelectedIndex >= treeSize { continue }
				if e.FileSelectedIndex >= treeSize { continue }
				if !e.IsMouseUnderFile(mx) { continue }
				end = e.SelectAndOpenFile()
				if end && e.IsFilesSearch  {
					tree, _ := ReadDirTree(e.Cwd, "", false, 0)
					e.Tree = tree
					e.Tree.IsDirOpen = true
					SetDirOpenFlag(&e.Tree, e.InputFile)
					e.IsFilesSearch = false
					e.FilesSearchPattern = []rune{}
				}
			}

		case *EventKey:
			key := ev.Key()

			if key == KeyCtrlQ { e.Screen.Fini(); os.Exit(1) }
			if key == KeyCtrlN { e.NewFileOrDir() }
			if key == KeyCtrlF { e.IsFilesSearch = !e.IsFilesSearch }
			if key == KeyEscape && !e.IsFilesSearch { end = true; e.FilesPanelWidth =  0 }
			if key == KeyEscape && e.IsFilesSearch { end = true; e.IsFilesSearch = false; e.CleanFilesSearch();e.Screen.Show() }
			if key == KeyDown { e.FileSelectedIndex =  Min(treeSize-1, e.FileSelectedIndex+1) }
			if key == KeyUp { e.FileSelectedIndex = Max(0, e.FileSelectedIndex-1) }
			if key == KeyLeft && e.IsFilesSearch && patternx > 0 { patternx-- }
			if key == KeyRight && e.IsFilesSearch && patternx < len(e.FilesSearchPattern) { patternx++ }
			if key == KeyCtrlT {
				end = true
				e.IsFilesSearch = false
				e.FilesPanelWidth = 0
			}
			if key == KeyBackspace2  && e.IsFilesSearch && patternx > 0 && len(e.FilesSearchPattern) > 0 {
				patternx--
				e.FilesSearchPattern = Remove(e.FilesSearchPattern, patternx)
				if len(e.FilesSearchPattern) != 0 {
					tree, _ := ReadDirTree(e.Cwd, string(e.FilesSearchPattern), true, 0)
					tree = FilterIfLeafEmpty(tree)
					e.Tree = tree
				} else {
					tree, _ := ReadDirTree(e.Cwd, "", false, 0)
					e.Tree = tree
				}

				e.Tree.IsDirOpen = true
				e.FileScrollingOffset = 0
				_, i := FindFirstFile(e.Tree, 0)
				e.FileSelectedIndex = i
			}
			if key == KeyRune {
				e.IsFilesSearch = true
				e.FilesSearchPattern = InsertTo(e.FilesSearchPattern, patternx, ev.Rune())
				patternx++
				tree, _ := ReadDirTree(e.Cwd, string(e.FilesSearchPattern), true, 0)
				tree = FilterIfLeafEmpty(tree)
				e.Tree = tree
				e.Tree.IsDirOpen = true
				e.FileScrollingOffset = 0
				_, i := FindFirstFile(e.Tree, 0)
				e.FileSelectedIndex = i
			}

			if key == KeyEnter || !e.IsFilesSearch  && (key == KeyLeft || key == KeyRight) {
				end = e.SelectAndOpenFile()
				if end && e.IsFilesSearch  {
					tree, _ := ReadDirTree(e.Cwd, "", false, 0)
					e.Tree = tree
					e.Tree.IsDirOpen = true
					SetDirOpenFlag(&e.Tree, e.InputFile)
					e.IsFilesSearch = false
					e.FilesSearchPattern = []rune{}
				}
			}
		}
	}

	e.IsFileSelection = false
}

func (e *Editor) DrawTree(fileInfo FileInfo, level int, fileindex *int, aty *int) {

	isNeedToShow := *fileindex >= e.FileScrollingOffset

	if isNeedToShow {
		if *aty >= e.ROWS { return }

		style := StyleDefault
		isSelectedFile := e.IsFileSelection && e.FileSelectedIndex != -1 && *fileindex  == e.FileSelectedIndex
		if fileInfo.IsDir { style = style.Foreground(Color(AccentColor2)) } else {
			style = style.Foreground(Color(AccentColor3))
		}
		if isSelectedFile { style = style.Foreground(Color(AccentColor)) }

		if e.InputFile != "" && e.InputFile == fileInfo.FullName {
			style = style.Background(Color(AccentColor)).Foreground(ColorWhite)
		}

		for i := 0; i <= level; i++ {
			if i+1 >= e.FilesPanelWidth-2 { break }
			e.Screen.SetContent(i + 1, *aty, ' ', nil, StyleDefault)
		}

		label := []rune(" " + fileInfo.Name + " ")
		for i := 0; i < len(label); i++ {
			if i+1 + level >= e.FilesPanelWidth-2 { break }
			e.Screen.SetContent(i + 1 + level, *aty, label[i], nil, style)
		}
		//e.Screen.Show()
		*aty++
	}

	*fileindex++

	if fileInfo.IsDir && fileInfo.IsDirOpen {
		for _, child := range fileInfo.Childs {
			e.DrawTree(child, level+1, fileindex, aty)
		}
	}

}

func (e *Editor) DrawTreeSearch(filterPattern []rune, patternx int) {
	e.Screen.HideCursor()

	if e.IsFilesSearch {
		pref := " search: "
		e.Screen.ShowCursor(len(pref) + patternx, e.ROWS-1)
		for i, ch := range pref { // draw prefix
			e.Screen.SetContent(i, e.ROWS-1, ch, nil, StyleDefault)
		}

		for i, ch := range filterPattern { // draw pattern
			e.Screen.SetContent(i+len(pref), e.ROWS-1, ch, nil, StyleDefault)
		}
		for col := len(pref) + len(filterPattern); col < e.FilesPanelWidth- 1; col++ { // clean
			e.Screen.SetContent(col, e.ROWS-1, ' ', nil, StyleDefault)
		}
	}
}


func (e *Editor) CleanFilesSearch() {
	e.Screen.HideCursor()
	for col :=0; col < e.FilesPanelWidth - 1; col++ { // clean
		e.Screen.SetContent(col, e.ROWS-1, ' ', nil, StyleDefault)
	}
}

func (e *Editor) SelectAndOpenFile() bool {
	found, selectedFile := GetSelected(e.Tree, e.FileSelectedIndex)
	if found {
		if selectedFile.IsDir {
			selectedFile.IsDirOpen = !selectedFile.IsDirOpen
			return false
		} else {
			e.InputFile = selectedFile.FullName
			e.OpenFile(e.InputFile)
			return true
		}
	}
	return false
}
func (e *Editor) IsMouseUnderFile(mx int) bool {
	found, selectedFile := GetSelected(e.Tree, e.FileSelectedIndex)
	if found {
		if selectedFile.Level + len(selectedFile.Name) + 1 >= mx {
			if mx < selectedFile.Level + 2 { return false } // 2 is spacing
			return true
		} else {
			return false
		}
	}
	return false
}



func (e *Editor) OverlayFalse() {
	e.IsOverlay = false
}

func (e *Editor) UpdateColors() {
	if !e.IsColorize { return }
	if e.Lang == "" { return }
	//if len(e.Content) >= 10000 {
	//	line := string(e.Content[e.Row])
	//	linecolors := HighlighterGlobal.Colorize(line, e.Filename)
	//	e.Colors[e.Row] = linecolors[0]
	//} else {
	//	code := ConvertContentToString(e.Content)
	//	e.Colors = HighlighterGlobal.Colorize(code, e.Filename)
	//}

	start := time.Now()
	code := ConvertContentToString(e.Content)
	e.Colors = e.treeSitterHighlighter.Colorize(code)
	Log.Info("tree-sitter re colorized, elapsed: " + time.Since(start).String())
}

func (e *Editor) UpdateColorsFull() {
	if !e.IsColorize { return }
	if e.Lang == "" { return }

	//code := ConvertContentToString(e.Content)
	//e.Colors = HighlighterGlobal.Colorize(code, e.Filename)

	start := time.Now()
	code := ConvertContentToString(e.Content)
	e.Colors = e.treeSitterHighlighter.Colorize(code)
	Log.Info("tree-sitter re colorized, elapsed: " + time.Since(start).String())
}

func (e *Editor) UpdateColorsAtLine(at int) {
	if !e.IsColorize { return }
	if e.Lang == "" { return }
	if at >= len(e.Colors) { return }

	line := string(e.Content[at])
	if line == "" { e.Colors[at] = []int{}; return }
	//linecolors := HighlighterGlobal.Colorize(line, e.Filename)
	//e.Colors[at] = linecolors[0]

	start := time.Now()
	code := ConvertContentToString(e.Content)
	e.Colors = e.treeSitterHighlighter.Colorize(code)
	Log.Info("tree-sitter re colorized, elapsed: " + time.Since(start).String())

}


// todo, get rid of this function, cause UpdateColors is slow for big files
func (e *Editor) UpdateNeeded() {
	e.Update = true
	e.IsContentChanged = true
	if len(e.Content) <= 10000 { go e.WriteFile() }
	e.UpdateColors()
	e.FindTests()
}

func (e *Editor) OnProcessRun(newRun bool) {
	if newRun && (e.Lang == "" || e.langConf.Cmd == "") { return }

	if e.ProcessPanelHeight == 0 {
		e.ProcessPanelHeight = 10
		e.COLUMNS, e.ROWS = e.Screen.Size()
		e.ROWS -= e.ProcessPanelHeight
	}


	var args = []string{e.AbsoluteFilePath }

	if e.langConf.CmdArgs != "" {
		args = append(strings.Split(e.langConf.CmdArgs, " "), e.AbsoluteFilePath)
	}

	cmd := e.langConf.Cmd

	if !newRun && e.Process != nil && e.Process.Cmd != nil {
		// use prev cmd and args
		cmd = e.Process.Cmd.Path
		args = e.Process.Cmd.Args[1:]
	}

	ResetSelectionColor()
	e.Process = NewProcess(cmd, args...)
	e.Process.Cmd.Env = append(os.Environ())

	if e.Lang == "python" {
		// printing immediately
		e.Process.Cmd.Env = append(e.Process.Cmd.Env, "PYTHONUNBUFFERED=1")
	}

	e.ProcessContent = [][]rune{}
	e.ProcessPanelScroll = 0
	e.ProcessPanelSpacing = 2

	e.Process.Start()

	go func() {
		for range e.Process.Updates {

			//e.ProcessContent = append(e.ProcessContent, []rune(line))

			//newLines := e.Process.Lines[len(e.ProcessContent):]
			newLines := e.Process.GetLines(len(e.ProcessContent))
			for _, line := range newLines {
				e.ProcessContent = append(e.ProcessContent, []rune(line))
				if len(e.ProcessContent) > e.ProcessPanelHeight {
					if e.ProcessPanelScroll >= len(e.ProcessContent) - e.ProcessPanelHeight - 1  {
						e.ProcessPanelScroll = len(e.ProcessContent) - e.ProcessPanelHeight + 1 // focusing
						e.ProcessPanelScroll = Max(0, e.ProcessPanelScroll)
					}
				}
			}

			e.DrawProcessPanel()
			e.Screen.Show()

			if e.Process.IsStopped() {
				if len(e.ProcessContent) > e.ProcessPanelHeight { // focusing
					e.ProcessPanelScroll = len(e.ProcessContent) - e.ProcessPanelHeight + 1
				}
				e.DrawProcessPanel()
				e.Screen.Show()
				break
			}
		}
	}()

}

func (e *Editor) OnProcessKeyHandle(key Key, keyrune rune) {
	if key == KeyCtrlF {
		e.OnProcessSearch()
	}

	if key == KeyUp {
		e.ProcessPanelCursorY--
		if e.ProcessPanelCursorY < 0 { e.ProcessPanelCursorY = 0 }
		if e.ProcessPanelCursorX > len(e.ProcessContent[e.ProcessPanelCursorY]) {
			e.ProcessPanelCursorX = len(e.ProcessContent[e.ProcessPanelCursorY])
		}
		e.FocusProcessPanel()
	}
	if key == KeyDown  {
		e.ProcessPanelCursorY++
		if e.ProcessPanelCursorY >= len(e.ProcessContent) { e.ProcessPanelCursorY = len(e.ProcessContent)-1 }
		if e.ProcessPanelCursorX > len(e.ProcessContent[e.ProcessPanelCursorY]) {
			e.ProcessPanelCursorX = len(e.ProcessContent[e.ProcessPanelCursorY])
		}
		e.FocusProcessPanel()
	}
	if key == KeyRight  {
		e.ProcessPanelCursorX++
		if e.ProcessPanelCursorX > len(e.ProcessContent[e.ProcessPanelCursorY]) {
			e.ProcessPanelCursorX = len(e.ProcessContent[e.ProcessPanelCursorY])
		}

		//if e.ProcessPanelCursorX >= e.TERMINAL_WIDHT - e.ProcessPanelSpacing {
		//	e.ProcessPanelHScroll = e.ProcessPanelCursorX - e.TERMINAL_WIDHT + e.ProcessPanelSpacing
		//}
	}
	if key == KeyLeft  {
		e.ProcessPanelCursorX--
		if e.ProcessPanelCursorX < 0 { e.ProcessPanelCursorX = 0 }
		if e.ProcessPanelCursorX > len(e.ProcessContent[e.ProcessPanelCursorY]) {
			e.ProcessPanelCursorX = len(e.ProcessContent[e.ProcessPanelCursorY])
		}

		//if e.ProcessPanelCursorX >= e.TERMINAL_WIDHT - e.ProcessPanelSpacing {
		//	e.ProcessPanelHScroll = e.ProcessPanelCursorX - e.TERMINAL_WIDHT + e.ProcessPanelSpacing
		//}
	}
	if key == KeyRune && keyrune == 'l' {
		e.ProcessPanelHScroll++
	}
	if key == KeyRune && keyrune == 'l' {
		e.ProcessPanelHScroll++
	}
	if key == KeyRune && keyrune == 's' {
		e.OnProcessStop()
	}
	if key == KeyRune && keyrune == 'f' {
		if len(e.ProcessContent) > e.ProcessPanelHeight { // focusing
			e.ProcessPanelScroll = len(e.ProcessContent) - e.ProcessPanelHeight + 1
		}
	}
}

func (e *Editor) OnProcessStop() {
	if e.Process != nil && e.Process.Cmd != nil && !e.Process.IsStopped() {
		e.Process.Stop()
		time.Sleep(time.Millisecond * 300) // give a time to show 'kill' message
	}

	if e.Dap.IsStarted {
		e.OnDebugStop()
	}

	if len(e.ProcessContent) > e.ProcessPanelHeight { // focusing
		e.ProcessPanelScroll = len(e.ProcessContent) - e.ProcessPanelHeight + 2
	}
	ResetSelectionColor()
}

func (e *Editor) OnFileUpdate() {
	row, col := e.Row, e.Col // save cursor
	x, y := e.X, e.Y         // safe scroll
	e.OpenFile(e.AbsoluteFilePath)

	// if row and col fits to content, restore cursor
	if row < len(e.Content) {
		e.Row = row
		e.X = x
		if col < len(e.Content[row]) {
			e.Col = col
			e.Y = y
		}
	}
	e.DrawEverything()
	e.Screen.Show()
}


func (e *Editor) OnFilesTreeUpdate(event notify.EventInfo) {
	fullname := event.Path()
	name := filepath.Base(fullname)
	dir := filepath.Dir(fullname)
	parentNode := FindByFullName(&e.Tree, dir)
	if parentNode == nil { return }

	switch event.Event() {
	case notify.Create:
		// find and add node to parent
		fileInfo, err := os.Stat(fullname)
		if err != nil { return }
		isDir := fileInfo.IsDir()

		f := FileInfo{
			Name: name, FullName: fullname,
			IsDir: isDir, IsDirOpen: false,
			Childs: []FileInfo{}, Level: parentNode.Level + 1,
		}

		parentNode.Childs = append(parentNode.Childs, f)
		SortTree(*parentNode)

	case notify.Remove:
		// find and remove node from parent
		for i, child := range parentNode.Childs {
			if child.FullName == fullname {
				parentNode.Childs = Remove(parentNode.Childs, i)
				break
			}
		}
	default:
		// update all
		//tree, _ := ReadDirTree(e.Cwd, "", false, 0)
		//tree.IsDirOpen = true
		//e.Tree = tree
	}

	e.DrawEverything()
	e.Screen.Show()
}

// draw logo
func (e *Editor) DrawLogo() {
	// https://patorjk.com/software/taag/#p=testall&h=0&f=3D%20Diagonal&t=edgo
	logo :=
`
 _______ ______   ______  _____ 
 |______ |     \ |  ____ |     |
 |______ |_____/ |_____| |_____|
                                
`
	// split logo by new line
	lines := strings.Split(logo, "\n")

	logoWidth := len(lines[1])
	//fromx := e.COLUMNS / 2 - logoWidth / 2
	fromy := e.ROWS / 2 - len(lines) / 2

	fromx := (e.COLUMNS + 28 - 2) / 2 - logoWidth / 2

	for i, line  := range lines {
		for j, ch := range line {
			e.Screen.SetContent(j + fromx, i + fromy, ch, nil,
				StyleDefault.Foreground(Color(AccentColor2)))
		}
	}

	e.Screen.Show()
}

func (e *Editor) DrawTest(line int, row int) {
	x := e.COLUMNS-2
	
	//x := e.FilesPanelWidth + e.LINES_WIDTH/2

	//for i:= 0; i < e.LINES_WIDTH; i++ {
	//	e.Screen.SetContent(e.FilesPanelWidth + i, row, ' ', nil, StyleDefault)
	//}

	e.Screen.SetContent(x, row, '▶', nil,
		StyleDefault.Foreground(Color(HighlighterGlobal.GetRunButtonStyle())))
	//e.Screen.Show()
}

func (e *Editor) NewFileOrDir() {
	inputName := make([]rune, 0)
	cursorx := 0
	end := false
	pref := " create "
	directory := ""

	// add current
	found, selectedFile := GetSelected(e.Tree, e.FileSelectedIndex)
	if found {
		rel, err := filepath.Rel(e.Cwd, selectedFile.FullName)
		if err != nil { return }

		if selectedFile.IsDir {
			directory = rel + string(filepath.Separator)
		} else {
			directory = filepath.Dir(rel)
		}
	}


	for !end {

		e.Screen.ShowCursor(len(pref) + len(directory) + cursorx, e.ROWS-1)

		for col := 0; col <= len(pref) + len(inputName) || col < e.FilesPanelWidth; col++ { // clean
			e.Screen.SetContent(col, e.ROWS-1, ' ', nil, StyleDefault)
		}

		for i, ch := range pref { // draw prefix
			e.Screen.SetContent(i, e.ROWS-1, ch, nil, StyleDefault)
		}
		for i, ch := range directory { // draw directory
			e.Screen.SetContent(i+len(pref), e.ROWS-1, ch, nil, StyleDefault)
		}

		for i, ch := range inputName { // draw inputName
			e.Screen.SetContent(i+len(pref)+len(directory), e.ROWS-1, ch, nil, StyleDefault)
		}

		e.Screen.Show()

		switch ev := e.Screen.PollEvent().(type) {
		case *EventKey:
			key := ev.Key()

			if key == KeyRune {
				inputName = InsertTo(inputName, cursorx, ev.Rune()); cursorx++
			}
			if key == KeyBackspace2 && cursorx > 0 && len(inputName) > 0 {
				cursorx--; inputName = Remove(inputName, cursorx)
			}
			if key == KeyLeft && cursorx > 0 { cursorx-- }
			if key == KeyRight && cursorx < len(inputName) { cursorx++ }
			if key == KeyEscape { end = true }

			if key == KeyEnter {
				name := filepath.Join(directory, string(inputName))
				nameabs, err := filepath.Abs(name)
				if err != nil { return }

				isCreateDir := strings.HasSuffix(string(inputName), string(filepath.Separator))

				if isCreateDir {
					// dir/  - create dir
					err := os.MkdirAll(nameabs, 0750)
					if err != nil { return }
				} else {  // create file
					file, err := os.Create(nameabs)
					defer file.Close()
					if err != nil { return }
				}

				end = true
			}
		}
	}

	for col := 0; col <= len(pref) + len(inputName); col++ { // clean
		e.Screen.SetContent(col, e.ROWS-1, ' ', nil, StyleDefault)
	}
}

func (e *Editor) OnCursorChanged() {
	clear(e.HighlightElements)

	if e.Selection.IsSelectionNonEmpty() { return }

	start := time.Now()
	nodename, noderange := e.treeSitterHighlighter.GetNodeAt(e.Row, e.Col, e.Row, e.Col)


	if strings.Contains(nodename, "identifier") {
		//if len(e.Content) >= noderange.Ssy { return }
		runes := e.Content[noderange.Ssy][noderange.Ssx:noderange.Sex]
		content := string(runes)


		searchResults := Search(e.Content, content)
		e.HighlightElements = make(map[int][]NodeRange)

		for _, searchResult := range searchResults {
			searchnodename, searchnoderange := e.treeSitterHighlighter.GetNodeAt(
				searchResult.Line, searchResult.Position, searchResult.Line, searchResult.Position + len(content))
			if nodename == searchnodename {
				if len(content) != searchnoderange.Sex - searchnoderange.Ssx { continue }
				e.HighlightElements[searchResult.Line] = append(e.HighlightElements[searchResult.Line], searchnoderange)
			}
		}

		//Log.Info(nodename,
		//	fmt.Sprintf("%d %d %d %d %s elapsed %s", noderange.Ssx,
		//		noderange.Ssy, noderange.Sex, noderange.Sey, content, elapsed))
	}

	elapsed := time.Since(start).String()
	Log.Info(nodename,
		fmt.Sprintf("%d %d %d %d elapsed %s", noderange.Ssx,
			noderange.Ssy, noderange.Sex, noderange.Sey, elapsed))
}

func (e *Editor) OnProcessSearch() {
	var end = false
	if e.ProcessPanelSearchPattern == nil { e.ProcessPanelSearchPattern = []rune{} }

	var patternx = len(e.ProcessPanelSearchPattern)
	var isChanged = true

	// loop until escape or enter pressed
	for !end {

		e.DrawProcessPanelSearch(e.ProcessPanelSearchPattern, patternx)
		e.Screen.Show()

		if isChanged && len(e.ProcessPanelSearchPattern) > 0 && len(e.ProcessPanelSearchResults) > 0 {

			var sy, sx = -1, -1
			result := e.ProcessPanelSearchResults[e.ProcessPanelSearchResultIndex]
			sy = result.Line; sx = result.Position

			if sx != -1 && sy != -1 {
				e.ProcessPanelCursorY = sy; e.ProcessPanelCursorX = sx;
				e.ProcessPanelSelection.Ssx = sx;
				e.ProcessPanelSelection.Ssy = sy;
				e.ProcessPanelSelection.Sex = sx + len(e.ProcessPanelSearchPattern);
				e.ProcessPanelSelection.Sey = sy;
				e.ProcessPanelSelection.IsSelected = true
				e.CleanProcessPanel()
				e.FocusProcessPanel()
				e.DrawProcessPanel()
				e.DrawProcessPanelSearch(e.ProcessPanelSearchPattern, patternx)
				e.Screen.Show()
			} else {
				e.ProcessPanelSelection.CleanSelection()
				e.CleanProcessPanel()
				e.DrawProcessPanel()
				e.DrawProcessPanelSearch(e.ProcessPanelSearchPattern, patternx)
				e.Screen.Show()
			}
			isChanged = false
		}

		switch ev := e.Screen.PollEvent().(type) { // poll and handle event
		case *EventResize:
			e.COLUMNS, e.ROWS = e.Screen.Size()

		case *EventKey:
			isChanged = false
			key := ev.Key()

			if key == KeyCtrlQ { e.Screen.Fini(); os.Exit(1) }

			if key == KeyRune {
				e.ProcessPanelSearchPattern = InsertTo(e.ProcessPanelSearchPattern, patternx, ev.Rune())
				patternx++
				isChanged = true
				e.ProcessPanelSearchResults = Search(e.ProcessContent, string(e.ProcessPanelSearchPattern))
				e.ProcessPanelSearchResultIndex = len(e.ProcessPanelSearchResults)-1
			}
			if key == KeyBackspace2 && patternx > 0 && len(e.ProcessPanelSearchPattern) > 0 {
				patternx--
				e.ProcessPanelSearchPattern = Remove(e.ProcessPanelSearchPattern, patternx)
				isChanged = true
				e.ProcessPanelSearchResults = Search(e.ProcessContent, string(e.ProcessPanelSearchPattern))
				e.ProcessPanelSearchResultIndex = len(e.ProcessPanelSearchResults)-1
			}
			if key == KeyLeft && patternx > 0 { patternx-- }
			if key == KeyRight && patternx < len(e.ProcessPanelSearchPattern) { patternx++ }
			if key == KeyDown {
				e.ProcessPanelSearchResultIndex++
				if e.ProcessPanelSearchResultIndex >= len(e.ProcessPanelSearchResults) { e.ProcessPanelSearchResultIndex = 0 }
				isChanged = true
			}
			if key == KeyUp {
				e.ProcessPanelSearchResultIndex--
				if e.ProcessPanelSearchResultIndex < 0 { e.ProcessPanelSearchResultIndex = len(e.ProcessPanelSearchResults) - 1}
				isChanged = true
			}
			if key == KeyCtrlX {
				e.ProcessPanelSearchPattern = []rune{}
				patternx = 0
			}

			if key == KeyESC || key == KeyCtrlF || key == KeyEnter {
				end = true
			}
		}
	}

	e.IsProcessPanelSearch = false
}

func (e *Editor) DrawProcessPanelSearch(pattern []rune, patternx int) {

	var prefix = []rune("  search: ")

	for i := 0; i < len(prefix); i++ {
		e.Screen.SetContent(i, e.TERMINAL_HEIGHT-1, prefix[i], nil, StyleDefault)
	}

	e.Screen.SetContent(len(prefix), e.TERMINAL_HEIGHT-1, ' ', nil, StyleDefault)

	for i := 0; i < len(pattern); i++ {
		e.Screen.SetContent(len(prefix) + i, e.TERMINAL_HEIGHT-1, pattern[i], nil, StyleDefault)
	}

	e.Screen.ShowCursor(len(prefix) + patternx, e.TERMINAL_HEIGHT-1)

	for i := len(prefix) + len(pattern); i < e.COLUMNS; i++ {
		e.Screen.SetContent(i, e.TERMINAL_HEIGHT-1, ' ', nil, StyleDefault)
	}

	if len(e.ProcessPanelSearchResults) > 0 {
		status := fmt.Sprintf("  %d/%d", e.ProcessPanelSearchResultIndex+1, len(e.ProcessPanelSearchResults))

		for i := 0; i < len(status); i++ {
			e.Screen.SetContent(len(prefix) + len(pattern) + i , e.TERMINAL_HEIGHT-1,
				rune(status[i]), nil, StyleDefault)
		}
	}
}

func (e *Editor) GoToLine() {
	var input = []rune{}
	var patternx = 0

	var end = false
	for !end {

		switch ev := e.Screen.PollEvent().(type) { // poll and handle event

		case *EventKey:
			key := ev.Key()

			if key == KeyRune {
				input = InsertTo(input, patternx, ev.Rune())
				patternx++
			}
			if key == KeyBackspace2 && patternx > 0 && len(input) > 0 {
				patternx--
				input = Remove(input, patternx)
			}
			if key == KeyESC  || key == KeyCtrlL { return }
			if key == KeyEnter {
				end = true
			}
		}
	}

	result, err := strconv.Atoi(string(input))
	if err != nil { return }

	e.Row = result - 1
	e.Col = 0

	if e.Row < 0 { e.Row = 0 } // fit to content
	if e.Row >= len(e.Content) { e.Row = len(e.Content) - 1 } // fit to content
	e.Focus()
}