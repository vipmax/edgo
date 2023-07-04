package lsp

import (
	"bufio"
	"fmt"
	"github.com/goccy/go-json"
	"io"
	"net/textproto"
	"os"
	"os/exec"
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
		time.Sleep(4*time.Second)

		err := cmd.Process.Signal(os.Kill )
		if err != nil {
			return
		}
		err2 := cmd.Process.Release()
		if err2 != nil {
			return
		}
		//request := BaseRequest{
		//	ID: 1, JSONRPC: "2.0", Method:	"textDocument/hover",
		//	Params: Params{
		//		TextDocument: TextDocument { URI:	"file://" + file },
		//		Position: Position { Line: 77 - 1, Character: 11 },
		//	},
		//}
		//
		//time.Sleep(4*time.Second)
		//send(stdin, request)
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

func TestSelect(t *testing.T) {
	ch := make(chan int)

	go func() {
		time.Sleep(10 * time.Second)
		ch <- 42
	}()

	select {
	case i := <-ch:
		// You can add your if condition here.
		if i > 40 {
			fmt.Println("Received:", i, "which is greater than 40.")
		} else {
			fmt.Println("Received:", i, "which is not greater than 40.")
		}
	case <-time.After(1 * time.Second):
		fmt.Println("Timeout.")
	}

	fmt.Println("Done")
}