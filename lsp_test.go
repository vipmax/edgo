package main

import (
	"encoding/json"
	"fmt"
	"os"
	"testing"
	"time"
)

func TestCompletion(t *testing.T) {
	dir, _ := os.Getwd()
	file := dir + "/lsp_test.go"
	text, _ := readFileToString(file)
	
	fmt.Println("starting lsp server")

	lsp := LspClient{}
	lsp.start()
	lsp.init(dir)
	lsp.didOpen(file)
	time.Sleep(time.Millisecond * 100)

	completion, _ := lsp.completion(file, text, 18-1, 8)
	fmt.Println("completion", completion)

	var options []string
	items := completion.Result.Items
	for _, item := range items {
		options = append(options, item.Label)
	}

	fmt.Println("options", options)
	fmt.Println("ending lsp server")
}

func TestParse(t *testing.T) {

	// Example JSON data
	jsonData := `{"jsonrpc":"2.0","result":{"isIncomplete":true,"items":[{"label":"dir","kind":6,"detail":"string","preselect":true,"sortText":"00000","filterText":"dir","insertTextFormat":1,"textEdit":{"range":{"start":{"line":17,"character":8},"end":{"line":17,"character":8}},"newText":"dir"}},{"label":"file","kind":6,"detail":"string","sortText":"00001","filterText":"file","insertTextFormat":1,"textEdit":{"range":{"start":{"line":17,"character":8},"end":{"line":17,"character":8}},"newText":"file"}},{"label":"t","kind":6,"detail":"*testing.T","sortText":"00002","filterText":"t","insertTextFormat":1,"textEdit":{"range":{"start":{"line":17,"character":8},"end":{"line":17,"character":8}},"newText":"t"}},{"label":"text","kind":6,"detail":"string","sortText":"00003","filterText":"text","insertTextFormat":1,"textEdit":{"range":{"start":{"line":17,"character":8},"end":{"line":17,"character":8}},"newText":"text"}},{"label":"fmt","kind":9,"detail":"\"fmt\"","sortText":"00004","filterText":"fmt","insertTextFormat":1,"textEdit":{"range":{"start":{"line":17,"character":8},"end":{"line":17,"character":8}},"newText":"fmt"}},{"label":"json","kind":9,"detail":"\"encoding/json\"","sortText":"00005","filterText":"json","insertTextFormat":1,"textEdit":{"range":{"start":{"line":17,"character":8},"end":{"line":17,"character":8}},"newText":"json"}},{"label":"os","kind":9,"detail":"\"os\"","sortText":"00006","filterText":"os","insertTextFormat":1,"textEdit":{"range":{"start":{"line":17,"character":8},"end":{"line":17,"character":8}},"newText":"os"}},{"label":"testing","kind":9,"detail":"\"testing\"","sortText":"00007","filterText":"testing","insertTextFormat":1,"textEdit":{"range":{"start":{"line":17,"character":8},"end":{"line":17,"character":8}},"newText":"testing"}},{"label":"time","kind":9,"detail":"\"time\"","sortText":"00008","filterText":"time","insertTextFormat":1,"textEdit":{"range":{"start":{"line":17,"character":8},"end":{"line":17,"character":8}},"newText":"time"}},{"label":"c","kind":6,"detail":"int","documentation":"cursor position, row and column\n","sortText":"00009","filterText":"c","insertTextFormat":1,"textEdit":{"range":{"start":{"line":17,"character":8},"end":{"line":17,"character":8}},"newText":"c"}},{"label":"colors","kind":6,"detail":"[][]int","documentation":"characters colors\n","sortText":"00010","filterText":"colors","insertTextFormat":1,"textEdit":{"range":{"start":{"line":17,"character":8},"end":{"line":17,"character":8}},"newText":"colors"}},{"label":"contains","kind":3,"detail":"func(slice []T, e T) bool","sortText":"00011","filterText":"contains","insertTextFormat":1,"textEdit":{"range":{"start":{"line":17,"character":8},"end":{"line":17,"character":8}},"newText":"contains"}},{"label":"content","kind":6,"detail":"[][]rune","documentation":"characters\n","sortText":"00012","filterText":"content","insertTextFormat":1,"textEdit":{"range":{"start":{"line":17,"character":8},"end":{"line":17,"character":8}},"newText":"content"}},{"label":"convertToString","kind":3,"detail":"func(replaceSpaces bool) string","sortText":"00013","filterText":"convertToString","insertTextFormat":1,"textEdit":{"range":{"start":{"line":17,"character":8},"end":{"line":17,"character":8}},"newText":"convertToString"}},{"label":"countConsecutiveSpaces","kind":3,"detail":"func(runes []rune, before int) int","sortText":"00014","filterText":"countConsecutiveSpaces","insertTextFormat":1,"textEdit":{"range":{"start":{"line":17,"character":8},"end":{"line":17,"character":8}},"newText":"countConsecutiveSpaces"}},{"label":"countTabsFromString","kind":3,"detail":"func(str string, stopIndex int) int","sortText":"00015","filterText":"countTabsFromString","insertTextFormat":1,"textEdit":{"range":{"start":{"line":17,"character":8},"end":{"line":17,"character":8}},"newText":"countTabsFromString"}},{"label":"countTabsInRow","kind":3,"detail":"func(i int) int","sortText":"00016","filterText":"countTabsInRow","insertTextFormat":1,"textEdit":{"range":{"start":{"line":17,"character":8},"end":{"line":17,"character":8}},"newText":"countTabsInRow"}},{"label":"countTabsInRowBefore","kind":3,"detail":"func(i int, before int) int","sortText":"00017","filterText":"countTabsInRowBefore","insertTextFormat":1,"textEdit":{"range":{"start":{"line":17,"character":8},"end":{"line":17,"character":8}},"newText":"countTabsInRowBefore"}},{"label":"filename","kind":6,"detail":"string","documentation":"file name to show\n","sortText":"00018","filterText":"filename","insertTextFormat":1,"textEdit":{"range":{"start":{"line":17,"character":8},"end":{"line":17,"character":8}},"newText":"filename"}},{"label":"findNextWord","kind":3,"detail":"func(chars []rune, from int) int","sortText":"00019","filterText":"findNextWord","insertTextFormat":1,"textEdit":{"range":{"start":{"line":17,"character":8},"end":{"line":17,"character":8}},"newText":"findNextWord"}},{"label":"findPrevWord","kind":3,"detail":"func(chars []rune, from int) int","sortText":"00020","filterText":"findPrevWord","insertTextFormat":1,"textEdit":{"range":{"start":{"line":17,"character":8},"end":{"line":17,"character":8}},"newText":"findPrevWord"}},{"label":"getSelection","kind":3,"detail":"func() string","sortText":"00021","filterText":"getSelection","insertTextFormat":1,"textEdit":{"range":{"start":{"line":17,"character":8},"end":{"line":17,"character":8}},"newText":"getSelection"}},{"label":"highlighter","kind":6,"detail":"Highlighter","sortText":"00022","filterText":"highlighter","insertTextFormat":1,"textEdit":{"range":{"start":{"line":17,"character":8},"end":{"line":17,"character":8}},"newText":"highlighter"}},{"label":"insert","kind":3,"detail":"func(a []T, index int, value T) []T","sortText":"00023","filterText":"insert","insertTextFormat":1,"textEdit":{"range":{"start":{"line":17,"character":8},"end":{"line":17,"character":8}},"newText":"insert"}},{"label":"isUnderSelection","kind":3,"detail":"func(x int, y int) bool","sortText":"00024","filterText":"isUnderSelection","insertTextFormat":1,"textEdit":{"range":{"start":{"line":17,"character":8},"end":{"line":17,"character":8}},"newText":"isUnderSelection"}},{"label":"lsp","kind":6,"detail":"LspClient","sortText":"00025","filterText":"lsp","insertTextFormat":1,"textEdit":{"range":{"start":{"line":17,"character":8},"end":{"line":17,"character":8}},"newText":"lsp"}},{"label":"matched","kind":6,"detail":"[]rune","sortText":"00026","filterText":"matched","insertTextFormat":1,"textEdit":{"range":{"start":{"line":17,"character":8},"end":{"line":17,"character":8}},"newText":"matched"}},{"label":"max","kind":3,"detail":"func(x int, y int) int","sortText":"00027","filterText":"max","insertTextFormat":1,"textEdit":{"range":{"start":{"line":17,"character":8},"end":{"line":17,"character":8}},"newText":"max"}},{"label":"maxMany","kind":3,"detail":"func(nums ...int) int","sortText":"00028","filterText":"maxMany","insertTextFormat":1,"textEdit":{"range":{"start":{"line":17,"character":8},"end":{"line":17,"character":8}},"newText":"maxMany"}},{"label":"maxString","kind":3,"detail":"func(arr []string) int","sortText":"00029","filterText":"maxString","insertTextFormat":1,"textEdit":{"range":{"start":{"line":17,"character":8},"end":{"line":17,"character":8}},"newText":"maxString"}},{"label":"min","kind":3,"detail":"func(x int, y int) int","sortText":"00030","filterText":"min","insertTextFormat":1,"textEdit":{"range":{"start":{"line":17,"character":8},"end":{"line":17,"character":8}},"newText":"min"}},{"label":"minMany","kind":3,"detail":"func(nums ...int) int","sortText":"00031","filterText":"minMany","insertTextFormat":1,"textEdit":{"range":{"start":{"line":17,"character":8},"end":{"line":17,"character":8}},"newText":"minMany"}},{"label":"r","kind":6,"detail":"int","documentation":"cursor position, row and column\n","sortText":"00032","filterText":"r","insertTextFormat":1,"textEdit":{"range":{"start":{"line":17,"character":8},"end":{"line":17,"character":8}},"newText":"r"}},{"label":"remove","kind":3,"detail":"func(slice []T, s int) []T","sortText":"00033","filterText":"remove","insertTextFormat":1,"textEdit":{"range":{"start":{"line":17,"character":8},"end":{"line":17,"character":8}},"newText":"remove"}},{"label":"ssx","kind":6,"detail":"int","documentation":"left shift for line number\n","sortText":"00034","filterText":"ssx","insertTextFormat":1,"textEdit":{"range":{"start":{"line":17,"character":8},"end":{"line":17,"character":8}},"newText":"ssx"}},{"label":"ssy","kind":6,"detail":"int","documentation":"left shift for line number\n","sortText":"00035","filterText":"ssy","insertTextFormat":1,"textEdit":{"range":{"start":{"line":17,"character":8},"end":{"line":17,"character":8}},"newText":"ssy"}},{"label":"x","kind":6,"detail":"int","documentation":"offset for scrolling for row and column\n","sortText":"00036","filterText":"x","insertTextFormat":1,"textEdit":{"range":{"start":{"line":17,"character":8},"end":{"line":17,"character":8}},"newText":"x"}},{"label":"y","kind":6,"detail":"int","documentation":"offset for scrolling for row and column\n","sortText":"00037","filterText":"y","insertTextFormat":1,"textEdit":{"range":{"start":{"line":17,"character":8},"end":{"line":17,"character":8}},"newText":"y"}},{"label":"lsp.process","kind":5,"detail":"*exec.Cmd","documentation":"The underlying process running the LSP server.\n","sortText":"00038","filterText":"lsp.process","insertTextFormat":1,"textEdit":{"range":{"start":{"line":17,"character":8},"end":{"line":17,"character":8}},"newText":"lsp.process"}},{"label":"lsp.stdin","kind":5,"detail":"io.WriteCloser","documentation":"The standard input pipe for sending data to the LSP server.\n","sortText":"00039","filterText":"lsp.stdin","insertTextFormat":1,"textEdit":{"range":{"start":{"line":17,"character":8},"end":{"line":17,"character":8}},"newText":"lsp.stdin"}},{"label":"lsp.stdout","kind":5,"detail":"io.ReadCloser","sortText":"00040","filterText":"lsp.stdout","insertTextFormat":1,"textEdit":{"range":{"start":{"line":17,"character":8},"end":{"line":17,"character":8}},"newText":"lsp.stdout"}},{"label":"BaseRequest","kind":22,"detail":"struct{...}","sortText":"00041","filterText":"BaseRequest","insertTextFormat":1,"textEdit":{"range":{"start":{"line":17,"character":8},"end":{"line":17,"character":8}},"newText":"BaseRequest"}},{"label":"COLUMNS","kind":6,"detail":"int","documentation":"term size\n","sortText":"00042","filterText":"COLUMNS","insertTextFormat":1,"textEdit":{"range":{"start":{"line":17,"character":8},"end":{"line":17,"character":8}},"newText":"COLUMNS"}},{"label":"ClientInfo","kind":22,"detail":"struct{...}","sortText":"00043","filterText":"ClientInfo","insertTextFormat":1,"textEdit":{"range":{"start":{"line":17,"character":8},"end":{"line":17,"character":8}},"newText":"ClientInfo"}},{"label":"CompletionItem","kind":22,"detail":"struct{...}","sortText":"00044","filterText":"CompletionItem","insertTextFormat":1,"textEdit":{"range":{"start":{"line":17,"character":8},"end":{"line":17,"character":8}},"newText":"CompletionItem"}},{"label":"CompletionResponse","kind":22,"detail":"struct{...}","sortText":"00045","filterText":"CompletionResponse","insertTextFormat":1,"textEdit":{"range":{"start":{"line":17,"character":8},"end":{"line":17,"character":8}},"newText":"CompletionResponse"}},{"label":"CompletionResult","kind":22,"detail":"struct{...}","sortText":"00046","filterText":"CompletionResult","insertTextFormat":1,"textEdit":{"range":{"start":{"line":17,"character":8},"end":{"line":17,"character":8}},"newText":"CompletionResult"}},{"label":"Context","kind":22,"detail":"struct{...}","sortText":"00047","filterText":"Context","insertTextFormat":1,"textEdit":{"range":{"start":{"line":17,"character":8},"end":{"line":17,"character":8}},"newText":"Context"}},{"label":"DidOpenParams","kind":22,"detail":"struct{...}","sortText":"00048","filterText":"DidOpenParams","insertTextFormat":1,"textEdit":{"range":{"start":{"line":17,"character":8},"end":{"line":17,"character":8}},"newText":"DidOpenParams"}},{"label":"DidOpenRequest","kind":22,"detail":"struct{...}","sortText":"00049","filterText":"DidOpenRequest","insertTextFormat":1,"textEdit":{"range":{"start":{"line":17,"character":8},"end":{"line":17,"character":8}},"newText":"DidOpenRequest"}},{"label":"Editor","kind":22,"detail":"struct{...}","sortText":"00050","filterText":"Editor","insertTextFormat":1,"textEdit":{"range":{"start":{"line":17,"character":8},"end":{"line":17,"character":8}},"newText":"Editor"}},{"label":"GreaterEqual","kind":3,"detail":"func(x int, y int, x1 int, y1 int) bool","sortText":"00051","filterText":"GreaterEqual","insertTextFormat":1,"textEdit":{"range":{"start":{"line":17,"character":8},"end":{"line":17,"character":8}},"newText":"GreaterEqual"}},{"label":"GreaterThan","kind":3,"detail":"func(x int, y int, x1 int, y1 int) bool","sortText":"00052","filterText":"GreaterThan","insertTextFormat":1,"textEdit":{"range":{"start":{"line":17,"character":8},"end":{"line":17,"character":8}},"newText":"GreaterThan"}},{"label":"Highlighter","kind":22,"detail":"struct{...}","sortText":"00053","filterText":"Highlighter","insertTextFormat":1,"textEdit":{"range":{"start":{"line":17,"character":8},"end":{"line":17,"character":8}},"newText":"Highlighter"}},{"label":"InitializeParams","kind":22,"detail":"struct{...}","sortText":"00054","filterText":"InitializeParams","insertTextFormat":1,"textEdit":{"range":{"start":{"line":17,"character":8},"end":{"line":17,"character":8}},"newText":"InitializeParams"}},{"label":"InitializeRequest","kind":22,"detail":"struct{...}","sortText":"00055","filterText":"InitializeRequest","insertTextFormat":1,"textEdit":{"range":{"start":{"line":17,"character":8},"end":{"line":17,"character":8}},"newText":"InitializeRequest"}},{"label":"InitializedRequest","kind":22,"detail":"struct{...}","sortText":"00056","filterText":"InitializedRequest","insertTextFormat":1,"textEdit":{"range":{"start":{"line":17,"character":8},"end":{"line":17,"character":8}},"newText":"InitializedRequest"}},{"label":"LS","kind":6,"detail":"int","documentation":"left shift for line number\n","sortText":"00057","filterText":"LS","insertTextFormat":1,"textEdit":{"range":{"start":{"line":17,"character":8},"end":{"line":17,"character":8}},"newText":"LS"}},{"label":"LessThan","kind":3,"detail":"func(x int, y int, x1 int, y1 int) bool","sortText":"00058","filterText":"LessThan","insertTextFormat":1,"textEdit":{"range":{"start":{"line":17,"character":8},"end":{"line":17,"character":8}},"newText":"LessThan"}},{"label":"LspClient","kind":22,"detail":"struct{...}","documentation":"LspClient represents a client for communicating with a Language Server Protocol (LSP) server.\n","sortText":"00059","filterText":"LspClient","insertTextFormat":1,"textEdit":{"range":{"start":{"line":17,"character":8},"end":{"line":17,"character":8}},"newText":"LspClient"}},{"label":"Params","kind":22,"detail":"struct{...}","sortText":"00060","filterText":"Params","insertTextFormat":1,"textEdit":{"range":{"start":{"line":17,"character":8},"end":{"line":17,"character":8}},"newText":"Params"}},{"label":"Position","kind":22,"detail":"struct{...}","sortText":"00061","filterText":"Position","insertTextFormat":1,"textEdit":{"range":{"start":{"line":17,"character":8},"end":{"line":17,"character":8}},"newText":"Position"}},{"label":"PositionResponse","kind":22,"detail":"struct{...}","sortText":"00062","filterText":"PositionResponse","insertTextFormat":1,"textEdit":{"range":{"start":{"line":17,"character":8},"end":{"line":17,"character":8}},"newText":"PositionResponse"}},{"label":"ROWS","kind":6,"detail":"int","documentation":"term size\n","sortText":"00063","filterText":"ROWS","insertTextFormat":1,"textEdit":{"range":{"start":{"line":17,"character":8},"end":{"line":17,"character":8}},"newText":"ROWS"}},{"label":"Range","kind":22,"detail":"struct{...}","sortText":"00064","filterText":"Range","insertTextFormat":1,"textEdit":{"range":{"start":{"line":17,"character":8},"end":{"line":17,"character":8}},"newText":"Range"}},{"label":"TextDocument","kind":22,"detail":"struct{...}","sortText":"00065","filterText":"TextDocument","insertTextFormat":1,"textEdit":{"range":{"start":{"line":17,"character":8},"end":{"line":17,"character":8}},"newText":"TextDocument"}},{"label":"TextDocumentContentChangeEvent","kind":22,"detail":"struct{...}","sortText":"00066","filterText":"TextDocumentContentChangeEvent","insertTextFormat":1,"textEdit":{"range":{"start":{"line":17,"character":8},"end":{"line":17,"character":8}},"newText":"TextDocumentContentChangeEvent"}},{"label":"TextDocumentDidChangeParams","kind":22,"detail":"struct{...}","sortText":"00067","filterText":"TextDocumentDidChangeParams","insertTextFormat":1,"textEdit":{"range":{"start":{"line":17,"character":8},"end":{"line":17,"character":8}},"newText":"TextDocumentDidChangeParams"}},{"label":"TextEdit","kind":22,"detail":"struct{...}","sortText":"00068","filterText":"TextEdit","insertTextFormat":1,"textEdit":{"range":{"start":{"line":17,"character":8},"end":{"line":17,"character":8}},"newText":"TextEdit"}},{"label":"VersionedTextDocumentIdentifier","kind":22,"detail":"struct{...}","sortText":"00069","filterText":"VersionedTextDocumentIdentifier","insertTextFormat":1,"textEdit":{"range":{"start":{"line":17,"character":8},"end":{"line":17,"character":8}},"newText":"VersionedTextDocumentIdentifier"}},{"label":"WorkspaceFolder","kind":22,"detail":"struct{...}","sortText":"00070","filterText":"WorkspaceFolder","insertTextFormat":1,"textEdit":{"range":{"start":{"line":17,"character":8},"end":{"line":17,"character":8}},"newText":"WorkspaceFolder"}},{"label":"append","kind":3,"detail":"func(slice []Type, elems ...Type) []Type","sortText":"00071","filterText":"append","insertTextFormat":1,"textEdit":{"range":{"start":{"line":17,"character":8},"end":{"line":17,"character":8}},"newText":"append"}},{"label":"bool","kind":7,"sortText":"00072","filterText":"bool","insertTextFormat":1,"textEdit":{"range":{"start":{"line":17,"character":8},"end":{"line":17,"character":8}},"newText":"bool"}},{"label":"byte","kind":7,"sortText":"00073","filterText":"byte","insertTextFormat":1,"textEdit":{"range":{"start":{"line":17,"character":8},"end":{"line":17,"character":8}},"newText":"byte"}},{"label":"cap","kind":3,"detail":"func(v Type) int","sortText":"00074","filterText":"cap","insertTextFormat":1,"textEdit":{"range":{"start":{"line":17,"character":8},"end":{"line":17,"character":8}},"newText":"cap"}},{"label":"close","kind":3,"detail":"func(c chan\u003c- Type)","sortText":"00075","filterText":"close","insertTextFormat":1,"textEdit":{"range":{"start":{"line":17,"character":8},"end":{"line":17,"character":8}},"newText":"close"}},{"label":"complex","kind":3,"detail":"func(r float64, i float64) complex128","sortText":"00076","filterText":"complex","insertTextFormat":1,"textEdit":{"range":{"start":{"line":17,"character":8},"end":{"line":17,"character":8}},"newText":"complex"}},{"label":"complex128","kind":7,"sortText":"00077","filterText":"complex128","insertTextFormat":1,"textEdit":{"range":{"start":{"line":17,"character":8},"end":{"line":17,"character":8}},"newText":"complex128"}},{"label":"complex64","kind":7,"sortText":"00078","filterText":"complex64","insertTextFormat":1,"textEdit":{"range":{"start":{"line":17,"character":8},"end":{"line":17,"character":8}},"newText":"complex64"}},{"label":"copy","kind":3,"detail":"func(dst []Type, src []Type) int","sortText":"00079","filterText":"copy","insertTextFormat":1,"textEdit":{"range":{"start":{"line":17,"character":8},"end":{"line":17,"character":8}},"newText":"copy"}},{"label":"delete","kind":3,"detail":"func(m map[Type]Type1, key Type)","sortText":"00080","filterText":"delete","insertTextFormat":1,"textEdit":{"range":{"start":{"line":17,"character":8},"end":{"line":17,"character":8}},"newText":"delete"}},{"label":"error","kind":8,"sortText":"00081","filterText":"error","insertTextFormat":1,"textEdit":{"range":{"start":{"line":17,"character":8},"end":{"line":17,"character":8}},"newText":"error"}},{"label":"false","kind":21,"sortText":"00082","filterText":"false","insertTextFormat":1,"textEdit":{"range":{"start":{"line":17,"character":8},"end":{"line":17,"character":8}},"newText":"false"}},{"label":"float32","kind":7,"sortText":"00083","filterText":"float32","insertTextFormat":1,"textEdit":{"range":{"start":{"line":17,"character":8},"end":{"line":17,"character":8}},"newText":"float32"}},{"label":"float64","kind":7,"sortText":"00084","filterText":"float64","insertTextFormat":1,"textEdit":{"range":{"start":{"line":17,"character":8},"end":{"line":17,"character":8}},"newText":"float64"}},{"label":"imag","kind":3,"detail":"func(c complex128) float64","sortText":"00085","filterText":"imag","insertTextFormat":1,"textEdit":{"range":{"start":{"line":17,"character":8},"end":{"line":17,"character":8}},"newText":"imag"}},{"label":"int","kind":7,"sortText":"00086","filterText":"int","insertTextFormat":1,"textEdit":{"range":{"start":{"line":17,"character":8},"end":{"line":17,"character":8}},"newText":"int"}},{"label":"int16","kind":7,"sortText":"00087","filterText":"int16","insertTextFormat":1,"textEdit":{"range":{"start":{"line":17,"character":8},"end":{"line":17,"character":8}},"newText":"int16"}},{"label":"int32","kind":7,"sortText":"00088","filterText":"int32","insertTextFormat":1,"textEdit":{"range":{"start":{"line":17,"character":8},"end":{"line":17,"character":8}},"newText":"int32"}},{"label":"int64","kind":7,"sortText":"00089","filterText":"int64","insertTextFormat":1,"textEdit":{"range":{"start":{"line":17,"character":8},"end":{"line":17,"character":8}},"newText":"int64"}},{"label":"int8","kind":7,"sortText":"00090","filterText":"int8","insertTextFormat":1,"textEdit":{"range":{"start":{"line":17,"character":8},"end":{"line":17,"character":8}},"newText":"int8"}},{"label":"len","kind":3,"detail":"func(v Type) int","sortText":"00091","filterText":"len","insertTextFormat":1,"textEdit":{"range":{"start":{"line":17,"character":8},"end":{"line":17,"character":8}},"newText":"len"}},{"label":"make","kind":3,"detail":"func(t Type, size ...int) Type","sortText":"00092","filterText":"make","insertTextFormat":1,"textEdit":{"range":{"start":{"line":17,"character":8},"end":{"line":17,"character":8}},"newText":"make"}},{"label":"new","kind":3,"detail":"func(Type) *Type","sortText":"00093","filterText":"new","insertTextFormat":1,"textEdit":{"range":{"start":{"line":17,"character":8},"end":{"line":17,"character":8}},"newText":"new"}},{"label":"panic","kind":3,"detail":"func(v any)","sortText":"00094","filterText":"panic","insertTextFormat":1,"textEdit":{"range":{"start":{"line":17,"character":8},"end":{"line":17,"character":8}},"newText":"panic"}},{"label":"print","kind":3,"detail":"func(args ...Type)","sortText":"00095","filterText":"print","insertTextFormat":1,"textEdit":{"range":{"start":{"line":17,"character":8},"end":{"line":17,"character":8}},"newText":"print"}},{"label":"println","kind":3,"detail":"func(args ...Type)","sortText":"00096","filterText":"println","insertTextFormat":1,"textEdit":{"range":{"start":{"line":17,"character":8},"end":{"line":17,"character":8}},"newText":"println"}},{"label":"real","kind":3,"detail":"func(c complex128) float64","sortText":"00097","filterText":"real","insertTextFormat":1,"textEdit":{"range":{"start":{"line":17,"character":8},"end":{"line":17,"character":8}},"newText":"real"}},{"label":"recover","kind":3,"detail":"func() any","sortText":"00098","filterText":"recover","insertTextFormat":1,"textEdit":{"range":{"start":{"line":17,"character":8},"end":{"line":17,"character":8}},"newText":"recover"}},{"label":"rune","kind":7,"sortText":"00099","filterText":"rune","insertTextFormat":1,"textEdit":{"range":{"start":{"line":17,"character":8},"end":{"line":17,"character":8}},"newText":"rune"}},{"label":"string","kind":7,"sortText":"00100","filterText":"string","insertTextFormat":1,"textEdit":{"range":{"start":{"line":17,"character":8},"end":{"line":17,"character":8}},"newText":"string"}},{"label":"true","kind":21,"sortText":"00101","filterText":"true","insertTextFormat":1,"textEdit":{"range":{"start":{"line":17,"character":8},"end":{"line":17,"character":8}},"newText":"true"}},{"label":"uint","kind":7,"sortText":"00102","filterText":"uint","insertTextFormat":1,"textEdit":{"range":{"start":{"line":17,"character":8},"end":{"line":17,"character":8}},"newText":"uint"}},{"label":"uint16","kind":7,"sortText":"00103","filterText":"uint16","insertTextFormat":1,"textEdit":{"range":{"start":{"line":17,"character":8},"end":{"line":17,"character":8}},"newText":"uint16"}},{"label":"uint32","kind":7,"sortText":"00104","filterText":"uint32","insertTextFormat":1,"textEdit":{"range":{"start":{"line":17,"character":8},"end":{"line":17,"character":8}},"newText":"uint32"}},{"label":"uint64","kind":7,"sortText":"00105","filterText":"uint64","insertTextFormat":1,"textEdit":{"range":{"start":{"line":17,"character":8},"end":{"line":17,"character":8}},"newText":"uint64"}},{"label":"uint8","kind":7,"sortText":"00106","filterText":"uint8","insertTextFormat":1,"textEdit":{"range":{"start":{"line":17,"character":8},"end":{"line":17,"character":8}},"newText":"uint8"}},{"label":"uintptr","kind":7,"sortText":"00107","filterText":"uintptr","insertTextFormat":1,"textEdit":{"range":{"start":{"line":17,"character":8},"end":{"line":17,"character":8}},"newText":"uintptr"}},{"label":"main","kind":3,"detail":"func()","sortText":"00114","filterText":"main","insertTextFormat":1,"textEdit":{"range":{"start":{"line":17,"character":8},"end":{"line":17,"character":8}},"newText":"main"}},{"label":"readFileToString","kind":3,"detail":"func(filePath string) (string, error)","sortText":"00115","filterText":"readFileToString","insertTextFormat":1,"textEdit":{"range":{"start":{"line":17,"character":8},"end":{"line":17,"character":8}},"newText":"readFileToString"}},{"label":"TestColorize","kind":3,"detail":"func(t *testing.T)","sortText":"00116","filterText":"TestColorize","insertTextFormat":1,"textEdit":{"range":{"start":{"line":17,"character":8},"end":{"line":17,"character":8}},"newText":"TestColorize"}},{"label":"TestParse","kind":3,"detail":"func(t *testing.T)","sortText":"00117","filterText":"TestParse","insertTextFormat":1,"textEdit":{"range":{"start":{"line":17,"character":8},"end":{"line":17,"character":8}},"newText":"TestParse"}},{"label":"nil","kind":6,"sortText":"00118","filterText":"nil","insertTextFormat":1,"textEdit":{"range":{"start":{"line":17,"character":8},"end":{"line":17,"character":8}},"newText":"nil"}}]},"id":1}
`
	// Parse the JSON data into a CompletionResponse object
	var response CompletionResponse
	err := json.Unmarshal([]byte(jsonData), &response)
	if err != nil {
		fmt.Println("Error parsing JSON:", err)
		return
	}

	// Access the parsed CompletionResponse object
	fmt.Println("JSONRPC:", response.JSONRPC)
	fmt.Println("ID:", response.ID)
	fmt.Println("Result:")
	fmt.Println("	IsIncomplete:", response.Result.IsIncomplete)
	fmt.Println("	Items:")
	for _, item := range response.Result.Items {
		fmt.Println("		Label:", item.Label)
		fmt.Println("		Kind:", item.Kind)
		fmt.Println("		Detail:", item.Detail)
		fmt.Println("		Preselect:", item.Preselect)
		fmt.Println("		SortText:", item.SortText)
		fmt.Println("		FilterText:", item.FilterText)
		fmt.Println("		InsertTextFormat:", item.InsertTextFormat)
		fmt.Println("		TextEdit:")
		fmt.Println("			Range:")
		fmt.Println("				Start Line:", item.TextEdit.Range.Start.Line)
		fmt.Println("				Start Character:", item.TextEdit.Range.Start.Character)
		fmt.Println("				End Line:", item.TextEdit.Range.End.Line)
		fmt.Println("				End Character:", item.TextEdit.Range.End.Character)
		fmt.Println("			NewText:", item.TextEdit.NewText)
	}
}
