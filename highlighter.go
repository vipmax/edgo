package main

import (
	"github.com/alecthomas/chroma"
	"github.com/alecthomas/chroma/lexers"
	"github.com/gdamore/tcell"
	"os"
	"strings"
	"time"
)

type Highlighter struct {
	logger Logger
}

var edgo, _ = chroma.NewStyle("edgo", chroma.StyleEntries{
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

var darcula, _ = chroma.NewStyle("darcula", chroma.StyleEntries{
	chroma.Comment: "#707070",

	//chroma.NameConstant: "#7A9EC2",

	chroma.Keyword: "#CC8242",
	chroma.KeywordNamespace: "#CC8242",

	chroma.String: "#6A8759",
	chroma.LiteralStringDouble: "#6A8759",
	chroma.Literal: "#6A8759",
	chroma.StringChar: "#6A8759",

	chroma.KeywordType: "#CC8242",
	chroma.KeywordDeclaration: "#CC8242",
	chroma.KeywordReserved: "#CC8242",
	//chroma.NameTag: "#FFC66D",
	//chroma.NameFunction: "#FFC66D",
	//
	chroma.NumberInteger: "#7A9EC2",

	//chroma.NameFunction: "#FFC66D",
	chroma.NameFunction: "#AD9E7E",
})



var theme = edgo
//var theme = darcula
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
	h.logger.info("colorize start")
	start := time.Now()

	// get lexer depending on filename
	lexer := lexers.Match(filename)
	if lexer == nil { lexer = lexers.Fallback }

	// get iterator for tokenizing the code
	iterator, err := lexer.Tokenise(nil, code)
	if err != nil {
		h.logger.info("tokenization error: " + err.Error())
		os.Exit(1)
	}

	tokensIntoLines := chroma.SplitTokensIntoLines(iterator.Tokens())
	textColors := [][]int{}

	for _, tokens := range tokensIntoLines {
		lineColors := []int{}
		for _, token := range tokens {
			chromeColor := theme.Get(token.Type).Colour.String()
			tcellColor := tcell.GetColor(chromeColor)
			color := int(tcellColor)
			if color == -1 { color = 15 } // sometimes it returs -1 and cursor is black, make it write

			// copy color for each token character
			for range token.Value { lineColors = append(lineColors, color) }
		}
		textColors = append(textColors, lineColors)
	}

	h.logger.info("colorize end, elapsed: " + time.Since(start).String())
	return textColors
}