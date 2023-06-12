package main

import (
	"fmt"
	"os"
	"path"
	"testing"
	"time"
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

func TestScalaCompletion(t *testing.T) {
	dir := "/Users/max/apps/scala/chrome4s"
	file := path.Join(dir, "/src/main/scala/chrome4s/Main.scala")
	text, _ := readFileToString(file)

	fmt.Println("starting lsp server for ", file)

	lsp := LspClient{}
	lsp.start("scala")
	lsp.init(dir)
	lsp.didOpen(file)
	time.Sleep(3*time.Second)
	completion, _ := lsp.completion(file, text, 17-1, 8)
	fmt.Println("completion", completion)

	var options []string
	items := completion.Result.Items
	for _, item := range items {
		options = append(options, item.Label)
	}

	fmt.Println("options", options)
	fmt.Println("ending lsp server")
}


func TestGoLangHover(t *testing.T) {
	dir, _ := os.Getwd()
	file := dir + "/lsp_test.go"

	fmt.Println("starting lsp server")

	lsp := LspClient{}
	lsp.start("go")
	lsp.init(dir)
	lsp.didOpen(file)

	hover, _ := lsp.hover(file,18-1, 13)
	fmt.Println("hover range: ", hover.Result.Range)
	fmt.Println("hover content:\n", hover.Result.Contents.Value)

	fmt.Println("ending lsp server")
}


func TestGoLangSignatureHelp(t *testing.T) {
	dir, _ := os.Getwd()
	file := dir + "/lsp_test.go"

	fmt.Println("starting lsp server")

	lsp := LspClient{}
	lsp.start("go")
	lsp.init(dir)
	lsp.didOpen(file)

	response, _ := lsp.signatureHelp(file,14-1, 36)
	fmt.Println("signatureHelp: ", response)

	fmt.Println("ending lsp server")
}
