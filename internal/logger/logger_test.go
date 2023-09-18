package logger

import (
	"fmt"
	"os"
	"slices"
	"strings"
	"testing"
	"time"
)

func TestLoggerInfo(t *testing.T) {
	err := os.Setenv("EDGO_LOG", "edgo.log")
	if err != nil { fmt.Printf("Failed to set variable: %v", err); return }

	var logger = Logger{ }
	logger.Start()

	logger.Info("async")
	logger.Info("hello")
	logger.Info("world")

	time.Sleep(1 * time.Second)

	bytes, _ := os.ReadFile("edgo.log")
	content := string(bytes)
	lines := strings.Split(content, "\n")
	lines = slices.Delete(lines, len(lines)-1, len(lines))

	if len(lines) != 3 {
		t.Errorf("Expected %d, got %d", 3, len(lines))
	}

	os.Remove("edgo.log")
}