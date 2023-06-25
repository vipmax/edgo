package logger

import (
	"fmt"
	"os"
	"testing"
)

func TestLoggerInfo(t *testing.T) {
	err := os.Setenv("EDGO_LOGFILE", "edgo.log")
	if err != nil { fmt.Printf("Failed to set variable: %v", err); return }

	var logger = Logger{ }
	logger.Start()
	go logger.Info("async")
	logger.Info("hello")
	logger.Info("world")
}