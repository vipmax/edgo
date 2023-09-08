package highlighter

import (
	. "edgo/internal/utils"
	"fmt"
	"strings"
	"testing"
	"time"
)

func TestTreeSitterHighlighter(t *testing.T) {
	treeSitterHighlighter := New()
	treeSitterHighlighter.SetLang("javascript")
	code := `
function hello() {
	console.log('hello') 
}
`
	start := time.Now()
	colors := treeSitterHighlighter.Colorize(code)
	fmt.Println("colorized, elapsed", time.Since(start))

	for i, colorsLine := range colors {
		fmt.Println(i, "line", colorsLine)
	}
}

func TestTreeSitterHighlighterFromFile(t *testing.T) {
	treeSitterHighlighter := New()
	treeSitterHighlighter.SetLang("go")

	code, _ := ReadFileToString("../editor/editor.go")

	start := time.Now()
	colors := treeSitterHighlighter.Colorize(code)
	fmt.Println("colorized, elapsed", time.Since(start))

	for i, colorsLine := range colors {
		fmt.Println(i+1, "line", colorsLine)
	}
}


func TestTreeSitterHighlighterEdit(t *testing.T) {
	treeSitterHighlighter := New()
	treeSitterHighlighter.SetLang("javascript")
	code := `function hello() {
	console.log('hello') 
}
`
	start := time.Now()
	colors := treeSitterHighlighter.Colorize(code)
	fmt.Println("colorized, elapsed", time.Since(start))

	codes := strings.Split(code, "\n")
	for i, colorsLine := range colors {
		fmt.Println(i+1, codes[i])
		fmt.Println(i+1, colorsLine)
	}

	code = `function hello() {
	console.log('hello world') 
}
`
	treeSitterHighlighter.AddMultipleCharEdit(code,1, 20,1,21)
	treeSitterHighlighter.ColorizeRange(code,0,0,2,0)

	codes = strings.Split(code, "\n")
	for i, colorsLine := range treeSitterHighlighter.Colors {
		fmt.Println(i+1, codes[i])
		fmt.Println(i+1, colorsLine)
	}
}