package config

import (
	"fmt"
	"testing"
)

func TestReadConfig(t *testing.T) {
	// Read conf
	conf := GetConfig()

	// Print config
	for _, lang := range conf.Langs {
		fmt.Printf("Name: %s, Lsp: %s, Comment: %s, Tab Width: %d\n",
			lang.Name, lang.Lsp, lang.Comment, lang.TabWidth,
		)
	}
}