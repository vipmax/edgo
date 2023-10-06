package tests

import (
	. "github.com/smacker/go-tree-sitter"
)

type Test interface {
	Query() string
	Find(tfinder *TestFinder, root *Node, filename string, code []byte) map[int]TestData
	Run(test TestData) []string
}

type TestData struct {
	Name string
	Filename string
	Line int
}

type TestFinder struct {
	TestQuery *Query
	Lang      string
}


func (this *TestFinder) Find(root *Node, filename string, code []byte) map[int]TestData {
	tests := make(map[int]TestData)

	qc := NewQueryCursor()
	qc.Exec(this.TestQuery, root)

	for {
		m, ok := qc.NextMatch()
		if !ok { break }
		m = qc.FilterPredicates(m, code)
		for i := range m.Captures {
			c := m.Captures[i]; node := c.Node;
			nodename := this.TestQuery.CaptureNameForId(c.Index)
			content := node.Content(code)
			isTestFound := nodename == "test-name"
			if isTestFound {
				line := int(node.StartPoint().Row)
				tests[line] = TestData{
					Name: content,
					Filename: filename,
					Line: line,
				}
			}
		}
	}

	return tests
}
