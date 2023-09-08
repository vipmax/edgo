package editor

import (
	"fmt"
	"os"
	"testing"
)

func TestProcessDirectory(t *testing.T) {

	// Get the current directory
	dir, err := os.Getwd()
	if err != nil {
		fmt.Println("Failed to get current directory:", err)
		return
	}

	// Recursively process the directory
	fileInfo, err := ReadDirTree(dir, "", true,0)
	if err != nil {
		fmt.Println("Failed to process directory:", err)
		return
	}

	// Print the result
	//fmt.Printf("%+v\n", fileInfo)

	PrintTree(fileInfo, 0)
	
}

func TestTreeSize(t *testing.T) {
	dir, _ := os.Getwd()
	tree, _ := ReadDirTree(dir,"", true,0)
	size := TreeSize(tree, 0)
	fmt.Println("size", size)
}

func TestGetSelected(t *testing.T) {
	dir, _ := os.Getwd()
	tree, _ := ReadDirTree(dir, "", true,0)
	found, fi := GetSelected(tree, 13)
	fmt.Println("selected", found, fi)
}