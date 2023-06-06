package main

import (
	"fmt"
	"os"
	"path"
	"testing"
)


func TestGoLangCompletion(t *testing.T) {
	dir, _ := os.Getwd()
	file := dir + "/lsp_test.go"
	text, _ := readFileToString(file)
	
	fmt.Println("starting lsp server")
	
	lsp := LspClient{}
	lsp.start("go")
	lsp.init(dir)
	lsp.didOpen(file)

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

func TestPythonCompletion(t *testing.T) {
	dir := "/Users/max/apps/python/editor/src/"
	file := path.Join(dir, "logger.py")
	text, _ := readFileToString(file)

	fmt.Println("starting lsp server")

	lsp := LspClient{}
	lsp.start("python")
	lsp.init(dir)
	lsp.didOpen(file)

	completion, _ := lsp.completion(file, text, 8-1, 20)
	fmt.Println("completion", completion)

	var options []string
	items := completion.Result.Items
	for _, item := range items {
		options = append(options, item.Label)
	}

	fmt.Println("options", options)
	fmt.Println("ending lsp server")
}

func TestTypescriptCompletion(t *testing.T) {
	dir := "/Users/max/apps/ts/lsp-examples/"
	file := path.Join(dir, "lsp-test-ts.ts")
	text, _ := readFileToString(file)

	fmt.Println("starting lsp server for ", file)

	lsp := LspClient{}
	lsp.start("typescript")
	lsp.init(dir)
	lsp.didOpen(file)

	completion, _ := lsp.completion(file, text, 31-1, 5)
	fmt.Println("completion", completion)

	var options []string
	items := completion.Result.Items
	for _, item := range items {
		options = append(options, item.Label)
	}

	fmt.Println("options", options)
	fmt.Println("ending lsp server")
}

func TestRustCompletion(t *testing.T) {
	dir := "/Users/max/apps/rust/lsp-examples/"
	file := path.Join(dir, "lsp-test-ts.ts")
	text, _ := readFileToString(file)

	fmt.Println("starting lsp server for ", file)

	lsp := LspClient{}
	lsp.start("typescript")
	lsp.init(dir)
	lsp.didOpen(file)

	completion, _ := lsp.completion(file, text, 31-1, 5)
	fmt.Println("completion", completion)

	var options []string
	items := completion.Result.Items
	for _, item := range items {
		options = append(options, item.Label)
	}

	fmt.Println("options", options)
	fmt.Println("ending lsp server")
}

