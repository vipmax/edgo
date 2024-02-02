package highlighter

import (
	. "edgo/internal/utils"
	"fmt"
	"testing"
	"time"
)


func TestTreeSitterHighlighterColors(t *testing.T) {
	treeSitterHighlighter := NewTreeSitter()
	treeSitterHighlighter.SetLang("go")
	code := `
function hello() {
	console.log('hello') 
}
`
	filecode, _ := ReadFileToString("../../internal/ui/editor.go")
	code = filecode

	treeSitterHighlighter.Parse(&code)

	start := time.Now()
	colors := treeSitterHighlighter.ColorRanges(1100, 11150, nil)
	fmt.Println("colorized, elapsed", time.Since(start).Nanoseconds())

	for i, colorsLine := range colors {
		fmt.Println(i, "line", colorsLine)
	}
}






func BenchmarkColorFromString(b *testing.B) {
	h := NewTreeSitter()
	col := h.ParseColor("#fc9994")
	fmt.Println(col)

	want := 33331604
	if col != want {
		b.Errorf("got %v want %v", col, want)
	}
}

func TestColorFromString(t *testing.T) {
	h := NewTreeSitter()
	col := h.ParseColor("#fc9994")
	fmt.Println(col)

	want := 33331604
	if col != want {
		t.Errorf("got %v want %v", col, want)
	}
}