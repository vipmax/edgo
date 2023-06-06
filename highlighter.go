package main

import (
	"fmt"
	"github.com/alecthomas/chroma"
	"github.com/alecthomas/chroma/lexers"
	"github.com/gdamore/tcell"
	"os"
)

type Highlighter struct {
}

func (h Highlighter) colorize(code string, filename string) [][]int {
	//start := time.Now()
	//defer log.Printf("Time taken: %s", time.Since(start))

	// Get the lexer for Go language.
	lexer := lexers.Match(filename)
	if lexer == nil {
		lexer = lexers.Fallback
		//os.Exit(1)
	}

	// Get the iterator for tokenizing the code.
	iterator, err := lexer.Tokenise(nil, code)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Tokenization error: %v", err)
		os.Exit(1)
	}

	tokensIntoLines := chroma.SplitTokensIntoLines(iterator.Tokens())
	textColors := [][]int{}

	for _, tokens := range tokensIntoLines {
		lineColors := []int{}
		for _, token := range tokens {
			//fmt.Printf("Line %d\nToken: %s\nType: %s\n\n", i+1, token.Value, token.Type)
			for range token.Value {
				lineColors = append(lineColors, int(getColor(token.Type)))
			}
		}
		textColors = append(textColors, lineColors)
	}

	return textColors
}

func getColor(tokenType chroma.TokenType) tcell.Color {
	switch tokenType {
	case chroma.Keyword, chroma.KeywordNamespace:
		return tcell.ColorHotPink
	case chroma.KeywordType, chroma.KeywordDeclaration, chroma.KeywordReserved, chroma.Name, chroma.NameTag, chroma.NameFunction:
		return tcell.ColorAquaMarine
	case chroma.String, chroma.StringChar, chroma.Literal, chroma.LiteralStringDouble:
		return tcell.ColorLightGreen
	case chroma.CommentSingle, chroma.CommentMultiline:
		return tcell.ColorDimGray
	case chroma.NumberInteger:
		return tcell.ColorDeepSkyBlue
	default:
		return tcell.ColorWhite // Default color
	}
}