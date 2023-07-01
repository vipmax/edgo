package process

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"strings"
)

type Process struct {
	Cmd  *exec.Cmd
	Out  chan string
	done chan struct{}
	ctx  context.Context
	stop context.CancelFunc
	Stopped bool
}

func NewProcess(command string, args ...string) *Process {
	ctx, stop := signal.NotifyContext(context.Background(), os.Kill)
	cmd := exec.CommandContext(ctx, command, args...)

	return &Process{
		Cmd:  cmd,
		Out:  make(chan string),
		done: make(chan struct{}),
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
			p.Out <- line
		}
	}()

	go func() {
		stderr, _ := p.Cmd.StderrPipe()
		scanner := bufio.NewScanner(stderr)
		for scanner.Scan() {
			line := scanner.Text()
			p.Out <- line
		}
	}()

	go p.runCmd()
}


func (p *Process) runCmd() {

	p.Out <- fmt.Sprintf("%s %s",
		p.Cmd.Path, strings.Join(p.Cmd.Args[1:], " "),
	)

	err := p.Cmd.Run()
	if err != nil {
		p.Out <- "Error: " + err.Error()
	}

	p.Out <- ""
	p.Out <- fmt.Sprintf("Process %d finished with exit code %d",
		p.Cmd.ProcessState.Pid(), p.Cmd.ProcessState.ExitCode(),
	)
	p.Stopped = true

}


func (p *Process) Stop() {
	if p.Stopped { return }
	p.stop()
	p.Stopped = true
}