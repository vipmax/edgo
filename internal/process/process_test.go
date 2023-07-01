package process

import (
	"bufio"
	"context"
	. "fmt"
	"os"
	"os/exec"
	"os/signal"
	"syscall"
	"testing"
	"time"
)

func TestProcess(t *testing.T) {

	process := NewProcess("/Users/max/opt/anaconda3/bin/python", "atest.py")
	process.Cmd.Env = append(os.Environ(), "PYTHONUNBUFFERED=1")

	process.Start()

	go func() {
		for line := range process.Out {
			Println("Output:", line)
		}
	}()

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
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)

	go run(ctx)

	time.Sleep(3 * time.Second)

	stop()

	time.Sleep(30 * time.Second)
}

func run(ctx context.Context) {
	cmd := exec.CommandContext(ctx, "python3", "atest.py")
	//cmd := exec.CommandContext(ctx, "go", "run", "cmd/test/main.go")
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