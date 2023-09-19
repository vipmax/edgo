package process

import (
	"bufio"
	"context"
	"fmt"
	"github.com/stretchr/testify/assert"
	"os"
	"os/exec"
	"os/signal"
	"testing"
	"time"
)


func TestKillProcess(t *testing.T) {
	os.Chdir("../../")
	fmt.Println(os.Getwd())
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)

	go run(ctx)

	time.Sleep(3 * time.Second)

	stop()

	time.Sleep(1 * time.Second)
}

func run(ctx context.Context) {
	//cmd := exec.CommandContext(ctx, "python3", "atest.py")
	//cmd := exec.CommandContext(ctx, "go", "run", "cmd/test/main.go")
	cmd := exec.CommandContext(ctx, "sleep", "10")
	cmd.Env = append(os.Environ(), "PYTHONUNBUFFERED=1")

	stdout, _ := cmd.StdoutPipe()

	go func() {
		scanner := bufio.NewScanner(stdout)
		for scanner.Scan() {
			line := scanner.Text()
			fmt.Println(line)
		}

		fmt.Println("done")

	}()

	cmd.Start()
}


func TestNewProcess(t *testing.T) {
	cmd := NewProcess("echo", "hello")
	assert.NotNil(t, cmd)
	assert.NotNil(t, cmd.Cmd)
	assert.NotNil(t, cmd.Lines)
	assert.NotNil(t, cmd.Updates)
}

func TestProcess_StartStop(t *testing.T) {
	cmd := NewProcess("echo", "hello")
	assert.NotNil(t, cmd)

	// Start the process
	cmd.Start()
	// Allow some time for the process to start
	time.Sleep(100 * time.Millisecond)

	// Check if it started successfully
	assert.False(t, cmd.Stopped)

	// Stop the process
	cmd.Stop()
	// Allow some time for the process to stop
	time.Sleep(100 * time.Millisecond)

	// Check if it stopped successfully
	assert.True(t, cmd.Stopped)
}

func TestProcess_StopTwice(t *testing.T) {
	cmd := NewProcess("echo", "hello")
	assert.NotNil(t, cmd)

	// Start the process
	cmd.Start()
	time.Sleep(100 * time.Millisecond) // Allow some time for the process to start

	// Stop the process twice
	cmd.Stop()
	cmd.Stop() // Second stop should be a no-op

	// Check if it stopped successfully
	assert.True(t, cmd.Stopped)
}

func TestProcessOutput(t *testing.T) {
	cmd := NewProcess("echo", "hello")
	assert.NotNil(t, cmd)

	cmd.Start() // Start the process

	time.Sleep(10 * time.Millisecond) // Allow some time for the process to start

	for range cmd.Updates { } // wait for no updates anymore

	lines := cmd.GetLines(0)

	for _, line := range lines {
		fmt.Println("-> ",line)
	}

	// Check if the output was captured correctly
	assert.Len(t, lines, 4)
	assert.Contains(t, lines[0], "echo hello")
	assert.Contains(t, lines[1], "hello")
	assert.Equal(t, lines[2], "")
	assert.Contains(t, lines[3], "finished with exit code 0")
	assert.Equal(t, cmd.IsStopped(), true)
}

func TestProcessErrorOutput(t *testing.T) {
	cmd := NewProcess("sh", "-c", "echo hello >&2")
	assert.NotNil(t, cmd)

	cmd.Start() // Start the process

	time.Sleep(10 * time.Millisecond) // Allow some time for the process to start

	for range cmd.Updates { } // wait for no updates anymore

	lines := cmd.GetLines(0)

	for _, line := range lines {
		fmt.Println("-> ",line)
	}

	// Check if the output was captured correctly
	assert.Len(t, lines, 4)
	assert.Contains(t, lines[0], "echo hello")
	assert.Contains(t, lines[1], "hello")
	assert.Equal(t, lines[2], "")
	assert.Contains(t, lines[3], "finished with exit code 0")
	assert.Equal(t, cmd.IsStopped(), true)
}

func TestProcessStop(t *testing.T) {
	cmd := NewProcess("sleep", "10") // sleep 10 seconds
	assert.NotNil(t, cmd)

	cmd.Start() // Start the process
	time.Sleep(10 * time.Millisecond)

	assert.Equal(t, cmd.IsStopped(), false)

	cmd.Stop() // Stop the process
	time.Sleep(10 * time.Millisecond)

	assert.Equal(t, cmd.IsStopped(), true)
}


func TestProcessCommandNotFound(t *testing.T) {
	cmd := NewProcess("sleepp", "10")
	assert.NotNil(t, cmd)

	cmd.Start() // Start the process
	// Allow some time for the process to start
	time.Sleep(10 * time.Millisecond)

	for range cmd.Updates { } // wait for no updates anymore

	lines := cmd.GetLines(0)
	for _, line := range lines {
		fmt.Println("-> ",line)
	}

	// Check if the output was captured correctly
	assert.Len(t, lines, 2)
	assert.Equal(t, lines[0], "sleepp 10")
	assert.Contains(t, lines[1], "executable file not found in $PATH")
	assert.Equal(t, cmd.IsStopped(), true)
}

