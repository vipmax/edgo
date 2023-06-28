package highlighter

import (
	. "edgo/internal/logger"
	. "edgo/internal/themes"

	"github.com/alecthomas/chroma"
	"github.com/alecthomas/chroma/lexers"
	"github.com/alecthomas/chroma/styles"
	"github.com/gdamore/tcell"
	"os"
	"strings"
	"time"
)

var HighlighterGlobal = Highlighter{}
type Highlighter struct {

}

//var theme = IdeaLight	
//var theme = Edgo

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
var AccentColor2 = 30 // pink

func DetectLang(filename string) string {
	lexer := lexers.Match(filename)
	if lexer == nil { return "" }
	config := lexer.Config()
	if config == nil { return "" }
	return strings.ToLower(config.Name)
}

func (h *Highlighter) SetTheme(name string) {
	theme = styles.Get(name)
	AccentColor = int(tcell.GetColor(theme.Get(chroma.Keyword).Colour.String()))
	AccentColor2 = int(tcell.GetColor(theme.Get(chroma.KeywordType).Colour.String()))
}

func (h *Highlighter) Colorize(code string, filename string) [][]int {
	if code == "" { return [][]int{nil} }

	start := time.Now()

	// get lexer depending on Name
	lexer := lexers.Match(filename)
	if lexer == nil { lexer = lexers.Fallback }

	// get iterator for tokenizing the code
	iterator, err := lexer.Tokenise(nil, code)
	if err != nil {
		Log.Info("tokenization error: " + err.Error())
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

	Log.Info("colorize end, elapsed: " + time.Since(start).String())
	return textColors
}