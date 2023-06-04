package main

import (
	"fmt"
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
