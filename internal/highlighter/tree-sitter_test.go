package highlighter

import (
	"context"
	"edgo/internal/utils"
	"fmt"
	sitter "github.com/smacker/go-tree-sitter"
	"github.com/smacker/go-tree-sitter/golang"
	"github.com/smacker/go-tree-sitter/javascript"
	"github.com/smacker/go-tree-sitter/python"
	"testing"
	"time"
)

func walk(node *sitter.Node) {
	// Process the current node here (e.g., print node information).
	fmt.Println(node.StartPoint(), node.EndPoint(), node.Type(), node.String())

	// Visit each child node recursively.
	childCount := int(node.NamedChildCount())
	for i := 0; i < childCount; i++ {
		child := node.NamedChild(i)
		walk(child)
	}
}

func print(node *sitter.Node, code []byte) {
	// Process the current node here (e.g., print node information).
	fmt.Println(node.StartPoint(), node.EndPoint(), node.Type(), node.Content(code))

	// Visit each child node recursively.
	childCount := int(node.NamedChildCount())
	for i := 0; i < childCount; i++ {
		child := node.NamedChild(i)
		print(child, code)
	}
}

func TestTreeSitter(t *testing.T) {
	//code, _ := utils.ReadFileToString("tree-sitter_test.go")
	code, _ := utils.ReadFileToString("../editor/editor.go")

	sourceCode := []byte(code)
	//lang := javascript.GetLanguage()
	lang := golang.GetLanguage()
	start := time.Now()
	n, err := sitter.ParseCtx(context.Background(), sourceCode, lang)
	if err != nil { fmt.Println(err)}

	fmt.Println("parsed, elapsed", time.Since(start))

	//fmt.Println(n)

	walk(n)
}

func TestTreeSitterGo(t *testing.T) {
	//code, _ := utils.ReadFileToString("tree-sitter_test.go")
	code := `
package main

import (
	"fmt"
	"time"
)

func main() {
	start := time.Now()
	var count = 0
	for i := 0; i <= 100000000; i++ {
		count += i
		fmt.Println(count)
		time.Sleep(time.Millisecond * 10)
	}
	fmt.Println(count, "elapsed", time.Since(start))
}
`

	sourceCode := []byte(code)
	//lang := javascript.GetLanguage()
	lang := golang.GetLanguage()
	start := time.Now()
	n, err := sitter.ParseCtx(context.Background(), sourceCode, lang)
	if err != nil { fmt.Println(err)}

	fmt.Println("parsed, elapsed", time.Since(start))

	//fmt.Println(n)

	walk(n)
}

func TestTreeSitterPython(t *testing.T) {
	//code, _ := utils.ReadFileToString("tree-sitter_test.go")
	code := `
import time

print("starting")
start_time = time.time()

for i in range(100000):
    print(i)
    time.sleep(0.01)

print("done")

elapsed_time = time.time() - start_time
print("Elapsed time:", elapsed_time, q"seconds")

`

	sourceCode := []byte(code)
	lang := python.GetLanguage()
	start := time.Now()
	n, err := sitter.ParseCtx(context.Background(), sourceCode, lang)
	if err != nil { fmt.Println(err)}

	fmt.Println("parsed, elapsed", time.Since(start))

	//fmt.Println(n)

	walk(n)
}


func TestTreeSitterJs(t *testing.T) {
	code := []byte(`
function hello() { 
	console.log('hello') 
}
`)
	parser := sitter.NewParser()
	parser.SetLanguage(javascript.GetLanguage())

	start := time.Now()
	tree, err := parser.ParseCtx(context.Background(),nil, code)
	if err != nil { fmt.Println(err)}

	fmt.Println("parsed, elapsed", time.Since(start))
	print(tree.RootNode(), code)
}



