package process

import (
	"bufio"
	"context"
	"fmt"
	"github.com/creack/pty"
	"github.com/stretchr/testify/assert"
	"io"
	"os"
	"os/exec"
	"os/signal"
	"testing"
	"time"
)

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

func TestWriteStdin(t *testing.T) {
	cmd := NewProcess("cat")
	assert.NotNil(t, cmd)

	cmd.Start() // Start the process

	time.Sleep(10 * time.Millisecond) // Allow some time for the process to start


	input := "Hello, stdin!"
	go cmd.WriteStdin(input)

	for range cmd.Updates { } // wait for no updates anymore

	lines := cmd.GetLines(0)
	fmt.Println(lines)

	expectedOutput := "Hello, stdin!"
	if len(lines) == 0 || lines[0] != expectedOutput {
		t.Errorf("Expected process to receive input '%s', but got %v", expectedOutput, lines)
	}
}


func TestWriteStdinBash(t *testing.T) {
	cmd := NewProcess("bash")
	assert.NotNil(t, cmd)

	cmd.Start() // Start the process

	time.Sleep(10 * time.Millisecond) // Allow some time for the process to start


	input := "ls"
	cmd.WriteStdin(input)

	for range cmd.Updates { } // wait for no updates anymore

	cmd.WriteStdin(input)

	lines := cmd.GetLines(0)
	fmt.Println(lines)


}

func TestPty(t *testing.T) {
	c := exec.Command("cat")
	f, err := pty.Start(c)
	if err != nil {
		panic(err)
	}

	go func() {
		f.Write([]byte("foo\n"))
		f.Write([]byte("bar\n"))
		f.Write([]byte("baz\n"))
		//f.Write([]byte{4}) // EOT
	}()

	io.Copy(os.Stdout, f)
}

func TestPtycmdStop(t *testing.T) {
	shell := os.Getenv("SHELL")
	//shell := "bash"

	ctx, cancel := context.WithTimeout(context.Background(), 2 * time.Second)
	c := exec.CommandContext(ctx, shell, "-il")

	//c := exec.Command(shell)
	f, err := pty.Start(c)

	if err != nil { t.Fatal(err) }

	go func() {
		err := c.Wait()
		if err != nil { t.Fatal(err) }
		fmt.Println("wait done ")
	}()


	go func() {
		fmt.Println("pwd")
		f.Write([]byte("pwd\n"))
		time.Sleep(1000 * time.Millisecond)

		fmt.Println("ll")
		f.Write([]byte("ll\n"))
		time.Sleep(1000 * time.Millisecond)

		//f.Write([]byte{4}) // EOT
	}()


	go io.Copy(os.Stdout, f)
	// Create a reader to manually read from the pty
	//reader := bufio.NewReader(f)

	// Read and process the output line by line
	//for {
	//	line, err := reader.ReadString('\n')
	//	if err != nil {
	//		if err == io.EOF { break }
	//		t.Fatal(err)
	//	}
	//
	//	line = strings.ReplaceAll(line,"\r\n","")
	//	line = stripansi.Strip(line)
	//	fmt.Println(line)
	//}

	//f.WriteString("exit\r\n")
	time.Sleep(10000 * time.Millisecond)
	cancel()
	//c.Cancel()
	//c.Process.Kill()
	//c.Process.Release()

	err = f.Close()
	if err != nil { t.Fatal(err) }

	fmt.Println("done")
}