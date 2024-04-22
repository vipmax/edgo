package process

import (
	"fmt"
	"testing"
)

func TestTerminalStart(t *testing.T) {
	terminal, err := NewTerminal()
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	defer terminal.Pty.Close()

	// Example: Execute "ls" command on the terminal
	err = terminal.ExecuteCommand("ls")
	if err != nil {
		fmt.Println("Error:", err)
	}

	// Access all output lines
	fmt.Println("Output Lines:")
	terminal.mutex.Lock()
	for _, line := range terminal.Lines {
		fmt.Print(line)
	}
	terminal.mutex.Unlock()
}
