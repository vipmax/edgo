package main

import (
	"fmt"
	"testing"
)


func TestFormat(t *testing.T) {
	leftText := "Left"; rightText := "Right"; maxWidth := 30
	formattedText := formatText(leftText, rightText, maxWidth)
	fmt.Println(formattedText)
}
