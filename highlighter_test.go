package main

import (
	"fmt"
	"github.com/alecthomas/chroma"
	"github.com/alecthomas/chroma/styles"
	"testing"
)

func TestColorizle(t *testing.T) {
	file := "highlighter_test.golang"
	filecontent, _ := readFileToString(file)

	h := Highlighter{}

	characterColors := h.colorize(filecontent, file)
	
	for _, color := range characterColors {
		fmt.Println(color)
	}
}

func TestColorize(t *testing.T) {
	file := "highlighter_test.golang"
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

func TestRgbToAnsi(t *testing.T) {
	style := styles.Get("github")
	if style == nil { style = styles.Fallback }
	fmt.Println(style)
	colour := style.Get(chroma.Comment).Background

	ansi256 := RgbToAnsi256(colour.Red(), colour.Green(), colour.Blue())
	fmt.Println("ansi256", ansi256)
}