package lsp

import (
	. "edgo/internal/logger"
	"fmt"
	"os"
	"path"
	"strings"
	"testing"
)

func TestLspClientStart(t *testing.T) {
	lsp := LspClient{}
	started := lsp.Start("gopls")
	defer lsp.Stop()

	if !started {
		t.Errorf("Error, lsp not started")
		return
	}

	if lsp.cmd == nil {
		t.Errorf("Error, cmd is nil")
		return
	}

	pid := lsp.cmd.Process.Pid
	fmt.Println("lsp pid is", pid)

	process, err := os.FindProcess(pid)
	if err != nil {
		t.Errorf("Error finding cmd with id %d: %s\n", process.Pid, err)
		return
	}

	if lsp.isStopped {
		t.Errorf("Expected lsp not to be stopped")
		return
	}
}

func TestLspClientStop(t *testing.T) {
	lsp := LspClient{Lang: "go"}
	lsp.Start("gopls")

	pid := lsp.cmd.Process.Pid
	fmt.Println("lsp pid is", pid)

	//time.Sleep(10*time.Second)
	lsp.Stop()
	//time.Sleep(10*time.Second)

	if lsp.isStopped == false {
		t.Errorf("Expected lsp to be stopped, got false")
	}
}

func TestLspClientInitialize(t *testing.T) {
	lsp := LspClient{Lang: "go"}
	lsp.Start("gopls")

	pid := lsp.cmd.Process.Pid
	fmt.Println("lsp pid is", pid)

	currentDir, _ := os.Getwd()
	lsp.Init(currentDir)

	if lsp.IsReady == false {
		t.Errorf("Expected lsp to be ready, got false")
	}
}

func TestLspClientHover(t *testing.T) {
	err := os.Setenv("EDGO_LOG", "edgo.log")
	Log.Start()

	lsp := LspClient{Lang: "go"}
	lsp.Start("gopls")

	currentDir, _ := os.Getwd()
	lsp.Init(currentDir)

	file := path.Join(currentDir, "lsp_client_test.go")
	text, _ := os.ReadFile(file)
	lsp.DidOpen(file, string(text))

	response, err := lsp.Hover(file, 77-1, 7)

	fmt.Println(response, err)

	expected := "func (*Logger).Start()"
	got := response.Result.Contents.Value
	if got != expected {
		t.Errorf("Expected lsp hover result to be %s, got something else %s", expected, got)
	}
}

func TestLspClientCompletion(t *testing.T) {
	err := os.Setenv("EDGO_LOG", "edgo.log")
	Log.Start()

	lsp := LspClient{Lang: "go"}
	lsp.Start("gopls")

	currentDir, _ := os.Getwd()
	lsp.Init(currentDir)

	file := path.Join(currentDir, "lsp_client_test.go")
	text, _ := os.ReadFile(file)
	lsp.DidOpen(file, string(text))

	response, err := lsp.Completion(file, 100-1, 8)

	fmt.Println(response, err)

	expected := "Start"
	if len(response.Result.Items) != 1 && response.Result.Items[0].Label != expected {
		t.Errorf("Expected lsp Completion result to be %s, got something else", expected)
	}
}

func TestLspClientDefinition(t *testing.T) {
	err := os.Setenv("EDGO_LOG", "edgo.log")
	Log.Start()

	lsp := LspClient{Lang: "go"}
	lsp.Start("gopls")

	currentDir, _ := os.Getwd()
	lsp.Init(currentDir)

	file := path.Join(currentDir, "lsp_client_test.go")
	text, _ := os.ReadFile(file)
	lsp.DidOpen(file, string(text))

	response, err := lsp.Definition(file, 124-1, 8)

	fmt.Println(response, err)

	expectedSuffix := "logger/logger.go"
	if len(response.Result) == 1 && !strings.Contains(response.Result[0].URI, expectedSuffix) {
		t.Errorf("Expected lsp Definition result to be %s, got something else", expectedSuffix)
	}
}

func TestLspClientSignatureHelp(t *testing.T) {
	err := os.Setenv("EDGO_LOG", "edgo.log")
	Log.Start()

	lsp := LspClient{Lang: "go"}
	lsp.Start("gopls")

	currentDir, _ := os.Getwd()
	lsp.Init(currentDir)

	file := path.Join(currentDir, "lsp_client_test.go")
	text, _ := os.ReadFile(file)
	lsp.DidOpen(file, string(text))

	response, err := lsp.SignatureHelp(file, 156-1, 21)

	fmt.Println(response, err)

	expected := "Join"
	if len(response.Result.Signatures) != 1 && !strings.Contains(response.Result.Signatures[0].Label, expected) {
		t.Errorf("Expected lsp Definition result to be %s, got something else", expected)
	}
}

func TestLspClientReferences(t *testing.T) {
	err := os.Setenv("EDGO_LOG", "edgo.log")
	Log.Start()

	lsp := LspClient{Lang: "go"}
	lsp.Start("gopls")

	currentDir, _ := os.Getwd()
	lsp.Init(currentDir)

	file := path.Join(currentDir, "internal","lsp", "lsp_client_test.go")
	text, _ := os.ReadFile(file)
	s := string(text)
	lsp.DidOpen(currentDir, s)

	response, err := lsp.References(file, 174-1, 2)

	fmt.Println(response, err)
	// todo fix, something wrong with base dir
}
