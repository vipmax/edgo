package main

import (
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
	fmt.Println("")
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

