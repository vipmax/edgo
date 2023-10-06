package ui

import (
	. "edgo/internal/highlighter"
	"edgo/internal/process"
	. "edgo/internal/tests"
	"edgo/internal/utils"
	"github.com/gdamore/tcell"
	sitter "github.com/smacker/go-tree-sitter"
	"os"
	"strings"
)

func getTestByLang(lang string, filepath string) Test {
	switch lang {
	case "go":
		if !strings.HasSuffix(filepath, "test.go") { return nil }
		return &GoTest{}

	case "python":
		return &PythonTest{}

	case "javascript":
		return &JavascriptTest{}

	default:
	}

	return nil
}

func (e *Editor) FindTests() {
	clear(e.Tests)

	if e.TestFinder.TestQuery == nil || e.TestFinder.Lang != e.Lang {
		e.Test = getTestByLang(e.Lang, e.AbsoluteFilePath)
		if e.Test == nil { return }
		queryStr := e.Test.Query()
		q, _ := sitter.NewQuery([]byte(queryStr), e.treeSitterHighlighter.GetLang())
		e.TestFinder.TestQuery = q
		e.TestFinder.Lang = e.Lang
	}

	codeBytes := []byte(utils.ConvertContentToString(e.Content))
	rootNode := e.treeSitterHighlighter.GetTree().RootNode()
	e.Tests = e.Test.Find(&e.TestFinder, rootNode, e.AbsoluteFilePath, codeBytes)
}


func (e *Editor) RunTest(test TestData) {
	if e.Test == nil { return }

	args := e.Test.Run(test)

	if e.Lang == "" || e.langConf.Cmd == "" { return }

	if e.ProcessPanelHeight == 0 {
		e.ProcessPanelHeight = 10
		e.COLUMNS, e.ROWS = e.Screen.Size()
		e.ROWS -= e.ProcessPanelHeight
	}

	ResetSelectionColor()

	e.Process = process.NewProcess(args[0], args[1:]...)
	e.Process.Cmd.Env = append(os.Environ())
	if e.Lang == "python" {
		e.Process.Cmd.Env = append(e.Process.Cmd.Env, "PYTHONUNBUFFERED=1")
	}
	e.ProcessContent = [][]rune{}
	e.ProcessPanelScroll = 0
	e.ProcessPanelSpacing = 2

	e.Process.Start()

	go func() {
		for range e.Process.Updates {
			newLines := e.Process.GetLines(len(e.ProcessContent))
			for _, line := range newLines {
				e.ProcessContent = append(e.ProcessContent, []rune(line))
			}

			e.DrawProcessPanel()
			e.Screen.Show()

			if e.Process.IsStopped() {
				exitCode := e.Process.GetExitCode()
				SeparatorStyle = tcell.StyleDefault.Foreground(197 )
				if exitCode == 0 { SeparatorStyle = tcell.StyleDefault.Foreground(tcell.GetColor("#90EE90")) }

				for i := 0; i < e.COLUMNS-7; i++ { e.Screen.SetContent(i, e.ROWS, '─', nil, SeparatorStyle) }
				e.Screen.Show()
			}
		}
	}()

}