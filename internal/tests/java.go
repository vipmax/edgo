package tests

import (
	"fmt"
	. "github.com/smacker/go-tree-sitter"
	"strings"
)

type JavaTest struct {

}

func (this *JavaTest) TestQuery() string {
	return`
        (method_declaration
		  (modifiers 
			(marker_annotation
				name: (identifier) @string))
		  name: (identifier) @test-name
		)
`
}

func (this *JavaTest) Find(tfinder *TestFinder, root *Node, filename string, code []byte) map[int]TestData {
	if !strings.Contains(filename, "test") { return nil }
	return tfinder.Find(root, filename, code)
}

func (this *JavaTest) Run(test TestData) []string {
	cmd := fmt.Sprintf("mvn test -Dtest=*#%s", test.Name)
	args := strings.Split(cmd, " ")
	return args
}
