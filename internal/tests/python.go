package tests

import (
	"fmt"
	. "github.com/smacker/go-tree-sitter"
	"strings"
)

type PythonTest struct {

}

func (this *PythonTest) TestQuery() string {
	return`
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

func (this *PythonTest) Find(tfinder *TestFinder, root *Node, filename string, code []byte) map[int]TestData {
	if !strings.Contains(filename, "test") { return nil }
	return tfinder.Find(root, filename, code)
}

func (this *PythonTest) Run(test TestData) []string {
	cmd := fmt.Sprintf(`python3 -m pytest -k %s %s`, test.Name, test.Filename)
	args := strings.Split(cmd, " ")
	return args
}
