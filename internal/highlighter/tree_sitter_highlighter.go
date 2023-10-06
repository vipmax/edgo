package highlighter

import (
	"context"
	. "edgo/internal/langs"
	. "edgo/internal/logger"
	"fmt"
	"github.com/gdamore/tcell"
	sitter "github.com/smacker/go-tree-sitter"
	"github.com/smacker/go-tree-sitter/bash"
	"github.com/smacker/go-tree-sitter/c"
	"github.com/smacker/go-tree-sitter/cpp"
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
	"strconv"
	"strings"
	"time"
	"unicode/utf8"
)

type TreeSitterHighlighter struct {
	parser    *sitter.Parser
	tree      *sitter.Tree
	Colors    [][]int
	lines 	  []string
	lang      string
	language  *sitter.Language
	query     *sitter.Query
	colorsMap map[string]string
}

func NewTreeSitter() *TreeSitterHighlighter {
	parser := sitter.NewParser()

	return &TreeSitterHighlighter{
		parser: parser,
		tree:   nil,
		Colors: [][]int{},
	}
}

type TreeNode struct {
	Fullname         string
	Shortname        string
	Content          string
	StartByte        uint32
	EndByte          uint32
	StartPointRow    uint32
	StartPointColumn uint32
	EndPointRow      uint32
	EndPointColumn   uint32
	Childs           []TreeNode
}

func (tn TreeNode) String() string {
	return tn.Fullname
}



func treeNode(node *sitter.Node, content string) TreeNode {
	return TreeNode{
		Fullname:         node.Type(),
		Shortname:        node.Type(),
		Content: 		  content,
		StartByte:        node.StartByte(),
		EndByte:          node.EndByte(),
		StartPointRow:    node.StartPoint().Row,
		StartPointColumn: node.StartPoint().Column,
		EndPointRow:      node.EndPoint().Row,
		EndPointColumn:   node.EndPoint().Column,
	}
}
func populateTreeNode(node *sitter.Node, codeBytes []byte) TreeNode {
	tree := treeNode(node, node.Content(codeBytes))

	childCount := int(node.ChildCount())
	for i := 0; i < childCount; i++ {
		child := node.Child(i)
		//if child.Type() == "\n" { continue }
		childTree := populateTreeNode(child, codeBytes)
		tree.Childs = append(tree.Childs, childTree)
	}

	return tree
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


	err = Unmarshal(yamlFile, &h.colorsMap)
	if err != nil {
		log.Fatalf("Error unmarshaling YAML: %v", err)
	}

	if value, ok := h.colorsMap["accent_color"]; ok {
		AccentColor = h.ParseColor(value)
	}
	if value, ok := h.colorsMap["accent_color2"]; ok {
		AccentColor2 =  h.ParseColor(value)
	}
	if value, ok := h.colorsMap["accent_color3"]; ok {
		AccentColor3 =  h.ParseColor(value)
	}

	//fmt.Println("Cases and Return Values:")
	//for caseName, returnValue := range h.colorsMap {
	//	fmt.Printf("%s: %d\n", caseName, returnValue)
	//}
}

