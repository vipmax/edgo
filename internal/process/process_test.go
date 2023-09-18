package process

import (
	"bufio"
	"context"
	. "fmt"
	"github.com/stretchr/testify/assert"
	"os"
	"os/exec"
	"os/signal"
	"syscall"
	"testing"
	"time"
)

func TestProcess(t *testing.T) {
	os.Chdir("../../")
	Println(os.Getwd())

	//process := NewProcess("python3", "atest.py")
	process := NewProcess("go", "run", "cmd/test/main.go")
	process.Cmd.Env = append(os.Environ(), "PYTHONUNBUFFERED=1")

	go func() {
		for line := range process.Out {
			Println("Output:", line)
		}
	}()

	process.Start()

	//Println("Process started with PID:", process.Cmd.Process.Pid)

	// Kill the process after 5 seconds
	time.Sleep(3 * time.Second)
	process.Stop()
	//Println("Process killed with PID:", process.Cmd.Process.Pid)

	time.Sleep(30 * time.Second)

	Println("Child process finished.")
}

func TestProcessKill(t *testing.T) {
	cmd := exec.Command("sleep", "100", "&")
	err := cmd.Start()
	if err != nil { Println(err) }

	Println("Process started with PID:", cmd.Process.Pid)

	// Kill the process after 5 seconds
	time.Sleep(5 * time.Second)

	err = cmd.Process.Kill()
	if err != nil { Println(err) }
	cmd.Process.Release()

	Println("Process killed with PID:", cmd.Process.Pid)
	time.Sleep(30 * time.Second)
}


func TestKill2(t *testing.T) {
	cmd := exec.Command("sleep", "100")
	cmd.SysProcAttr = &syscall.SysProcAttr{Setsid: true}
	cmd.Start()

	Printf("Parent PID: %d\n", cmd.Process.Pid)

	time.Sleep(3 * time.Second)
	go func( ) {
		sig := make(chan os.Signal, 1)
		signal.Notify(sig, os. Interrupt)
		<-sig
		signal.Reset()
	}()
	time.Sleep(30 * time.Second)
}

func TestKill3(t *testing.T) {
	os.Chdir("../../")
	Println(os.Getwd())
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)

	go run(ctx)

	time.Sleep(3 * time.Second)

	stop()

	time.Sleep(30 * time.Second)
}

func run(ctx context.Context) {
	//cmd := exec.CommandContext(ctx, "python3", "atest.py")
	cmd := exec.CommandContext(ctx, "go", "run", "cmd/test/main.go")
	cmd.Env = append(os.Environ(), "PYTHONUNBUFFERED=1")

	stdout, _ := cmd.StdoutPipe()

	go func() {
		scanner := bufio.NewScanner(stdout)
		for scanner.Scan() {
			line := scanner.Text()
			Println(line)
		}

		Println("done")

	}()

	cmd.Start()
}


func TestNewProcess(t *testing.T) {
	cmd := NewProcess("echo", "hello")
	assert.NotNil(t, cmd)
	assert.NotNil(t, cmd.Cmd)
	assert.NotNil(t, cmd.Out)
	assert.NotNil(t, cmd.done)
	assert.NotNil(t, cmd.Lines)
	assert.NotNil(t, cmd.Update)
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
	// Define a simple command that outputs "hello"
	cmd := NewProcess("echo", "hello")
	assert.NotNil(t, cmd)

	// Start the process
	cmd.Start()
	time.Sleep(100 * time.Millisecond) // Allow some time for the process to start

	for range cmd.Update {
		Println(cmd.Lines)
	}

	// Check if the output was captured correctly
	assert.Len(t, cmd.Lines, 4)
	assert.Contains(t, cmd.Lines[0], "echo hello")
	assert.Contains(t, cmd.Lines[1], "hello")
	assert.Equal(t, cmd.Lines[2], "")
	assert.Contains(t, cmd.Lines[3], "finished with exit code 0")

	//// Stop the process
	//cmd.Stop()
	//time.Sleep(100 * time.Millisecond) // Allow some time for the process to stop
}
