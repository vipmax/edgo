package main

import (
	"bufio"
	"fmt"
	"github.com/goccy/go-json"
	"io"
	"net/textproto"
	"os"
	"os/exec"
	"path"
	"strconv"
	"testing"
	"time"
)

func TestGopls(t *testing.T) {
	dir, _ := os.Getwd()
	file := dir + "/lsp_test.go"

	cmd := exec.Command("gopls")

	stdin, _ := cmd.StdinPipe()
	stdout, _ := cmd.StdoutPipe()

	reader := textproto.NewReader(bufio.NewReader(stdout))

	err := cmd.Start()
	if err != nil { fmt.Println(err); return }

	initializeRequest := InitializeRequest{
		ID: 0, JSONRPC: "2.0",
		Method: "initialize",
		Params: InitializeParams{
			RootURI: "file://" + dir, RootPath: dir,
			WorkspaceFolders: []WorkspaceFolder{{ Name: "edgo", URI:	"file://" + dir }},
			Capabilities: capabilities,
			ClientInfo: ClientInfo{ Name: "edgo",Version: "1.0.0"},
		},
	}

	send(stdin, initializeRequest)

	m := receive(reader)
	fmt.Println("<-", m)

	initializedRequest := InitializedRequest{
		JSONRPC: "2.0", Method:	"initialized", Params:	struct{}{},
	}
	send(stdin, initializedRequest)

	go func() {
		filecontent, _ := os.ReadFile(file)

		didOpenRequest := DidOpenRequest{
			JSONRPC: "2.0", Method:	"textDocument/didOpen",
			Params: DidOpenParams{
				TextDocument: TextDocument{
					LanguageID: "go", Text: string(filecontent),
					URI: "file://" + file, Version:		1,
				},
			},
		}

		time.Sleep(2*time.Second)
		send(stdin, didOpenRequest)
	}()

	go func() {
		request := BaseRequest{
			ID: 1, JSONRPC: "2.0", Method:	"textDocument/hover",
			Params: Params{
				TextDocument: TextDocument { URI:	"file://" + file },
				Position: Position { Line: 77 - 1, Character: 11 },
			},
		}

		time.Sleep(4*time.Second)
		send(stdin, request)
	}()

	messagesChan := make(chan string)

	// writing messages to channel async
	go func() {
		for {
			message := receive(reader)
			messagesChan <- message
		}
	}()

	msg := <- messagesChan
	fmt.Println("<-", msg)

	// reading messages from channel
	for message := range messagesChan {
		fmt.Println("<-", message)
	}

}

func send(stdin io.WriteCloser, o interface{})	{
	m, err := json.Marshal(o)
	if err != nil { panic(err) }

	message := fmt.Sprintf("Content-Length: %d\r\n\r\n%s", len(m), m)
	fmt.Println("->", string(m))
	_, err = stdin.Write([]byte(message))
	if err != nil { panic(err) }
}

func receive(reader *textproto.Reader) string {
	headers, err := reader.ReadMIMEHeader()
	if err != nil { fmt.Println(err); return "" }

	length, err := strconv.Atoi(headers.Get("Content-Length"))
	if err != nil { fmt.Println(err); return ""}

	body := make([]byte, length)
	if _, err := reader.R.Read(body); err != nil { fmt.Println(err); return "" }

	return string(body)
}


func TestGoLangCompletion(t *testing.T) {
	dir, _ := os.Getwd()
	file := dir + "/lsp_test.go"

	fmt.Println("starting lsp server")

	lsp := LspClient{}
	lsp.start("go")
	lsp.init(dir)
	lsp.didOpen(file)

	completion, _ := lsp.completion(file, 18-1, 8)
	fmt.Println("completion", completion)

	var options []string
	items := completion.Result.Items
	for _, item := range items {
		options = append(options, item.Label)
	}

	fmt.Println("options", options)
	fmt.Println("ending lsp server")
}

func TestPythonCompletion(t *testing.T) {
	dir := "/Users/max/apps/python/editor/src/"
	file := path.Join(dir, "logger.py")

	fmt.Println("starting lsp server")

	lsp := LspClient{}
	lsp.start("python")
	lsp.init(dir)
	lsp.didOpen(file)

	completion, _ := lsp.completion(file, 8-1, 20)
	fmt.Println("completion", completion)

	var options []string
	items := completion.Result.Items
	for _, item := range items {
		options = append(options, item.Label)
	}

	fmt.Println("options", options)
	fmt.Println("ending lsp server")
}

func TestTypescriptCompletion(t *testing.T) {
	dir := "/Users/max/apps/ts/lsp-examples/"
	file := path.Join(dir, "lsp-test-ts.ts")

	fmt.Println("starting lsp server for ", file)

	lsp := LspClient{}
	lsp.start("typescript")
	lsp.init(dir)
	lsp.didOpen(file)

	completion, _ := lsp.completion(file, 31-1, 5)
	fmt.Println("completion", completion)

	var options []string
	items := completion.Result.Items
	for _, item := range items {
		options = append(options, item.Label)
	}

	fmt.Println("options", options)
	fmt.Println("ending lsp server")
}

func TestRustCompletion(t *testing.T) {
	dir := "/Users/max/apps/rust/lsp-examples/"
	file := path.Join(dir, "lsp-test-ts.ts")

	fmt.Println("starting lsp server for ", file)

	lsp := LspClient{}
	lsp.start("typescript")
	lsp.init(dir)
	lsp.didOpen(file)

	completion, _ := lsp.completion(file,	31-1, 5)
	fmt.Println("completion", completion)

	var options []string
	items := completion.Result.Items
	for _, item := range items {
		options = append(options, item.Label)
	}

	fmt.Println("options", options)
	fmt.Println("ending lsp server")
}

func TestScalaCompletion(t *testing.T) {
	dir := "/Users/max/apps/scala/chrome4s"
	file := path.Join(dir, "/src/main/scala/chrome4s/Main.scala")

	fmt.Println("starting lsp server for ", file)

	lsp := LspClient{}
	lsp.start("scala")
	lsp.init(dir)
	lsp.didOpen(file)
	time.Sleep(3*time.Second)
	completion, _ := lsp.completion(file, 17-1, 8)
	fmt.Println("completion", completion)

	var options []string
	items := completion.Result.Items
	for _, item := range items {
		options = append(options, item.Label)
	}

	fmt.Println("options", options)
	fmt.Println("ending lsp server")
}


func TestGoLangHover(t *testing.T) {
	dir, _ := os.Getwd()
	file := dir + "/lsp_test.go"

	fmt.Println("starting lsp server")

	lsp := LspClient{}
	lsp.start("go")
	lsp.init(dir)
	lsp.didOpen(file)

	hover, _ := lsp.hover(file,18-1, 13)
	fmt.Println("hover range: ", hover.Result.Range)
	fmt.Println("hover content:\n", hover.Result.Contents.Value)

	fmt.Println("ending lsp server")
}


func TestGoLangSignatureHelp(t *testing.T) {
	dir, _ := os.Getwd()
	file := dir + "/lsp_test.go"

	fmt.Println("starting lsp server")

	lsp := LspClient{}
	lsp.start("go")
	lsp.init(dir)
	lsp.didOpen(file)

	response, _ := lsp.signatureHelp(file,14-1, 36)
	fmt.Println("signatureHelp: ", response)

	fmt.Println("ending lsp server")
}
