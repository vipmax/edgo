package main

import (
	"fmt"
	"os"
	"testing"
)

func TestColorize(t *testing.T) {
	dir, _ := os.Getwd()
	file := dir + "/lsp_test.go"
	text, _ := readFileToString(file)

	fmt.Println("starting lsp server")

	lsp := LspClient{}
	lsp.start()
	lsp.init(dir)
	lsp.didOpen(file)
	//lsp.definition(file, 21, 7)
	completion := lsp.completion(file, text, 18-1, 8)
	fmt.Println("completion", completion)

	var options []string
	items, ok := completion["result"].(map[string]interface{})["items"]
	if ok {
		for _, item := range items.([]interface{}) {
			if item, ok := item.(map[string]interface{}); ok {
				if label, ok := item["label"].(string); ok {
					options = append(options, label)
				}
			}
		}
	}

	fmt.Println("options", options)
	fmt.Println("ending lsp server")
}
