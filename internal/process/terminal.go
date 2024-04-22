package process

import (
	"bufio"
	"context"
	"fmt"
	"github.com/creack/pty"
	"io"
	"os"
	"os/exec"
	"strings"
	"sync"
	"time"
)

// Terminal represents a terminal session
type Terminal struct {
	Cmd     *exec.Cmd          // Command being executed
	Pty     *os.File           // Pty master file
	Lines   []string           // Output lines from the executed commands
	mutex   sync.Mutex         // Mutex to synchronize access to OutputLines
	Updates chan struct{}      // channel to notify about new lines
	stop    context.CancelFunc // function to cancel process
	Stopped bool               // true if process stopped
}

// NewTerminal creates a new Terminal instance
func NewTerminal() (*Terminal, error) {
	//shell := os.Getenv("SHELL")
	shell := "bash"
	if shell == "" {
		return nil, fmt.Errorf("SHELL environment variable not set")
	}

	//cmd := exec.Command(shell)

	//ctx, stop := signal.NotifyContext(context.Background(), os.Kill)
	ctx, stop := context.WithCancel(context.Background())
	cmd := exec.CommandContext(ctx, shell)
	cmd.Env = append(cmd.Env, "BASH_SILENCE_DEPRECATION_WARNING=1")

	ptyMaster, err := pty.Start(cmd)
	if err != nil {
		return nil, err
	}

	t := &Terminal{
		Cmd:          cmd,
		Pty:          ptyMaster,
		Lines:  make([]string, 0),
		mutex:        sync.Mutex{},
		Updates: make(chan struct{}),
		stop: stop,
	}

	t.start()

	go func() {
		err := cmd.Wait()
		if err != nil { fmt.Println(err) }
		fmt.Println("wait done ")
	}()

	return t, err
}

// ExecuteCommand runs a command on the terminal and reads the output line by line
func (t *Terminal) ExecuteCommand(command string ) error {
	_, err := t.Pty.Write([]byte(command + "\n"))
	if err != nil { return err }
	return nil
}

func (t *Terminal) start() {
	reader := bufio.NewReader(t.Pty)

	go func() {
		for {
			line, err := reader.ReadString('\n')
			if err != nil {
				if err == io.EOF { break }
				return
			}
			//line = stripansi.Strip(line)
			t.mutex.Lock()
			line = strings.ReplaceAll(line,"\r\n","")
			//line = stripansi.Strip(line)
			t.Lines = append(t.Lines, line)
			t.mutex.Unlock()
		}
	}()

	go func() {
		// if process output is too fast - it will be hard to read
		// The idea is to check output changes every 30ms
		// Write update message only if changes found
		lastMessagesLen := 0

		for !t.IsStopped() {
			t.mutex.Lock()
			currentLen := len(t.Lines)
			if currentLen != lastMessagesLen {
				t.Updates <- struct{}{}
				lastMessagesLen = currentLen
			}
			t.mutex.Unlock()
			<-time.After(time.Millisecond * time.Duration(30))
		}
		// this goroutine will be finished after process stopped
	}()


}


func (t *Terminal) Stop()  {
	_, err := t.Pty.Write([]byte{4})
	if err != nil { return }

	err = t.Pty.Close()
	if err != nil { return }

	//t.stop()

	t.mutex.Lock()
	close(t.Updates)
	t.Stopped = true
	t.mutex.Unlock()
}

func (t *Terminal) IsStopped() bool {
	t.mutex.Lock()
	defer t.mutex.Unlock()
	return t.Stopped
}

func (t *Terminal) GetLines(offset int) []string {
	t.mutex.Lock()
	defer t.mutex.Unlock()
	//return p.Lines[offset:]

	// Return a copy to avoid concurrent modification
	return append([]string{}, t.Lines[offset:]...)
}