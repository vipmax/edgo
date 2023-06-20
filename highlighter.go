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

//var theme = IdeaLight
//var theme = EdgoLight
var theme = EdgoDark
//var theme = Darcula
//var theme = styles.Get("edgo")
//var theme = styles.Get("dracula")
//var theme = styles.Get("nord")
//var theme = styles.Get("monokai")
//var theme = styles.Get("paraiso-dark")
//var theme = styles.Get("vulcan")
//var theme = styles.Get("witchhazel")
//var theme = styles.Get("xcode-dark")


var SelectionColor = 7
var OverlayColor = -1
var AccentColor = 303


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

	AccentColor = int(tcell.GetColor(theme.Get(chroma.Keyword).Colour.String()))

	tokensIntoLines := chroma.SplitTokensIntoLines(iterator.Tokens())
	textColors := [][]int{}

	for _, tokens := range tokensIntoLines {
		lineColors := []int{}
		for _, token := range tokens {
			chromaColor := theme.Get(token.Type).Colour.String()
			tcellColor := tcell.GetColor(chromaColor)
			color := int(tcellColor)
			// copy color for each token character
			for range token.Value { lineColors = append(lineColors, color) }
		}
		textColors = append(textColors, lineColors)
	}

	h.logger.info("colorize end, elapsed: " + time.Since(start).String())
	return textColors
}