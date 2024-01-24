package process

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"os/signal"
	"strings"
	"sync"
	"time"

	"github.com/acarl005/stripansi"
)

type Process struct {
	Cmd       *exec.Cmd          // command to run
	cancelF   context.CancelFunc // function to cancel process
	Stopped   bool               // true if process stopped
	Lines     []string           // all stdout/err lines
	muLines   sync.Mutex         // Mutex to protect access to Lines
	muStopped sync.Mutex         // Mutex to protect access to Stopped
	Updates   chan struct{}      // channel to notify about new lines
	UpdateInterval int           // time interval to fire updates
	Stdin     chan string        // channel to send input to stdin
}


func NewProcess(command string, args ...string) *Process {
	ctx, stop := signal.NotifyContext(context.Background(), os.Kill)
	cmd := exec.CommandContext(ctx, command, args...)

	return &Process{
		Cmd:     cmd,
		Lines:   []string{},
		Updates: make(chan struct{}),
		cancelF: stop,
		UpdateInterval: 30,
		Stdin:   make(chan string),
	}
}

func (p *Process) Start() {
	wg := sync.WaitGroup{}
	wg.Add(2)
	// if process is too fast, reading std out/err goroutines
	// won't have time to start, and output will be missed/empty

	go func() {
		stdout, _ := p.Cmd.StdoutPipe()
		scanner := bufio.NewScanner(stdout)
		wg.Done()
		for scanner.Scan() {
			line := stripansi.Strip(scanner.Text())
			p.appendLine(line)
			// it is not good idea to send updates here,
			// if process output is too fast, it will be too many updates
			// channels are too slow for this
			// p.Updates <- struct{}{}
		}
		// this goroutine will be finished after process exit
	}()

	go func() {
		stderr, _ := p.Cmd.StderrPipe()
		scanner := bufio.NewScanner(stderr)
		wg.Done()
		for scanner.Scan() {
			line := stripansi.Strip(scanner.Text())
			p.appendLine(line)
		}
		// this goroutine will be finished after process exit
	}()

	go func() {
		// if process output is too fast - it will be hard to read
		// The idea is to check output changes every 30ms
		// Write update message only if changes found
		lastMessagesLen := 0

		for !p.IsStopped() {
			p.muLines.Lock()
			currentLen := len(p.Lines)
			if currentLen != lastMessagesLen {
				p.Updates <- struct{}{}
				lastMessagesLen = currentLen
			}
			p.muLines.Unlock()
			<-time.After(time.Millisecond * time.Duration(p.UpdateInterval))
		}
		// this goroutine will be finished after process exit
	}()

	wg.Wait()
	go p.runCmd()
}

func (p *Process) appendLine(line string) {
	p.muLines.Lock()
	defer p.muLines.Unlock()
	p.Lines = append(p.Lines, line)
}

func (p *Process) IsStopped() bool {
	p.muStopped.Lock()
	defer p.muStopped.Unlock()
	return p.Stopped
}

func (p *Process) GetExitCode() int {
	p.muStopped.Lock()
	defer p.muStopped.Unlock()
	if p.Cmd.ProcessState == nil { return -1 }
	return p.Cmd.ProcessState.ExitCode()
}

func (p *Process) GetLines(offset int) []string {
	p.muLines.Lock()
	defer p.muLines.Unlock()
	//return p.Lines[offset:]

	// Return a copy to avoid concurrent modification
	return append([]string{}, p.Lines[offset:]...)
}

func (p *Process) runCmd() {
	p.appendLine(fmt.Sprintf("%s %s",
		p.Cmd.Path, strings.Join(p.Cmd.Args[1:], " "),
	))
	p.Updates <- struct{}{}
	start := time.Now()

	err := p.Cmd.Run() // its blocks until process exiting
	if err != nil {
		p.appendLine( "Error: " + err.Error())
	}

	elapsed := time.Since(start)

	if p.Cmd.ProcessState != nil {
		p.appendLine("")
		p.appendLine(fmt.Sprintf(
			"Process %d finished with exit code %d, elapsed %s",
			p.Cmd.ProcessState.Pid(), p.Cmd.ProcessState.ExitCode(), elapsed.String(),
		))
	}

	p.muStopped.Lock()
	p.Stopped = true
	p.muStopped.Unlock()

	p.Updates <- struct{}{}
	close(p.Updates)
}


func (p *Process) Stop() {
	if p.IsStopped() { return }
	p.cancelF()
	p.muStopped.Lock()
	p.Stopped = true
	p.muStopped.Unlock()
}

func (p *Process) GetPid() int {
	p.muStopped.Lock()
	defer p.muStopped.Unlock()
	if p.Cmd.Process != nil { return p.Cmd.Process.Pid }
	return -1
}

func (p *Process) WriteStdin(input string) {
	if p.IsStopped() {
		panic("Process is already stopped. Cannot write to stdin.")
		return
	}

	//p.muLines.Lock()
	//defer p.muLines.Unlock()

	//if p.Cmd.Stdin == nil {
	//	fmt.Println("Process stdin is not available.")
	//	return
	//}

	stdin, err := p.Cmd.StdinPipe()

	_, err = io.WriteString(stdin, input + "\n")
	if err != nil {
		fmt.Println("Error writing to process stdin:", err)
	}
	stdin.Close()

	//_, err = stdin.Write([]byte(input + "\n"))
	//if err != nil {
	//	fmt.Println("Error writing to process stdin:", err)
	//}
}