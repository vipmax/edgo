package highlighter

import (
	. "edgo/internal/utils"
	"fmt"
	"github.com/alecthomas/chroma"
	"github.com/alecthomas/chroma/styles"
	"github.com/go-enry/go-enry/v2"
	"os"
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

	//col = ColorFromString("#c6a5fc")
	//fmt.Println(col)
	//
	//col = ColorFromString("#d992f9")
	//fmt.Println(col)
	//
	//col = ColorFromString("#f78bff")
	//fmt.Println(col)
}

func TestLangDetect(t *testing.T) {
	file := "highlighter_test.go"
	lang := DetectLang(file)
	fmt.Println(lang)
}

func TestLangDetect2(t *testing.T) {
	file := "highlighter_test.go"
	lang, safe := enry.GetLanguageByExtension(file)
	fmt.Println(lang, safe)
}

func BenchmarkLangDetect(b *testing.B) {
	for i := 0; i < b.N; i++ {
		file := "highlighter_test.go"
		DetectLang(file)
		//fmt.Println(lang)
	}
	// BenchmarkLangDetect-8   	     417	   2869136 ns/op
}

func BenchmarkLangDetect2(b *testing.B) {
	for i := 0; i < b.N; i++ {
		file := "highlighter_test.go"
		enry.GetLanguageByExtension(file)
	}
	// BenchmarkLangDetect2-8   	27731184	        42.21 ns/op
}

func BenchmarkLangDetect3(b *testing.B) {
	for i := 0; i < b.N; i++ {
		filename := "highlighter_test.go"
		content, _ := os.ReadFile(filename)
		enry.GetLanguage(filename, content)
	}
	// BenchmarkLangDetect3-8   	   55464	     21176 ns/op
}


func BenchmarkLangDetectReal(b *testing.B) {
	filename := "highlighter_test.go"
	for i := 0; i < b.N; i++ {
		language, _ := enry.GetLanguageByExtension(filename)
		if language == "" { language, _ = enry.GetLanguageByFilename(filename) }

		info, _ := enry.GetLanguageInfo(language)
		Use(info)
	}
	// BenchmarkLangDetectReal-8   	11881623	       100.2 ns/op
}