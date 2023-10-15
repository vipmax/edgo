package tests

import (
	. "github.com/smacker/go-tree-sitter"
	"strings"
)

type JavascriptTest struct {

}

func (this *JavascriptTest) TestQuery() string {
	return`
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

func (this *JavascriptTest) Find(tfinder *TestFinder, root *Node, filename string, code []byte) map[int]TestData {
	if !strings.Contains(filename, "test") { return nil }
	return tfinder.Find(root, filename, code)
}

func (this *JavascriptTest) Run(test TestData) []string {
	test.Name = strings.ReplaceAll(test.Name, "'", "")
	test.Name = strings.ReplaceAll(test.Name, "\"", "")
	//cmd = fmt.Sprintf(`node node_modules/jest/bin/jest.js %s -t '%s'`, test.Filename, test.Name)
	args := []string { "node", "node_modules/jest/bin/jest.js", test.Filename, "-t", test.Name}
	return args
}
