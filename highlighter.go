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

func (h Highlighter) colorize(code string) [][]int {
	//start := time.Now()
	//defer log.Printf("Time taken: %s", time.Since(start))

	// Get the lexer for Go language.
	lexer := lexers.Get("go")
	if lexer == nil {
		fmt.Fprintln(os.Stderr, "Lexer not found")
		os.Exit(1)
	}

	// Get the iterator for tokenizing the code.
	iterator, err := lexer.Tokenise(nil, code)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Tokenization error: %v", err)
		os.Exit(1)
	}

	tokensIntoLines := chroma.SplitTokensIntoLines(iterator.Tokens())
	colors := [][]int{}

	for _, tokens := range tokensIntoLines {
		color := []int{}
		for _, token := range tokens {
			//fmt.Printf("Line %d\nToken: %s\nType: %s\n\n", i+1, token.Value, token.Type)
			for j := range token.Value {
				j = j + 1
				var c = 0

				if token.Type == chroma.Keyword {
					c = int(tcell.ColorHotPink)
				}
				if token.Type == chroma.KeywordType {
					c = int(tcell.ColorAquaMarine)
				}
				if token.Type == chroma.KeywordDeclaration {
					c = int(tcell.ColorAquaMarine)
				}
				if token.Type == chroma.KeywordNamespace {
					c = int(tcell.ColorHotPink)
				}
				if token.Type == chroma.Name {
					c = int(tcell.ColorAquaMarine)
				}
				if token.Type == chroma.String {
					c = int(tcell.ColorLightGreen)
				}
				if token.Type == chroma.StringChar {
					c = int(tcell.ColorLightGreen)
				}
				if token.Type == chroma.CommentSingle {
					c = int(tcell.ColorDimGray)
				}
				if token.Type == chroma.NumberInteger {
					c = int(tcell.ColorDeepSkyBlue)
				}
				if token.Type == chroma.NameFunction {
					c = int(tcell.ColorAquaMarine)
				}

				color = append(color, c)
			}
		}
		colors = append(colors, color)
	}

	return colors
}
