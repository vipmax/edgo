package main

import (
	"fmt"
	"os"
	"testing"
)

func TestLoggerInfo(t *testing.T) {
	err := os.Setenv("EDGO_LOGFILE", "edgo.log")
	if err != nil { fmt.Printf("Failed to set variable: %v", err); return }

	var logger = Logger{ }
	logger.start()
	go logger.info("async")
	logger.info("hello")
	logger.info("world")
}