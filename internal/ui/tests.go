package ui

import (
	"edgo/internal/process"
	"edgo/internal/highlighter"
	"edgo/internal/utils"
	"fmt"
	"github.com/gdamore/tcell"
	sitter "github.com/smacker/go-tree-sitter"
	"os"
	"path/filepath"
	"strings"
)

type Test struct {
	Name string
	Filename string
	Line int
}

type TestLang interface {
	Query() string
}

func QueryGo() string {
	return `
        (
          function_declaration name: (identifier) @test-name
          (#match? @test-name "Test*")
        )
`
}

func QueryPython() string {
	return `
        (
          (class_definition
  			body: (block
			  (function_definition
				name: (identifier) @method @test-name)))
          (#match? @method "test")
        )

        (
          class_definition name: (identifier) @test-name
          (#match? @test-name "[Tt]est")
        )
`
}

func QueryJs() string {
	return `
(expression_statement
    (call_expression
      function: (identifier) @method-name
      (#match? @method-name "^(describe|test|it)")
      arguments: (arguments [
        ((string) @test-name)
        ((template_string) @test-name)
      ]
    )))

`
}

type TestFinder struct {
	testQuery *sitter.Query
	lang string
}

func (this *TestFinder) Find(ts *highlighter.TreeSitterHighlighter, filename string, code []byte) map[int]Test {
	tests := make(map[int]Test)

	qc := sitter.NewQueryCursor()
	qc.Exec(this.testQuery, ts.GetTree().RootNode())

	for {
		m, ok := qc.NextMatch()
		if !ok { break }
		for i := range m.Captures {
			c := m.Captures[i]; node := c.Node;
			nodetype := node.Type()
			nodename := this.testQuery.CaptureNameForId(c.Index)
			content := node.Content(code)
			highlighter.Use(nodetype, nodename)

			//isTestFound := true
			//isTestFound := strings.Contains(strings.ToLower(content), "test")
			isTestFound := nodename == "test-name"

			if isTestFound {
				line := int(node.StartPoint().Row)
				tests[line] = Test{
					Name: content,
					Filename: filename,
					Line: line,
				}
			}
		}
	}

	return tests
}

var tfinder = TestFinder{}

func (e *Editor) FindTests() {
	clear(e.Tests)
	switch e.Lang {
	case "go":
		if tfinder.testQuery == nil || tfinder.lang != e.Lang {
			q, _ := sitter.NewQuery([]byte(QueryGo()), e.treeSitterHighlighter.GetLang())
			tfinder.testQuery = q
			tfinder.lang = e.Lang
		}

	case "python":
		if tfinder.testQuery == nil || tfinder.lang != e.Lang {
			q, _ := sitter.NewQuery([]byte(QueryPython()), e.treeSitterHighlighter.GetLang())
			tfinder.testQuery = q
			tfinder.lang = e.Lang
		}

	case "javascript":
		if tfinder.testQuery == nil || tfinder.lang != e.Lang {
			q, _ := sitter.NewQuery([]byte(QueryJs()), e.treeSitterHighlighter.GetLang())
			tfinder.testQuery = q
			tfinder.lang = e.Lang
		}

	default:
		 return
	}

	codeBytes := []byte(utils.ConvertContentToString(e.Content))

	e.Tests = tfinder.Find(e.treeSitterHighlighter, e.AbsoluteFilePath, codeBytes)
}


func (e *Editor) RunTest(test Test) {
	cmd := ""
	var args []string

	switch e.Lang {
	case "go":
		dir := filepath.Dir(test.Filename)
		cmd = fmt.Sprintf("go test -timeout 30s -run ^%s$ -test.v %s", test.Name, dir)
		args = strings.Split(cmd, " ")

	case "python":
		cmd = fmt.Sprintf(`python3 -m unittest -k %s %s`, test.Name, test.Filename)
		args = strings.Split(cmd, " ")

	case "javascript":
		test.Name = strings.ReplaceAll(test.Name, "'", "")
		test.Name = strings.ReplaceAll(test.Name, "\"", "")
		//cmd = fmt.Sprintf(`node node_modules/jest/bin/jest.js %s -t '%s'`, test.Filename, test.Name)
		args = []string { "node", "node_modules/jest/bin/jest.js", test.Filename, "-t", test.Name}

	default:
		return
	}



	if e.Lang == "" || e.langConf.Cmd == "" { return }

	if e.ProcessPanelHeight == 0 {
		e.ProcessPanelHeight = 10
		e.COLUMNS, e.ROWS = e.Screen.Size()
		e.ROWS -= e.ProcessPanelHeight
	}

	highlighter.ResetSelectionColor()

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
				highlighter.SeparatorStyle = tcell.StyleDefault.Foreground(197 )
				if exitCode == 0 { highlighter.SeparatorStyle = tcell.StyleDefault.Foreground(tcell.GetColor("#90EE90")) }

				for i := 0; i < e.COLUMNS-7; i++ { e.Screen.SetContent(i, e.ROWS, '─', nil, highlighter.SeparatorStyle) }
				e.Screen.Show()
			}
		}
	}()

}