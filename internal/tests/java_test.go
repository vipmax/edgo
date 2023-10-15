package tests

import (
	"context"
	"fmt"
	. "github.com/smacker/go-tree-sitter"
	"github.com/smacker/go-tree-sitter/java"
	"testing"
)

func TestJavaFindTest(t *testing.T) {
	lang := "javascript"

	code := `
import static org.junit.jupiter.api.Assertions.assertEquals;

import calc.Calculator;
import org.junit.jupiter.api.Test;


class MyFirstJUnitJupiterTests {

    private final Calculator calculator = new Calculator();

    @Test
    void addition() {
        System.out.println("test");
        assertEquals(2, calculator.add(1, 1));
    }
	
	void simpleFunction() {
        System.out.println("hi");
	}
}
`
	expectedTest := map[int]TestData{
		12: {
			Name:     `addition`,
			Filename: "test.java",
			Line:     12,
		},

	}

	codeBytes := []byte(code)

	test := JavaTest{}
	query := test.TestQuery()

	language := java.GetLanguage()
	q, _ := NewQuery([]byte(query), language)

	testFinder := TestFinder{TestQuery: q, Lang: lang}
	node, _ := ParseCtx(context.Background(), codeBytes, language)

	tests := test.Find(&testFinder, node, "test.java", codeBytes)
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

