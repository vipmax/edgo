package runner

import (
	"context"
	"fmt"
	. "github.com/smacker/go-tree-sitter"
	"github.com/smacker/go-tree-sitter/golang"
	"testing"
)

func TestGoFindTest(t *testing.T) {
	filename := "main.go"

	lang := "go"

	code := `
package main

import (
	"fmt"
	"time"
)

func main() {
	start := time.Now()
	var count = 0
	for i := 0; i <= 10000000; i++ {
		count += i
	}
	
	fmt.Println(count, "elapsed", time.Since(start))
}

func main2() int { return 1 }
`

	expectedTest := map[int]RunData{
		8: {
			Name:     "main",
			Filename: filename,
			Line:     8,
		},
	}

	codeBytes := []byte(code)

	run := GoRun{}
	query := run.Query()

	language := golang.GetLanguage()
	q, _ := NewQuery([]byte(query), language)

	testFinder := RunQueryFinder{Query: q, Lang: lang}
	node, _ := ParseCtx(context.Background(), codeBytes, language)

	tests := run.Find(&testFinder, node, filename, codeBytes)
	fmt.Println(tests)

	if tests == nil { t.Errorf("tests cant be nil this case") }
	if len(tests) != len(expectedTest) {
		t.Errorf("tests must be same size %d %d", len(tests), len(expectedTest))
	}

	for line, expected := range expectedTest {
		actual, found := tests[line]
		if !found {
			t.Errorf("Expected test on line %d, but not found", line)
			continue
		}

		if actual != expected {
			t.Errorf("Expected test on line %d to be %v, but got %v", line, expected, actual)
		}
	}
}

