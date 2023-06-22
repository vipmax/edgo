package main

import (
	"github.com/alecthomas/chroma"
	"github.com/alecthomas/chroma/lexers"
	"github.com/alecthomas/chroma/styles"
	"github.com/gdamore/tcell"
	"os"
	"strings"
	"time"
)

type Highlighter struct {

}

//var theme = IdeaLight	.чфьц er
//var theme = E}

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


var SelectionColor = 246 // gray
var OverlayColor = -1 // transparent
var AccentColor = 303 // pink


func detectLang(filename string) string {
	lexer := lexers.Match(filename)
	if lexer == nil { return "" }
	config := lexer.Config()
	if config == nil { return "" }
	return strings.ToLower(config.Name)
}

func (h *Highlighter) setTheme(name string) {
	theme = styles.Get(name)
	AccentColor = int(tcell.GetColor(theme.Get(chroma.Keyword).Colour.String()))

}

func (h *Highlighter) colorize(code string, filename string) [][]int {
	start := time.Now()

	// get lexer depending on filename
	lexer := lexers.Match(filename)
	if lexer == nil { lexer = lexers.Fallback }

	// get iterator for tokenizing the code
	iterator, err := lexer.Tokenise(nil, code)
	if err != nil {
		logger.info("tokenization error: " + err.Error())
		os.Exit(1)
	}

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

	logger.info("colorize end, elapsed: " + time.Since(start).String())
	return textColors
}