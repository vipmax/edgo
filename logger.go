package main

import (
	"log"
	"os"
	"strings"
	"time"
)

type Logger struct {
	isEnabled bool
	file   *os.File
	stream chan string
	logger *log.Logger
	layout string
}

func (this *Logger) start() {
	logfilename, exists := os.LookupEnv("EDGO_LOGFILE")
	if !exists { this.isEnabled = false; return }

	this.isEnabled = true

	file, err := os.OpenFile(logfilename, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil { log.Fatal(err) }
	this.file = file

	this.logger = log.New(file, "", log.LstdFlags)
	this.logger.SetFlags(0)
	this.logger.SetOutput(file)
	this.layout = "2006-01-02 15:04:05.000"

	this.stream = make(chan string)

	go func() {
		for message := range this.stream {
			this.log(message)
		}
	}()

}

func (this *Logger) log(message string) {
	if !this.isEnabled { return }
	now := time.Now().Format(this.layout)
	this.logger.Printf("%s %s", now, message)
}

func (this *Logger) info(args ...string) {
	if !this.isEnabled { return }
	message := strings.Join(args, " ")
	this.stream <- message
}
func (this *Logger) error(args ...string) {
	if !this.isEnabled { return }
	message :=  "[error]" + strings.Join(args, " ")
	this.stream <- message
}

func (this *Logger) stop() {
	if !this.isEnabled { return }
	close(this.stream)
	err := this.file.Close()
	if err != nil { return }
}