func TestTreeSitterJsEdit(t *testing.T) {
	parser := sitter.NewParser()
	parser.SetLanguage(javascript.GetLanguage())

	code := []byte(`function hello() { console.log('hello') }`)
	oldEndIndex := uint32(len(code))

	start := time.Now()

	tree, err := parser.ParseCtx(context.Background(),nil, code)
	if err != nil { fmt.Println(err)}

	elapsedFirst := time.Since(start)
	fmt.Println("parsed, elapsed", elapsedFirst)
	print(tree.RootNode(), code)

	fmt.Println("Edit input")

	code = []byte(`function hello2() { console.log('hello') }`)
	newEndIndex := uint32(len(code))

	tree.Edit(sitter.EditInput{
		StartIndex:  14,
		OldEndIndex: oldEndIndex,
		NewEndIndex: newEndIndex,
		StartPoint: sitter.Point{Row: 0, Column: 14 },
		OldEndPoint: sitter.Point{Row: 0, Column: 14},
		NewEndPoint: sitter.Point{Row: 0, Column: 15},
	})

	start = time.Now()
	tree, err = parser.ParseCtx(context.Background(), tree, code)
	if err != nil { fmt.Println(err)}
	elapsedSecond := time.Since(start)
	fmt.Println("parsed again, elapsed", elapsedSecond)
	speedup := float64(elapsedFirst) / float64(elapsedSecond)
	fmt.Printf("Speedup factor: %.2f\n", speedup)

	print(tree.RootNode(), code)
}


func TestTreeSitterJsEditDelete(t *testing.T) {
	parser := sitter.NewParser()
	parser.SetLanguage(javascript.GetLanguage())

	code := []byte(`function hello() { console.log('hello') }`)
	oldEndIndex := uint32(len(code))

	start := time.Now()

	tree, err := parser.ParseCtx(context.Background(),nil, code)
	if err != nil { fmt.Println(err)}

	elapsedFirst := time.Since(start)
	fmt.Println("parsed, elapsed", elapsedFirst)
	print(tree.RootNode(), code)

	fmt.Println("Edit input")

	code = []byte(`function hel() { console.log('hello') }`)
	newEndIndex := uint32(len(code))

	tree.Edit(sitter.EditInput{
		StartIndex:  14,
		OldEndIndex: oldEndIndex,
		NewEndIndex: newEndIndex,
		StartPoint: sitter.Point{Row: 0, Column: 14 },
		OldEndPoint: sitter.Point{Row: 0, Column: oldEndIndex},
		NewEndPoint: sitter.Point{Row: 0, Column: newEndIndex},
	})

	start = time.Now()
	tree, err = parser.ParseCtx(context.Background(), tree, code)
	if err != nil { fmt.Println(err)}
	elapsedSecond := time.Since(start)
	fmt.Println("parsed again, elapsed", elapsedSecond)
	speedup := float64(elapsedFirst) / float64(elapsedSecond)
	fmt.Printf("Speedup factor: %.2f\n", speedup)

	print(tree.RootNode(), code)
}

func TestTreeSitterJsEditDeleteMultiple(t *testing.T) {
	parser := sitter.NewParser()
	parser.SetLanguage(javascript.GetLanguage())

	code := []byte(`function hello() { console.log('hello') }`)
	oldEndIndex := uint32(len(code))

	start := time.Now()

	tree, err := parser.ParseCtx(context.Background(),nil, code)
	if err != nil { fmt.Println(err)}

	elapsedFirst := time.Since(start)
	fmt.Println("parsed, elapsed", elapsedFirst)
	print(tree.RootNode(), code)

	fmt.Println("Edit input")

	code = []byte(`function hel() { console.log('hello') }`)
	newEndIndex := uint32(len(code))

	tree.Edit(sitter.EditInput{
		StartIndex:  14, OldEndIndex: oldEndIndex, NewEndIndex: newEndIndex,
		StartPoint: sitter.Point{Row: 0, Column: 14 },
		OldEndPoint: sitter.Point{Row: 0, Column: oldEndIndex},
		NewEndPoint: sitter.Point{Row: 0, Column: newEndIndex},
	})

	start = time.Now()
	tree, err = parser.ParseCtx(context.Background(), tree, code)
	if err != nil { fmt.Println(err)}
	elapsedSecond := time.Since(start)
	fmt.Println("parsed again, elapsed", elapsedSecond)
	speedup := float64(elapsedFirst) / float64(elapsedSecond)
	fmt.Printf("Speedup factor: %.2f\n", speedup)
	print(tree.RootNode(), code)

	fmt.Println("Edit input")

	code = []byte(`function h() { console.log('hello') }`)
	newEndIndex = uint32(len(code))

	tree.Edit(sitter.EditInput{
		StartIndex:  10, OldEndIndex: oldEndIndex, NewEndIndex: newEndIndex,
		StartPoint: sitter.Point{Row: 0, Column: 10 },
		OldEndPoint: sitter.Point{Row: 0, Column: oldEndIndex},
		NewEndPoint: sitter.Point{Row: 0, Column: newEndIndex},
	})

	start = time.Now()
	tree, err = parser.ParseCtx(context.Background(), tree, code)
	if err != nil { fmt.Println(err)}
	elapsedSecond = time.Since(start)
	fmt.Println("parsed again, elapsed", elapsedSecond)
	speedup = float64(elapsedFirst) / float64(elapsedSecond)
	fmt.Printf("Speedup factor: %.2f\n", speedup)
	print(tree.RootNode(), code)
}


