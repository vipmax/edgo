package dap

import (
	"bufio"
	"context"
	. "edgo/internal/logger"
	"fmt"
	"os"
	"os/exec"
	"os/signal"
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

	if dap.IsStopped { t.Errorf("Expected lsp not to be stopped") }
}




func TestDapClientInitialize(t *testing.T) {
	os.Chdir("../..")

	os.Setenv("EDGO_LOG", "edgo.log")
	Log.Start()

	dap := DapClient{Lang: "go", Conntype: "tcp", Port: 54752}
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

	dap := DapClient{Lang: "go", Conntype: "tcp", Port: 54752}
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

	dap := DapClient{Lang: "go", Conntype: "tcp", Port: 54752}
	args := "dap --listen=127.0.0.1:54752 --log=true --log-output=dap --log-dest=dlv.log"

	dap.Start("dlv", strings.Split(args, " ")...)
	dap.Init(currentDir)
	dap.Launch("./cmd/test/main.go")
	
	setBreakpointResult := dap.SetBreakpoint("/Users/max/apps/go/edgo/cmd/test/main.go", 9)
	
	if setBreakpointResult == false {
		t.Errorf("Expected setBreakpointResult to be true, got false")
	}

	setBreakpointResult2 := dap.SetBreakpoint("/Users/max/apps/go/edgo/cmd/test/main.go", 10)

	if setBreakpointResult2 == false {
		t.Errorf("Expected setBreakpointResult to be true, got false")
	}
}

func TestDapClientContinue(t *testing.T) {
	os.Chdir("../..")
	currentDir, _ := os.Getwd()
	
	os.Setenv("EDGO_LOG", "edgo.log")
	Log.Start()

	dap := DapClient{Lang: "go", Conntype: "tcp", Port: 54752}
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

	dap := DapClient{Lang: "go", Conntype: "tcp", Port: 54752}
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
func TestDapClientDisconnect(t *testing.T) {
	os.Chdir("../..")
	currentDir, _ := os.Getwd()

	os.Setenv("EDGO_LOG", "edgo.log")
	Log.Start()

	dap := DapClient{Lang: "go", Conntype: "tcp", Port: 54752}
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

	dap.Disconnect()

}

func TestDapClientContinueWithEventsRwice(t *testing.T) {
	os.Chdir("../..")
	currentDir, _ := os.Getwd()

	os.Setenv("EDGO_LOG", "edgo.log")
	Log.Start()

	dap := DapClient{Lang: "go", Conntype: "tcp", Port: 54752}
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

	dap.Stop()
	time.Sleep(time.Second)

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

	threadId = 1
	dap.Continue(threadId)
	time.Sleep(time.Second)
	dap.Continue(threadId)
	time.Sleep(time.Second)

}

func TestDp(t *testing.T) {

	args := "/Users/max/.pyenv/versions/3.11.2/bin/python -m debugpy --listen localhost:54752 --log-to-stderr --wait-for-client atest.py"
	split := strings.Split(args, " ")

	ctx, _ := signal.NotifyContext(context.Background(), os.Kill)
	cmd := exec.CommandContext(ctx, split[0], split[1:]...)
	cmd.Env = append(os.Environ())


	go func() {
		err := cmd.Run()
		if err != nil {
			fmt.Println("start err", err.Error())
		}
	}()

	go func() {
		stdout, _ := cmd.StderrPipe()
		scanner := bufio.NewScanner(stdout)
		for scanner.Scan() {
			message := scanner.Text()
			fmt.Println("message", message)
		}
	}()

	time.Sleep(time.Second*10)

}

func TestDapClientStartPython(t *testing.T) {
	os.Chdir("../..")
	currentDir, _ := os.Getwd()

	os.Setenv("EDGO_LOG", "edgo.log")
	os.Setenv("PYTHONUNBUFFERED", "1")

	Log.Start()

	dap := DapClient{Lang: "python", Conntype: "tcp", Port: 54752}
	args := "-m debugpy --listen localhost:54752 --log-to dappy.log --wait-for-client atest.py"
	cmd := "python3"
	dap.Start(cmd, strings.Split(args, " ")...)

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

	//if !started { t.Errorf("Error, dap not started") }
	//if dap.cmd == nil { t.Errorf("Error, cmd is nil") }
	//
	//pid := dap.cmd.Process.Pid
	//fmt.Println("dap pid is", pid)
	//
	//process, err := os.FindProcess(pid)
	//if err != nil { t.Errorf("Error finding cmd with id %d: %s\n", process.Pid, err) }
	//
	//if dap.IsStopped { t.Errorf("Expected lsp not to be stopped") }

	initResult := dap.Init(currentDir)
	fmt.Println("initResult", initResult)
	attachResult := dap.Attach(false)
	fmt.Println("attachResult", attachResult)
	configurationDoneResult := dap.ConfigurationDone()
	fmt.Println("configurationDoneResult", configurationDoneResult)

	dap.SetBreakpoint("/Users/max/apps/go/edgo/atest.py", 6)
	time.Sleep(time.Second)

	dap.Continue(1)
	time.Sleep(time.Second)
	dap.Continue(1)
	time.Sleep(time.Second)
	dap.Continue(1)
	time.Sleep(time.Second)

	stackTraceResponse := dap.Stacktrace(1, 1)
	fmt.Println("stackTraceResponse", stackTraceResponse)

	frameid := stackTraceResponse.ResponseBody.StackFrames[0].ID

	scopesResponse := dap.Scopes(frameid)
	fmt.Println("scopesResponse", scopesResponse)
	scope := scopesResponse.ResponseBody.Scopes[0]

	variablesResponse := dap.Variables(scope.VariablesReference)
	fmt.Println("variablesResponse", variablesResponse)

	time.Sleep(time.Second*100000)

}