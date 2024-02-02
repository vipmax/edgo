package highlighter

import (
	"context"
	. "edgo/internal/highlighter/langs"
	. "edgo/internal/utils"
	"fmt"
	"github.com/gdamore/tcell"
	sitter "github.com/smacker/go-tree-sitter"
	"github.com/smacker/go-tree-sitter/bash"
	"github.com/smacker/go-tree-sitter/c"
	"github.com/smacker/go-tree-sitter/cpp"
	"github.com/smacker/go-tree-sitter/css"
	"github.com/smacker/go-tree-sitter/golang"
	"github.com/smacker/go-tree-sitter/html"
	"github.com/smacker/go-tree-sitter/java"
	"github.com/smacker/go-tree-sitter/javascript"
	"github.com/smacker/go-tree-sitter/python"
	"github.com/smacker/go-tree-sitter/rust"
	"github.com/smacker/go-tree-sitter/typescript/typescript"
	"github.com/smacker/go-tree-sitter/yaml"
	. "gopkg.in/yaml.v2"
	"log"
	"os"
	"strings"
	"unicode/utf8"
)

type TreeSitterHighlighter struct {
	parser         *sitter.Parser
	tree           *sitter.Tree
	lines          []string
	lang           string
	language       *sitter.Language
	query          *sitter.Query
	colorsMap      map[string]string
	themePath      string
	injectionLangs map[string]*TreeSitterHighlighter
}

func NewTreeSitter() *TreeSitterHighlighter {
	parser := sitter.NewParser()

	return &TreeSitterHighlighter{
		parser: parser,
		tree:   nil,
	}
}


var defaultColors =
`
identifier: "#a5fcd9"
field_identifier: "#a5fcd9"
property_identifier: "#a5fcd9"
property: "#a5fcd9"
string: "#a5fc94"
keyword: "#ec6aad"
constant: "#ec6aad"
number: "#ec6aad"
integer: "#ec6aad"
float: "#ec6aad"
variable.builtin: "#ec6aad"
function: "#afaff9"
function.call: "#afaff9"
method: "#afaff9"
comment: "#767676"
namespace: "#c6a5fc"
type: "#c6a5fc"
tag.attribute: "#c6a5fc"
accent_color: "#ec6aad"
accent_color2: "#a5fcd9"
`

func (h *TreeSitterHighlighter) ParseColor(colour string) int {
	return int(tcell.GetColor(colour))
}

func (h *TreeSitterHighlighter) SetTheme(themePath string) {
	yamlFile, err := os.ReadFile(themePath)
	if err != nil {
		//log.Println("Error reading theme YAML file: %v", err)
		yamlFile = []byte(defaultColors)
	}

	h.themePath = themePath

	err = Unmarshal(yamlFile, &h.colorsMap)
	if err != nil {
		log.Fatalf("Error unmarshaling YAML: %v", err)
	}

	if value, ok := h.colorsMap["accent_color"]; ok {
		AccentColor = h.ParseColor(value)
	}
	if value, ok := h.colorsMap["accent_color2"]; ok {
		AccentColor2 = h.ParseColor(value)
	}
	if value, ok := h.colorsMap["accent_color3"]; ok {
		AccentColor3 = h.ParseColor(value)
	}
}

func GetSitterLang(lang string) *sitter.Language {
	switch lang {
	case "javascript": return javascript.GetLanguage()
	case "typescript": return typescript.GetLanguage()
	case "go": return golang.GetLanguage()
	case "python": return python.GetLanguage()
	case "html": return html.GetLanguage()
	case "css": return css.GetLanguage()
	case "yaml": return yaml.GetLanguage()
	case "rust": return rust.GetLanguage()
	case "bash": return bash.GetLanguage()
	case "c": return c.GetLanguage()
	case "c++": return cpp.GetLanguage()
	case "java":return java.GetLanguage()
	default: return javascript.GetLanguage()
	}
}

