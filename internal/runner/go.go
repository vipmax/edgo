package runner

import (
	"fmt"
	. "github.com/smacker/go-tree-sitter"
	"strings"
)

type GoRun struct {

}

func (this *GoRun) Query() string {
	return`
         (
          function_declaration name: (identifier) @main-name
          (#eq? @main-name "main")
         )
`
}

func (this *GoRun) Find(runQueryFinder *RunQueryFinder, root *Node, filename string, code []byte) map[int]RunData {
	if !strings.HasSuffix(filename, ".go") { return nil }
	return runQueryFinder.Find(root, filename, code)
}

func (this *GoRun) Run(test RunData) []string {
	cmd := fmt.Sprintf("go run %s", test.Filename)
	args := strings.Split(cmd, " ")
	return args
}

