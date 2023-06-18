package main

import (
	"fmt"
	"github.com/alecthomas/chroma"
	"github.com/alecthomas/chroma/lexers"
//	"github.com/alecthomas/chroma/styles"
	"math"
	"os"
	"strings"
	"time"
)

type Highlighter struct {
	logger Logger
}

var defaultStyle, _ = chroma.NewStyle("edgo", chroma.StyleEntries{
	chroma.Comment: "#a8a8a8",

	chroma.Keyword: "#FF69B4",
	chroma.KeywordNamespace: "#FF69B4",

	chroma.String: "#90EE90",
	chroma.LiteralStringDouble: "#90EE90",
	chroma.Literal: "#90EE90",
	chroma.StringChar: "#90EE90",

	chroma.KeywordType: "#7FFFD4",
	chroma.KeywordDeclaration: "#7FFFD4",
	chroma.KeywordReserved: "#7FFFD4",
	chroma.NameTag: "#7FFFD4",
	chroma.NameFunction: "#7FFFD4",

	chroma.NumberInteger: "#00BFFF",
})

var theme = defaultStyle
//var theme = styles.Get("dracula")
//var theme = styles.Get("nord")
//var theme = styles.Get("monokai")
//var theme = styles.Get("paraiso-dark")
//var theme = styles.Get("vulcan")
//var theme = styles.Get("witchhazel")
//var theme = styles.Get("xcode-dark")


func detectLang(filename string) string {
	lexer := lexers.Match(filename)
	if lexer == nil { return "" }
	config := lexer.Config()
	if config == nil { return "" }
	return strings.ToLower(config.Name)
}
func (h *Highlighter) colorize(code string, filename string) [][]int {
	start := time.Now()
	defer h.logger.info("[highlighter] colorize elapsed: " + time.Since(start).String())

	// get lexer depending on filename
	lexer := lexers.Match(filename)
	if lexer == nil { lexer = lexers.Fallback }

	// get iterator for tokenizing the code
	iterator, err := lexer.Tokenise(nil, code)
	if err != nil {
		h.logger.info("[highlighter] tokenization error: " + err.Error())
		os.Exit(1)
	}

	tokensIntoLines := chroma.SplitTokensIntoLines(iterator.Tokens())
	textColors := [][]int{}

	for _, tokens := range tokensIntoLines {
		lineColors := []int{}
		for _, token := range tokens {
			color := getColor(token.Type)
			for range token.Value {
				lineColors = append(lineColors, color)
			}
		}
		textColors = append(textColors, lineColors)
	}

	return textColors
}


func getColor(tokenType chroma.TokenType) int {
	colour := theme.Get(tokenType).Colour
	ansi256color := RgbToAnsi256(colour.Red(), colour.Green(), colour.Blue())
	return ansi256color
}

func RgbToAnsi256(r, g, b uint8) int {
	if r == g && g == b {
		if r < 8 { return 16 }
		if r > 248 { return 231 }
		return int(math.Round(float64(r-8)/247*24)) + 232
	}

	ansi := 16 +
		36*int(math.Round(float64(r)/255*5)) +
		6*int(math.Round(float64(g)/255*5)) +
		int(math.Round(float64(b)/255*5))

	return ansi
}

func Ansi256ToRGB(c int) (int, int, int) {
	if c < 16 {
		// handle standard colors
	} else if c < 232 {
		// handle 6x6x6 color cube
		c -= 16
		r := c/36
		c -= r * 36
		g := c/6
		b := c - g * 6
		return r * 51, g * 51, b * 51
	} else {
		// handle grayscale ramp
		c -= 232
		v := c * 10 + 8
		return v, v, v
	}
	return 0, 0, 0
}

func RgbToHex(r, g, b int) string {
	return fmt.Sprintf("#%02X%02X%02X", r, g, b)
}