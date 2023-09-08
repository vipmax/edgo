package highlighter

import (
	. "edgo/internal/utils"
	"fmt"
	"github.com/alecthomas/chroma"
	"github.com/alecthomas/chroma/styles"
	"testing"
	"time"
)

func TestColorize(t *testing.T) {
	file := "highlighter_test.go"
	filecontent, _ := ReadFileToString(file)

	h := Highlighter{}
	characterColors := h.Colorize(filecontent, file)

	for i, color := range characterColors {
		fmt.Println(i+1, color)
	}
}

func TestColorizeJs(t *testing.T) {
	file := "highlighter_test.js"
	code := `
function hello() { 
	console.log('hello') 
}
`

	h := Highlighter{}

	start := time.Now()
	characterColors := h.Colorize(code, file)
	fmt.Println("colorized, elapsed", time.Since(start))

	for _, color := range characterColors {
		fmt.Println(color)
	}
}


func TestGetStyle(t *testing.T) {
	style := styles.Get("github")
	if style == nil { style = styles.Fallback }
	fmt.Println(style)
	fmt.Println(style.Get(chroma.Comment).Background)
}

func TestGetColor(t *testing.T) {
	col := ColorFromString("#fc9994")
	fmt.Println(col)

	col = ColorFromString("#c6a5fc")
	fmt.Println(col)

	col = ColorFromString("#d992f9")
	fmt.Println(col)
}