func TestTreeSitterJsEditEnter(t *testing.T) {
	parser := sitter.NewParser()
	parser.SetLanguage(javascript.GetLanguage())

	code := []byte(`function hello() { console.log('hello') }`)
	oldEndIndex := uint32(len(code))

	start := time.Now()

	tree, err := parser.ParseCtx(context.Background(),nil, code)
	if err != nil { fmt.Println(err)}

	elapsedFirst := time.Since(start)
	fmt.Println("parsed, elapsed", elapsedFirst)
	print(tree.RootNode(), code)

	fmt.Println("Edit input")

	code = []byte(`function hello() { 
console.log('hello')}`)
	newEndIndex := uint32(len(code))

	tree.Edit(sitter.EditInput{
		StartIndex:  19,
		OldEndIndex: oldEndIndex,
		NewEndIndex: newEndIndex,
		StartPoint: sitter.Point{Row: 1, Column: 0 },
		OldEndPoint: sitter.Point{Row: 0, Column: oldEndIndex},
		NewEndPoint: sitter.Point{Row: 1, Column: newEndIndex},
	})

	start = time.Now()
	tree, err = parser.ParseCtx(context.Background(), tree, code)
	if err != nil { fmt.Println(err)}
	elapsedSecond := time.Since(start)
	fmt.Println("parsed again, elapsed", elapsedSecond)
	speedup := float64(elapsedFirst) / float64(elapsedSecond)
	fmt.Printf("Speedup factor: %.2f\n", speedup)

	print(tree.RootNode(), code)
}



func TestTreeSitterQuery(t *testing.T) {
	code := []byte(`
function hello() { 
	// comment line 
	console.log('hello') 
	if (true) { console.log('true') }
	return "value"
}
`)

	query := `
[
  "function"
  "if"
  "return"
] @keyword

(comment) @comment
`
	lang := javascript.GetLanguage()
	start := time.Now()
	q, _ := sitter.NewQuery([]byte(query), lang)
	qc := sitter.NewQueryCursor()

	n, _ := sitter.ParseCtx(context.Background(), code, lang)
	fmt.Println("parsed , elapsed", time.Since(start))

	// Execute the query

	qc.Exec(q, n)
	fmt.Println("query exec, elapsed", time.Since(start))

	for {
		m, ok := qc.NextMatch()
		if !ok { break }
		m = qc.FilterPredicates(m, code)
		for _, c := range m.Captures {
			name := q.CaptureNameForId(c.Index)
			content := c.Node.Content(code)
			fmt.Println(c.Node.StartPoint(), c.Node.EndPoint(), name, c.Node.Type(), content)
		}
	}
}

func TestTreeSitterQuery2(t *testing.T) {
	sourceCode := []byte(`
package main
import "fmt"
func main() {
	// comment line 
	fmt.Println("Hello, world!")
}
`)

	// Query with predicates
	query := `
(identifier) @keyword
`
	// Parse source code
	lang := golang.GetLanguage()
	n, _ := sitter.ParseCtx(context.Background(), sourceCode, lang)

	// Execute the query
	q, _ := sitter.NewQuery([]byte(query), lang)
	qc := sitter.NewQueryCursor()
	qc.Exec(q, n)

	// Iterate over query results
	for {
		m, ok := qc.NextMatch()
		if !ok { break }
		// Apply predicates filtering
		m = qc.FilterPredicates(m, sourceCode)
		for _, c := range m.Captures {
			fmt.Println(c.Node.StartPoint(), c.Node.EndPoint(), c.Node.Type(), c.Node.Content(sourceCode))
		}
	}
}
