package tests

import (
	"context"
	"fmt"
	. "github.com/smacker/go-tree-sitter"
	"github.com/smacker/go-tree-sitter/python"
	"testing"
)

func TestPythonFindTest(t *testing.T) {
	lang := "python"

	code := `import pytest
from random import randint

class TestYo:
    def mess(self, value):
        return randint(0, value)

    def test_pass(self):
        assert 1 == 1

    def test_fail_sometimes(self):
        assert 1 == self.mess(1)

`
	expectedTest := map[int]TestData{
		3: {
			Name:     "TestYo",
			Filename: "test_yo.py",
			Line:     3,
		},
		7: {
			Name:     "test_pass",
			Filename: "test_yo.py",
			Line:     7,
		},
		10: {
			Name:     "test_fail_sometimes",
			Filename: "test_yo.py",
			Line:     10,
		},
	}

	codeBytes := []byte(code)

	pythonTest := PythonTest{}
	query := pythonTest.TestQuery()

	language := python.GetLanguage()
	q, _ := NewQuery([]byte(query), language)

	testFinder := TestFinder{TestQuery: q, Lang: lang}
	node, _ := ParseCtx(context.Background(), codeBytes, language)

	tests := pythonTest.Find(&testFinder, node, "test_yo.py", codeBytes)
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

