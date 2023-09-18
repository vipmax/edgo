package io

import (
	"fmt"
	"github.com/rjeczalik/notify"
	"os"
	"path/filepath"
	"testing"
	"time"
)


func TestFileWatcher(t *testing.T) {
	tempDir := os.TempDir()
	fmt.Println("Temporary directory:", tempDir)
	file := filepath.Join(tempDir, "example.txt")
	os.Remove(file)
	os.WriteFile(file, []byte((`Hello, world!`)), 0644)


	fw := NewFileWatcher(100)
	fw.UpdateFile(file)
	fw.UpdateStats()

	updateChan := make(chan struct{})
	fw.StartWatch(func() {
		fmt.Println("file content changed")
		updateChan <- struct{}{}
	})

	go func() {
		time.Sleep(100 * time.Millisecond)
		os.WriteFile(file, []byte((`file content changed`)), 0644)
	}()


	select {
	case <-time.After(time.Second * 1):
		t.Error("File content did not change")
	case <-updateChan:
		// File content changed
		os.Remove(file)
		os.Remove(tempDir)
	}
}


func TestDirWatcher(t *testing.T) {
	tempDir := os.TempDir()
	fmt.Println("Temporary directory:", tempDir)
	file := filepath.Join(tempDir, "example.txt")

	updateChan := make(chan struct{})

	dw := NewDirWatcher(tempDir)
	dw.StartWatch(func(e notify.EventInfo) {
		switch e.Event() {
			case notify.Create: fmt.Println("create", e)
			case notify.Remove: fmt.Println("remove", e)
			default: fmt.Println("default", e)
		}
		updateChan <- struct{}{}
	})

	go func() {
		os.WriteFile(file, []byte((`Hello, world!`)), 0644)
	}()

	select {
	case <-time.After(time.Second * 1):
		t.Error("did not change")
	case <-updateChan:
		// File content changed
		os.Remove(file)
		os.Remove(tempDir)
	}
}