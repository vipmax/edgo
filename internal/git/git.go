package git

import (
	. "edgo/internal/utils"
	"fmt"
	gogit "github.com/go-git/go-git/v5"
	"github.com/sergi/go-diff/diffmatchpatch"
	"os"
	"path/filepath"
	"strings"
)

func GetLastCommitFileContent(filePath string) (string, error) {
	cwd, err := os.Getwd()
	if err != nil { return "", fmt.Errorf("error getting current working directory: %w", err) }

	r, err := gogit.PlainOpen(filepath.Join(cwd))
	if err != nil { return "", fmt.Errorf("error opening git repository: %w", err) }

	ref, err := r.Head()
	if err != nil { return "", fmt.Errorf("error getting repository HEAD: %w", err) }

	commit, err := r.CommitObject(ref.Hash())
	if err != nil { return "", fmt.Errorf("error getting commit object: %w", err) }

	tree, err := commit.Tree()
	if err != nil { return "", fmt.Errorf("error getting commit tree: %w", err) }

	file, err := tree.File(filePath)
	if err != nil { return "", fmt.Errorf("error getting file from tree: %w", err) }

	content, err := file.Contents()
	if err != nil { return "", fmt.Errorf("error getting file contents: %w", err) }

	return content, nil
}

// Diff, currently it is works bad, too much cpu usage, disable it off
func Diff(fileContent string, newFileContent string) (Set, Set) {
	//dmp := diffmatchpatch.New()

	diffs := dmp.DiffMain(fileContent, newFileContent, true)
	//blockDiffs := calcBlockDiff(fileContent, newFileContent)
	//fmt.Println(blockDiffs)

	added := make(Set)
	removed := make(Set)

	//for _, diff := range blockDiffs {
	//	if diff.Ope == 1 { 	added.Add(diff.NewLineNumber) }
	//}

	lineNum := 1
	for _, diff := range diffs {
		//lines := strings.Split(diff.Text, "\n")

		if diff.Type == diffmatchpatch.DiffInsert {
			added.Add(lineNum)
			//for _, ch := range diff.Text {
			//	added.Add(lineNum)
			//	if ch == '\n' { lineNum++ }
			//}

			count := strings.Count(diff.Text, "\n")
			for i := 0; i < count; i++ {
				lineNum ++
				added.Add(lineNum)
			}

		} else if diff.Type == diffmatchpatch.DiffDelete {
			//removed.Add(lineNum)
			//for _, ch := range diff.Text {
			//	removed.Add(lineNum)
			//	if ch == '\n' {  }
			//}

		} else  {
			lineNum += strings.Count(diff.Text, "\n")
		}

		//lineNum += strings.Count(diff.Text, "\n")

	}
	return added, removed
}


type blockDiff struct {
	Ope           Ope
	Text          string
	NewLineNumber int
	OldLineNumber int
}

var dmp = diffmatchpatch.New()
func calcBlockDiff(oldText, newText string) []blockDiff {

	a, b, c := dmp.DiffLinesToChars(oldText, newText)
	diffs := dmp.DiffMain(a, b, true)
	diffByLines := dmp.DiffCharsToLines(diffs, c)
	result := make([]blockDiff, len(diffByLines))
	newLineNum := 1
	oldLineNum := 1
	for i, diff := range diffByLines {
		f := blockDiff{
			Ope:           Ope(diff.Type),
			Text:          diff.Text,
			NewLineNumber: -1,
			OldLineNumber: -1,
		}
		inc := strings.Count(diff.Text, "\n")
		switch f.Ope {
		case 1:
			f.NewLineNumber = newLineNum
			newLineNum += inc
		case -1:
			f.OldLineNumber = oldLineNum
			oldLineNum += inc
		case 0:
			f.NewLineNumber = newLineNum
			f.OldLineNumber = oldLineNum
			newLineNum += inc
			oldLineNum += inc
		}
		result[i] = f
	}
	return result
}

type Ope int8