func (h *TreeSitterHighlighter) SetLang(lang string) {
	if h.lang == lang { return }
	h.lang = lang
	h.language = GetSitterLang(lang)
	h.parser.SetLanguage(h.language)

	queryLang := MatchQueryLang(h.lang)
	q, err := sitter.NewQuery([]byte(queryLang), h.language)
	if err!= nil { panic(err) }
	h.query = q
}


func (h *TreeSitterHighlighter) matchExpression(expression string, fullexpression string) int {
	if value, ok := h.colorsMap[fullexpression]; ok { return  h.ParseColor(value) }
	if value, ok := h.colorsMap[expression]; ok { return  h.ParseColor(value) }
	return -1
}


/*
	comment for sitter.EditInput
	The StartIndex, OldEndIndex, and NewEndIndex parameters indicate the range of bytes you're modifying
	The StartPoint, OldEndPoint, and NewEndPoint parameters indicate the range of positions (line, column) affected by the edit.
*/

func (h *TreeSitterHighlighter) AddCharEdit(code *string, row int, col int, ch rune) {
	StartIndex := GetStartIndex(code, row, col)
	runeLen := uint32(utf8.RuneLen(ch))

	editInput := sitter.EditInput{
		StartIndex: StartIndex, OldEndIndex: StartIndex,
		NewEndIndex: StartIndex + runeLen,
		StartPoint:  sitter.Point{Row: 0, Column: 0},
		OldEndPoint: sitter.Point{Row: 0, Column: 0},
		NewEndPoint: sitter.Point{Row: 0, Column: 0},
	}
	h.tree.Edit(editInput)
	h.Parse(code)
}

func (h *TreeSitterHighlighter) RemoveCharEdit(code *string, row int, col int, ch rune) {
	StartIndex := GetStartIndex(code, row, col)
	Row := uint32(row); Column := uint32(col)
	runeLen := uint32(utf8.RuneLen(ch))

	editInput := sitter.EditInput{
		StartIndex: StartIndex, OldEndIndex: StartIndex + runeLen, NewEndIndex: StartIndex,
		StartPoint:  sitter.Point{Row: Row, Column: Column},
		OldEndPoint: sitter.Point{Row: Row, Column: Column + runeLen},
		NewEndPoint: sitter.Point{Row: Row, Column: Column},
	}
	h.tree.Edit(editInput)
	h.Parse(code)
}

func GetStartIndex(code *string, row int, col int) uint32 {
	r, c, startIndex := 0, 0, 0
	for _, char := range *code {
		runeLen := utf8.RuneLen(char)

		if r == row && c == col {
			break
		}
		startIndex += runeLen

		if char == '\n' {
			r++
			c = 0
		} else {
			c++
		}
	}
	return uint32(startIndex)
}


func Use(vals ...interface{}) { }



func (h *TreeSitterHighlighter) Parse(code *string) {
	tree, err := h.parser.ParseCtx(context.Background(), h.tree, []byte(*code))
	if err != nil { fmt.Println(err) }
	h.tree = tree
}

func (h *TreeSitterHighlighter) ReParse(code *string) {
	tree, err := h.parser.ParseCtx(context.Background(), nil, []byte(*code))
	if err != nil { fmt.Println(err) }
	h.tree = tree
}

func (h *TreeSitterHighlighter) ReParseBytes(codeBytes []byte) {
	tree, err := h.parser.ParseCtx(context.Background(), nil, codeBytes)
	if err != nil { fmt.Println(err) }
	h.tree = tree
}


