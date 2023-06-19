package main

import (
	"fmt"
	"github.com/alecthomas/chroma"
	"github.com/alecthomas/chroma/styles"
	"testing"
)

func TestColorize(t *testing.T) {
	file := "highlighter_test.go"
	filecontent, _ := readFileToString(file)

	h := Highlighter{}
	characterColors := h.colorize(filecontent, file)

	for _, color := range characterColors {
		fmt.Println(color)
	}
}


func TestGetStyle(t *testing.T) {
	style := styles.Get("github")
	if style == nil { style = styles.Fallback }
	fmt.Println(style)
	fmt.Println(style.Get(chroma.Comment).Background)
}