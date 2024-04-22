package runner

import (
	. "github.com/smacker/go-tree-sitter"

)

type Run interface {
	Query() string
	Find(tfinder *RunQueryFinder, root *Node, filename string, code []byte) map[int]RunData
	Run(data RunData) []string
}

func GetRunnerByLang(lang string, filepath string) Run {
	switch lang {
	case "go": return &GoRun{}
	default:
	}

	return nil
}


type RunData struct {
	Name string
	Filename string
	Line int
}

type RunQueryFinder struct {
	Query *Query
	Lang  string
}


func (this *RunQueryFinder) Find(root *Node, filename string, code []byte) map[int]RunData {
	results := make(map[int]RunData)

	qc := NewQueryCursor()
	qc.Exec(this.Query, root)

	for {
		m, ok := qc.NextMatch()
		if !ok { break }
		m = qc.FilterPredicates(m, code)

		for i := range m.Captures {
			c := m.Captures[i]; node := c.Node;
			nodename := this.Query.CaptureNameForId(c.Index)
			content := node.Content(code)
			isTestFound := nodename == "main-name"
			if isTestFound {
				line := int(node.StartPoint().Row)
				results[line] = RunData{
					Name: content,
					Filename: filename,
					Line: line,
				}
			}
		}
	}

	return results
}