func GetSitterLang(lang string) *sitter.Language {
	switch lang {
	case "javascript": return javascript.GetLanguage()
	case "typescript": return typescript.GetLanguage()
	case "go": return golang.GetLanguage()
	case "python": return python.GetLanguage()
	case "html": return html.GetLanguage()
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
	h.lang = lang
	h.language = GetSitterLang(lang)
	h.parser.SetLanguage(h.language)
}


func (h *TreeSitterHighlighter) matchExpression(expression string, fullexpression string) int {
	if value, ok := h.colorsMap[fullexpression]; ok { return  h.ParseColor(value) }
	if value, ok := h.colorsMap[expression]; ok { return  h.ParseColor(value) }
	return -1
}


func (h *TreeSitterHighlighter) Colorize(newCode string) [][]int {
	code := []byte(newCode)

	start := time.Now()
	tree, err := h.parser.ParseCtx(context.Background(), nil, code)
	if err != nil { fmt.Println(err) }
	h.tree = tree
	Log.Info("[Colorize] tree parsed, elapsed", time.Since(start).String())

	//treeForDebug := populateTreeNode(h.tree.RootNode(), code); Use(treeForDebug)

	h.lines = strings.Split(newCode, "\n")
	h.Colors = make([][]int, len(h.lines))

	for i, line := range h.lines {
		ints := make([]int, len(line))
		for j := range ints { ints[j] = -1 }
		h.Colors[i] = ints
	}

	// Execute the query to highlight keywords
	if h.query == nil {
		// create query only once
		startquery := time.Now()
		queryLang := MatchQueryLang(h.lang)
		q, _ := sitter.NewQuery([]byte(queryLang), h.language)
		h.query = q
		Log.Info("tree-sitter NewQuery, elapsed: " + time.Since(startquery).String())
	}

	h.ColorizeRange(newCode,
		int(h.tree.RootNode().StartPoint().Row), int(h.tree.RootNode().StartPoint().Column),
		int(h.tree.RootNode().EndPoint().Row), int(h.tree.RootNode().EndPoint().Column),
	)

	Log.Info("tree-sitter full colorize, elapsed: " + time.Since(start).String())

	return h.Colors
}


/*
	comment for sitter.EditInput
	The StartIndex, OldEndIndex, and NewEndIndex parameters indicate the range of bytes you're modifying
	The StartPoint, OldEndPoint, and NewEndPoint parameters indicate the range of positions (line, column) affected by the edit.
*/

func (h *TreeSitterHighlighter) EnterEdit(code string, row int, col int) {
	StartIndex := GetStartIndex(code, row, col)
	Row := uint32(row); Column := uint32(col)

	editInput := sitter.EditInput{
		StartIndex: StartIndex, OldEndIndex: StartIndex, NewEndIndex: StartIndex + 1,
		StartPoint:  sitter.Point{Row: Row, Column: Column},
		OldEndPoint: sitter.Point{Row: Row, Column: Column},
		NewEndPoint: sitter.Point{Row: Row + 1, Column: 0},
	}
	h.tree.Edit(editInput)
}

func (h *TreeSitterHighlighter) RemoveLineEdit(code string, row int, col int) {
	StartIndex := GetStartIndex(code, row, col)
	Row := uint32(row); Column := uint32(col)

	editInput := sitter.EditInput{
		StartIndex: StartIndex, OldEndIndex: StartIndex+1, NewEndIndex: StartIndex,
		StartPoint:  sitter.Point{Row: Row, Column: Column},
		OldEndPoint: sitter.Point{Row: Row + 1, Column: 0},
		NewEndPoint: sitter.Point{Row: Row, Column: Column},
	}
	h.tree.Edit(editInput)
}

func (h *TreeSitterHighlighter) AddCharEdit(code string, row int, col int) {
	StartIndex := GetStartIndex(code, row, col)
	Row := uint32(row); Column := uint32(col)

	editInput := sitter.EditInput{
		StartIndex: StartIndex, OldEndIndex: StartIndex,
		NewEndIndex: StartIndex + 1,
		StartPoint:  sitter.Point{Row: Row, Column: Column},
		OldEndPoint: sitter.Point{Row: Row, Column: Column},
		NewEndPoint: sitter.Point{Row: Row, Column: Column + 1},
	}
	h.tree.Edit(editInput)
}

func (h *TreeSitterHighlighter) AddMultipleCharEdit(code string, startrow int, startcol int, endrow int, endcol int) {
	StartIndex := GetStartIndex(code, startrow, startcol)
	EndIndex := GetStartIndex(code, endrow, endcol)

	editInput := sitter.EditInput{
		StartIndex: StartIndex, OldEndIndex: StartIndex, NewEndIndex: EndIndex,
		StartPoint:  sitter.Point{Row: uint32(startrow), Column: uint32(startcol)},
		OldEndPoint: sitter.Point{Row: uint32(startrow), Column: uint32(startcol)},
		NewEndPoint: sitter.Point{Row: uint32(endrow), Column: uint32(endcol)},
	}
	h.tree.Edit(editInput)
}

func (h *TreeSitterHighlighter) RemoveCharEdit(code string, row int, col int) {
	StartIndex := GetStartIndex(code, row, col)
	Row := uint32(row); Column := uint32(col)

	editInput := sitter.EditInput{
		StartIndex: StartIndex, OldEndIndex: StartIndex + 1, NewEndIndex: StartIndex,
		StartPoint:  sitter.Point{Row: Row, Column: Column},
		OldEndPoint: sitter.Point{Row: Row, Column: Column +1},
		NewEndPoint: sitter.Point{Row: Row, Column: Column},
	}
	h.tree.Edit(editInput)
}


func GetStartIndex(code string, row int, col int) uint32 {
	r, c, startIndex := 0, 0, 0
	for _, char := range code {
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


func (h *TreeSitterHighlighter) colorizeChildNodes(node *sitter.Node, code []byte) {
	//tn := populateTreeNode(node, code) // for debug print, delete it later
	//Use(tn)

	color := h.matchExpression(node.Type(), node.Type())
	//content := node.Content(code)
	content := code[node.StartByte():node.EndByte()]
	if color > 0 {
		h.colorizeNode(node, content, color)
	}

	childCount := int(node.NamedChildCount())
	for i := 0; i < childCount; i++ {
		child := node.NamedChild(i)
		h.colorizeChildNodes(child, code)
	}
}

func (h *TreeSitterHighlighter) colorizeNode(node *sitter.Node, nodeContent []byte, color int) {
	//tn := treeNode(node, string(nodeContent)); Use(tn)
	s := string(nodeContent)
	i := node.StartPoint().Row
	j := int(node.StartPoint().Column)  // todo; node.StartPoint().Column is in bytes, needs to recalculate it to position
	column := j

	jj := 0; column = 0
	for _, char := range h.lines[i] {
		if jj >= j { break }
		runeLen := utf8.RuneLen(char)
		jj += runeLen
		column += 1

	}

	for _, character := range s {
		if character == '\n' { i++; column = 0; continue }
		if column >= len(h.Colors[i]) {
			h.Colors[i] = append(h.Colors[i], color)
		} else  {
			h.Colors[i][column] = color
		}

		column++
	}
}

func Use(vals ...interface{}) { }


func (h *TreeSitterHighlighter) ColorizeRange(newcode string,
	StartPointRow, StartPointColumn, EndPointRow, EndPointColumn int) {

	code := []byte(newcode)
	starttime := time.Now()
	tree, err := h.parser.ParseCtx(context.Background(), h.tree, code)
	if err != nil { fmt.Println(err) }

	h.tree = tree
	Log.Info("tree-sitter edit, elapsed: " + time.Since(starttime).String())
	h.lines = strings.Split(newcode, "\n")

	//treeDebug := populateTreeNode(h.tree.RootNode(), code); Use(treeDebug)

	rootNode := h.tree.RootNode()
	node := rootNode.NamedDescendantForPointRange(
		sitter.Point{Row: uint32(StartPointRow), Column: uint32(StartPointColumn)},
		sitter.Point{Row: uint32(EndPointRow), Column: uint32(EndPointColumn)},
	)

	nodeType := node.Type()
	nodename := strings.Split(nodeType, ".")[0]

	content := code[node.StartByte():node.EndByte()]
	//tn := treeNode(node, content)
	//Use(tn)

	Log.Info(fmt.Sprintf("tree-sitter edit node {%d %d} {%d %d} type=%s content=%s ",
		node.StartPoint().Row, node.StartPoint().Column, node.EndPoint().Row, node.EndPoint().Column,
		nodeType, strconv.Itoa(len(content))))

	if nodeType == "ERROR" {
		//color := h.matchExpression(nodename, nodeType)
		//h.colorizeNode(node, content,9)
		return
	}

	//h.colorizeedNode(node, content,-1) // reset colors

	color := h.matchExpression(nodename, nodeType)
	if color > 0 {
		h.colorizeNode(node, content, color)
	} else if node.ChildCount() > 0 {
		s := time.Now()
		h.colorizeChildNodes(node, code)
		Log.Info("tree-sitter colorizeChildNodes, elapsed: " + time.Since(s).String())
	}

	// highlight with query
	h.colorizeWithQuery(node, code)

	Log.Info("tree-sitter ColorizeRange, elapsed: " + time.Since(starttime).String())
}

func (h *TreeSitterHighlighter) colorizeWithQuery(node *sitter.Node, code []byte) {
	starttime := time.Now()

	qc := sitter.NewQueryCursor()
	qc.Exec(h.query, node)

	for {
		m, ok := qc.NextMatch()
		if !ok { break }
		for _, c := range m.Captures {
			name := h.query.CaptureNameForId(c.Index)
			nodename := strings.Split(name, ".")[0]
			color := h.matchExpression(nodename, name)

			//tn := treeNode(c.Node, content); Use(tn)
			if color > 0 {
				content := code[c.Node.StartByte():c.Node.EndByte()]
				h.colorizeNode(c.Node, content, color)
			}
		}
	}

	Log.Info("tree-sitter colorizeWithQuery, elapsed: " + time.Since(starttime).String())
}

func (h *TreeSitterHighlighter) GetTree() *sitter.Tree {
	return h.tree
}
func (h *TreeSitterHighlighter) GetLang() *sitter.Language {
	return h.language
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
