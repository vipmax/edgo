package git

import (
	"fmt"
	"github.com/pmezard/go-difflib/difflib"
	"github.com/sergi/go-diff/diffmatchpatch"
	"log"
	"os"
	"strings"
	"testing"
)


func TestGit(t *testing.T) {
	//os.Chdir("../../")

	fname := "internal/config/config.go"

	fileContent, err := GetLastCommitFileContent(fname)
	if err != nil { log.Fatal(err) }

	bytes, err := os.ReadFile(fname)
	if err != nil { log.Fatal(err) }
	newFileContent := string(bytes)

	added, removed := Diff(fileContent, newFileContent)

	fmt.Println("Added line numbers:")
	added.Print()

	fmt.Println("Removed line numbers:")
	removed.Print()
}




func TestGit2(t *testing.T) {

	str1 := `
    Line 1
	Line 2
	Line 3
	Line 4
	Line 5`

	str2 := `
    Line 1
	Line 2 changed
	Line 4
	Line 5
	Line 6`

	added, removed := Diff(str1, str2)

	fmt.Println("Added line numbers:")
	added.Print()

	fmt.Println("Removed line numbers:")
	removed.Print()
}

func TestGit22(t *testing.T) {

	str1 := `Line 1`

	str2 := `Line 2`

	added, removed := Diff(str1, str2)

	fmt.Println("Added line numbers:")
	added.Print()

	fmt.Println("Removed line numbers:")
	removed.Print()
}

func TestGit3(t *testing.T) {
	diff := difflib.UnifiedDiff{
		A:        difflib.SplitLines("foo\nbar\n"),
		B:        difflib.SplitLines("foo\nbaz\n"),
		FromFile: "Original",
		ToFile:   "Current",
		Context:  3,
	}
	text, _ := difflib.GetUnifiedDiffString(diff)
	fmt.Printf(text)
}

func TestGit4(t *testing.T) {
	fname := "tmux.conf"

	fileContent, err := GetLastCommitFileContent(fname)
	if err != nil { log.Fatal(err) }

	bytes, err := os.ReadFile(fname)
	if err != nil { log.Fatal(err) }
	newFileContent := string(bytes)

	diff := difflib.UnifiedDiff{
		A:        difflib.SplitLines(fileContent),
		B:        difflib.SplitLines(newFileContent),
		FromFile: "Original",
		ToFile:   "Current",
		Context:  3,
	}
	text, _ := difflib.GetUnifiedDiffString(diff)
	fmt.Printf(text)

	lines := strings.Split(text, "\n")

	for _, line := range lines {
		if len(line) > 0 {
			// '-' denotes line deletion
			if line[0] == '-' && line[1] != '-' {
				fmt.Println("Deleted: ", line[1:])
			}

			// '+' denotes line addition
			if line[0] == '+' && len(line) == 1  || line[0] == '+' && line[1] != '+' {
				fmt.Println("Added: ", line[1:])
			}

			// '@@ -1,5 +1,6 @@' denotes line range of changes
			if line[0] == '@' && line[1] == '@' {
				// split the line to get the line numbers
				parts := strings.Split(line, " ")
				// parts[1] is the line range in the original text
				// parts[2] is the line range in the current text
				fmt.Println("Line range in original text: ", parts[1])
				fmt.Println("Line range in current text: ", parts[2])
			}
		}
	}
}

func TestGit5(t *testing.T) {

	fname := "internal/utils/utils.go"

	fileContent, _ := GetLastCommitFileContent(fname)
	bytes, _ := os.ReadFile(fname)
	newFileContent := string(bytes)


	//const (
	//	text1 = "Lorem ipsum dolor."
	//	text2 = "Lorem dolor sit amet."
	//)

	dmp := diffmatchpatch.New()

	diffs := dmp.DiffMain(fileContent, newFileContent, false)

	fmt.Println(dmp.DiffPrettyText(diffs))
}