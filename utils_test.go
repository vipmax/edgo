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

func TestDetectGoLang(t *testing.T) {
	language := detectLang("highlighter_test.go")
	fmt.Println(language)
	if language != "Go"{
		t.Error("language must be Go", language)
	}
}

func TestDetectPythonLang(t *testing.T) {
	language := detectLang("test.py")
	fmt.Println(language)
	if language != "Python"{
		t.Error("language must be Python", language)
	}
}