func (h *TreeSitterHighlighter) ColorRanges(from, to int, codeBytes []byte) []ColoredByteRange {

	queryCursor := sitter.NewQueryCursor()
	queryCursor.Exec(h.query, h.tree.RootNode())
	queryCursor.SetPointRange(
		sitter.Point{Row: uint32(from), Column: 0},
		sitter.Point{Row: uint32(to), Column: 0},
	)

	colors := make([]ColoredByteRange, 0)

	for {
		m, ok := queryCursor.NextMatch()
		if !ok { break }
		for _, c := range m.Captures {
			name := h.query.CaptureNameForId(c.Index)
			split := strings.Split(name, ".")
			color := h.matchExpression(split[0], name)

			if !strings.Contains(name, "injection") {
				colors = append(colors, ColoredByteRange{
					StartByte: int(c.Node.StartByte()),
					EndByte:   int(c.Node.EndByte()),
					Color:     color,
				})
			} else {
				injLang := split[len(split)-1]

				if h.injectionLangs == nil { h.injectionLangs = make(map[string]*TreeSitterHighlighter) }
				injectionHighlighter, injLangFound := h.injectionLangs[injLang]
				if !injLangFound {
					injectionHighlighter = NewTreeSitter()
					injectionHighlighter.SetLang(injLang)
					injectionHighlighter.SetTheme(h.themePath)
					h.injectionLangs[injLang] = injectionHighlighter
				}
				contentInjection := codeBytes[c.Node.StartByte():c.Node.EndByte()]

				injectionHighlighter.ReParseBytes(contentInjection)
				countNewlines := CountNewlines(contentInjection)
				colorsInjection := injectionHighlighter.ColorRanges(0, countNewlines, contentInjection)

				startByte := int(c.Node.StartByte())
				for _, colorsInj := range colorsInjection {
					colorsInj.StartByte += startByte
					colorsInj.EndByte += startByte
					colors = append(colors, colorsInj)
				}
			}
		}
	}
	return colors
}


type ColoredByteRange struct {
	StartByte int
	EndByte   int
	Color     int
}

func (h *TreeSitterHighlighter) GetTree() *sitter.Tree {
	return h.tree
}
func (h *TreeSitterHighlighter) GetLang() *sitter.Language {
	return h.language
}
func (h *TreeSitterHighlighter) GetLangStr() string {
	return h.lang
}


type NodeRange struct {
	Ssy int
	Ssx int
	Sey int
	Sex int
}



type Path struct {
	Atx int
	Aty int
	Nodes []NodeRange
	Current int
}

func (p *Path) CurrentNode() NodeRange {
	return p.Nodes[p.Current]
}
func (p *Path) Next() NodeRange {
	p.Current += 1
	if p.Current >= len(p.Nodes) { p.Current = len(p.Nodes) - 1 }
	return p.Nodes[p.Current]
}
func (p *Path) Prev() NodeRange {
	p.Current -= 1
	if p.Current < 0 {
		p.Current = 0;
		return NodeRange{p.Aty,p.Atx,p.Aty,p.Atx}
	}
	return p.Nodes[p.Current]
}

func (h *TreeSitterHighlighter) GetNodePathAt(StartPointRow int, StartPointColumn int,
	EndPointRow int, EndPointColumn int) Path {

	rootNode := h.tree.RootNode()
	node := rootNode.NamedDescendantForPointRange(
		sitter.Point{Row: uint32(StartPointRow), Column: uint32(StartPointColumn)},
		sitter.Point{Row: uint32(EndPointRow), Column: uint32(EndPointColumn)},
	)

	path := Path{Aty: StartPointRow, Atx: StartPointColumn}

	for node != nil {
		r := NodeRange{int(node.StartPoint().Row),
			int(node.StartPoint().Column),
			int(node.EndPoint().Row),
			int(node.EndPoint().Column),
		}
		path.Nodes = append(path.Nodes, r)
		node = node.Parent()
	}

	return path
}

func (h *TreeSitterHighlighter) GetNodeAt(StartPointRow int, StartPointColumn int,
	EndPointRow int, EndPointColumn int) (string, NodeRange) {
	rootNode := h.tree.RootNode()
	node := rootNode.NamedDescendantForPointRange(
		sitter.Point{Row: uint32(StartPointRow), Column: uint32(StartPointColumn)},
		sitter.Point{Row: uint32(EndPointRow), Column: uint32(EndPointColumn)},
	)

	return node.Type(), NodeRange{int(node.StartPoint().Row),
		int(node.StartPoint().Column),
		int(node.EndPoint().Row),
		int(node.EndPoint().Column),
	}
}