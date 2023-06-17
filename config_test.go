package main

import (
	"fmt"
	"testing"
)

func TestReadConfig(t *testing.T) {
	// Read config
	config := GetConfig()

	// Print config
	for _, lang := range config.Langs {
		fmt.Printf("Name: %s, Lsp: %s, Comment: %s, Tab Width: %d\n",
			lang.Name, lang.Lsp, lang.Comment, lang.TabWidth,
		)
	}
}
