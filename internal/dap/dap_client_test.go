package dap

import (
	. "edgo/internal/logger"
	"fmt"
	"os"
	"strings"
	"time"

	// "path"
	// "strings"
	"testing"
)


func TestDapClientStart(t *testing.T) {
	os.Chdir("../..")

	dap := DapClient{Lang: "go"}
	started := dap.Start("dlv", "dap")
	defer dap.Stop()

	if !started { t.Errorf("Error, dap not started") }

	if dap.cmd == nil { t.Errorf("Error, cmd is nil") }

	pid := dap.cmd.Process.Pid
	fmt.Println("dap pid is", pid)

	process, err := os.FindProcess(pid)
	if err != nil { t.Errorf("Error finding cmd with id %d: %s\n", process.Pid, err) }

	if dap.isStopped { t.Errorf("Expected lsp not to be stopped") }
}

func TestDapClientInitialize(t *testing.T) {
	os.Chdir("../..")

	os.Setenv("EDGO_LOG", "edgo.log")
	Log.Start()

	dap := DapClient{Lang: "go", conntype: "tcp", port: 54752}
	args := "dap --listen=127.0.0.1:54752 --log=true --log-output=dap --log-dest=dlv.log"

	dap.Start("dlv", strings.Split(args, " ")...)

	pid := dap.cmd.Process.Pid
	fmt.Println("dap pid is", pid)

	currentDir, _ := os.Getwd()
	dap.Init(currentDir)

	if dap.IsReady == false {
		t.Errorf("Expected lsp to be ready, got false")
	}
}

func TestDapClientLaunch(t *testing.T) {
	os.Chdir("../..")
	currentDir, _ := os.Getwd()

	os.Setenv("EDGO_LOG", "edgo.log")
	Log.Start()

	dap := DapClient{Lang: "go", conntype: "tcp", port: 54752}
	args := "dap --listen=127.0.0.1:54752 --log=true --log-output=dap --log-dest=dlv.log"

	dap.Start("dlv", strings.Split(args, " ")...)
	dap.Init(currentDir)

	launchResult := dap.Launch("./cmd/test/main.go")
	
	if launchResult == false {
		t.Errorf("Expected launchResult to be true, got false")
	}
}

func TestDapClientBreakpoint(t *testing.T) {
	os.Chdir("../..")
	currentDir, _ := os.Getwd()
	
	os.Setenv("EDGO_LOG", "edgo.log")
	Log.Start()

	dap := DapClient{Lang: "go", conntype: "tcp", port: 54752}
	args := "dap --listen=127.0.0.1:54752 --log=true --log-output=dap --log-dest=dlv.log"

	dap.Start("dlv", strings.Split(args, " ")...)
	dap.Init(currentDir)
	dap.Launch("./cmd/test/main.go")
	
	setBreakpointResult := dap.SetBreakpoint("/Users/max/apps/go/edgo/cmd/test/main.go", 9)
	
	if setBreakpointResult == false {
		t.Errorf("Expected setBreakpointResult to be true, got false")
	}
}

func TestDapClientContinue(t *testing.T) {
	os.Chdir("../..")
	currentDir, _ := os.Getwd()
	
	os.Setenv("EDGO_LOG", "edgo.log")
	Log.Start()

	dap := DapClient{Lang: "go", conntype: "tcp", port: 54752}
	args := "dap --listen=127.0.0.1:54752 --log=true --log-output=dap --log-dest=dlv.log"

	dap.Start("dlv", strings.Split(args, " ")...)
	dap.Init(currentDir)
	dap.Launch("./cmd/test/main.go")
	dap.SetBreakpoint("/Users/max/apps/go/edgo/cmd/test/main.go", 9)
	
	threadId := 1
	continueResult := dap.Continue(threadId)
	
	if continueResult == false {
		t.Errorf("Expected setBreakpointResult to be true, got false")
	}
}

func TestDapClientContinueWithEvents(t *testing.T) {
	os.Chdir("../..")
	currentDir, _ := os.Getwd()

	os.Setenv("EDGO_LOG", "edgo.log")
	Log.Start()

	dap := DapClient{Lang: "go", conntype: "tcp", port: 54752}
	args := "dap --listen=127.0.0.1:54752 --log=true --log-output=dap --log-dest=dlv.log"

	dap.Start("dlv", strings.Split(args, " ")...)

	go func() {
		for eventMessage := range dap.EventMessages {
			fmt.Println("eventMessage", eventMessage)
		}
	}()

	go func() {
		for stdoutMessage := range dap.StdoutMessages {
			fmt.Println("stdoutMessage", stdoutMessage)
		}
	}()

	dap.Init(currentDir)
	time.Sleep(time.Second)
	dap.Launch("./cmd/test/main.go")
	time.Sleep(time.Second)
	dap.SetBreakpoint("/Users/max/apps/go/edgo/cmd/test/main.go", 9)

	threadId := 1
	dap.Continue(threadId)
	time.Sleep(time.Second)
	dap.Continue(threadId)
	time.Sleep(time.Second)

}
