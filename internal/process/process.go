package process

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"strings"
	"time"
)

type Process struct {
	Cmd  *exec.Cmd
	Out  chan string
	done chan struct{}
	ctx  context.Context
	stop context.CancelFunc
	Stopped bool

	Lines []string
	Update chan struct{}
}

func NewProcess(command string, args ...string) *Process {
	ctx, stop := signal.NotifyContext(context.Background(), os.Kill)
	cmd := exec.CommandContext(ctx, command, args...)

	return &Process{
		Cmd:  cmd,
		Out:  make(chan string),
		done: make(chan struct{}),
		Lines: []string{},
		Update: make(chan struct{}),
		ctx: ctx,
		stop: stop,
	}
}

func (p *Process) Start() {
	go func() {
		stdout, _ := p.Cmd.StdoutPipe()
		scanner := bufio.NewScanner(stdout)

		for scanner.Scan() {
			line := scanner.Text()
			p.Lines = append(p.Lines, line)
			//p.Out <- line
		}

	}()

	go func() {
		stderr, _ := p.Cmd.StderrPipe()
		scanner := bufio.NewScanner(stderr)
		for scanner.Scan() {
			line := scanner.Text()
			p.Lines = append(p.Lines, line)
			//p.Out <- line // slow for fast rate output
		}
	}()

	go func() {
		// idea is to check output changes every 10ms, update only if changes found
		lastMessagesLen := 0

		for !p.Stopped {
			<- time.After(time.Millisecond * 10)
			currentLen := len(p.Lines)
			if currentLen != lastMessagesLen {
				p.Update <- struct{}{}
				lastMessagesLen = currentLen
			}
		}
	}()

	go p.runCmd()
}


func (p *Process) runCmd() {

	p.Lines = append(p.Lines, fmt.Sprintf("%s %s",
		p.Cmd.Path, strings.Join(p.Cmd.Args[1:], " "),
	))
	p.Update <- struct{}{}

	err := p.Cmd.Run()
	if err != nil {
		p.Lines = append(p.Lines, "Error: " + err.Error())
	}

	p.Lines = append(p.Lines, "")
	p.Lines = append(p.Lines, fmt.Sprintf("Process %d finished with exit code %d",
		p.Cmd.ProcessState.Pid(), p.Cmd.ProcessState.ExitCode(),
	))
	p.Stopped = true
	p.Update <- struct{}{}

}


func (p *Process) Stop() {
	if p.Stopped { return }
	p.stop()
	p.Stopped = true
}