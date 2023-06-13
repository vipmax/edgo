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
	if language != "go"{
		t.Error("language must be Go", language)
	}
}

func TestDetectPythonLang(t *testing.T) {
	language := detectLang("test.py")
	fmt.Println(language)
	if language != "python"{
		t.Error("language must be Python", language)
	}
}

func TestNoSelection(t *testing.T) {
	var content = [][]rune{
		[]rune("Hello, world!"),
		[]rune("How are you doing today?"),
		[]rune("I hope you're doing well."),
	}

	got := getSelectionString(content, 0, 0, 0, 0)
	want := ""
	if got != want {
		t.Errorf("getSelectionString() = %v, want %v", got, want)
	}
}

func TestSingleCharacterSelection(t *testing.T) {
	var content = [][]rune{
		[]rune("Hello, world!"),
		[]rune("How are you doing today?"),
		[]rune("I hope you're doing well."),
	}

	got := getSelectionString(content, 0, 0, 1, 0)
	want := "H"
	if got != want {
		t.Errorf("getSelectionString() = %v, want %v", got, want)
	}
}

func TestMultipleCharacterSelection(t *testing.T) {
	var content = [][]rune{
		[]rune("Hello, world!"),
		[]rune("How are you doing today?"),
		[]rune("I hope you're doing well."),
	}

	got := getSelectionString(content, 0, 0, 5, 0)
	want := "Hello"
	if got != want {
		t.Errorf("getSelectionString() = %v, want %v", got, want)
	}
}

func TestMultipleLineSelection1(t *testing.T) {
	var cont = [][]rune{
		[]rune("Hello, world!"),
		[]rune("How are you doing today?"),
		[]rune("I hope you're doing well."),
	}

	got := getSelectionString(cont, 0, 0, 5, 1)
	want := "Hello, world!\nHow a"
	if got != want {
		t.Errorf("getSelectionString() = %v, want %v", got, want)
	}
}
func TestMultipleLineSelection2(t *testing.T) {
	var cont = [][]rune{
		[]rune("Hello, world!"),
		[]rune("How are you doing today?"),
		[]rune("I hope you're doing well."),
	}

	got := getSelectionString(cont, 0, 0, 11, 1)
	want := "Hello, world!\nHow are you"
	if got != want {
		t.Errorf("getSelectionString() = %v, want %v", got, want)
	}
}

func TestMultipleLineSelection3(t *testing.T) {
	var cont = [][]rune{
		[]rune("Hello, world!"),
		[]rune("How are you doing today?"),
		[]rune("I hope you're doing well."),
	}

	got := getSelectionString(cont, 6, 0, 23, 1)
	want := " world!\nHow are you doing today"
	if got != want {
		t.Errorf("getSelectionString() =\ngot=%s \nwant= %s", got, want)
	}
}