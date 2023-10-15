package tests

import (
	"context"
	"fmt"
	. "github.com/smacker/go-tree-sitter"
	"github.com/smacker/go-tree-sitter/javascript"
	"testing"
)

func TestJavascriptFindTest(t *testing.T) {
	lang := "javascript"

	code := `
function sum(a, b) {
    return a + b;
}


describe("math tests", () => {

    it("positive", () => {
        expect(sum(1, 1)).toBe(2);
    });

    it("negative", () => {
        expect(sum(-1, -1)).toBe(-2);
    });
    
    it("failed", () => {
        expect(sum(-1, -1)).toBe(2);
    });
   
});
`
	expectedTest := map[int]TestData{
		6: {
			Name:     `"math tests"`,
			Filename: "test.js",
			Line:     6,
		},
		8: {
			Name:     `"positive"`,
			Filename: "test.js",
			Line:     8,
		},
		12: {
			Name:     `"negative"`,
			Filename: "test.js",
			Line:     12,
		},
		16: {
			Name:     `"failed"`,
			Filename: "test.js",
			Line:     16,
		},
	}

	codeBytes := []byte(code)

	test := JavascriptTest{}
	query := test.TestQuery()

	language := javascript.GetLanguage()
	q, _ := NewQuery([]byte(query), language)

	testFinder := TestFinder{TestQuery: q, Lang: lang}
	node, _ := ParseCtx(context.Background(), codeBytes, language)

	tests := test.Find(&testFinder, node, "test.js", codeBytes)
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

