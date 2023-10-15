package tests

import (
	"fmt"
	. "github.com/smacker/go-tree-sitter"
	"path/filepath"
	"strings"
)

type GoTest struct {

}

func (this *GoTest) TestQuery() string {
	return`
         (
          function_declaration name: (identifier) @test-name
          (#match? @test-name "Test*")
        )
`
}

func (this *GoTest) Find(tfinder *TestFinder, root *Node, filename string, code []byte) map[int]TestData {
	if !strings.HasSuffix(filename, "test.go") { return nil }
	return tfinder.Find(root, filename, code)
}

func (this *GoTest) Run(test TestData) []string {
	dir := filepath.Dir(test.Filename)
	cmd := fmt.Sprintf("go test -timeout 30s -run ^%s$ -test.v %s", test.Name, dir)
	args := strings.Split(cmd, " ")
	return args
}
