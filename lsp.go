package main

import (
	"bufio"
	"fmt"
	"github.com/goccy/go-json"
	"io"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"sync"
	"time"
)

// LspClient represents a client for communicating with a Language Server Protocol (LSP) server.
type LspClient struct {
	process        *exec.Cmd      // The underlying process running the LSP server.
	stdin          io.WriteCloser // The standard input pipe for sending data to the LSP server.
	stdout         io.ReadCloser
	responses      map[int]map[string]interface{}
	responsesMutex sync.Mutex
	isReady   	   bool
	id        	   int
	lspCmd 	       []string
}

var langCommands = map[string][]string{
	"go":         {"gopls"},
	"python":     {"pylsp"},
	"typescript": {"typescript-language-server", "--stdio"},
	"javascript": {"typescript-language-server", "--stdio"},
	"html": 	  {"vscode-html-language-server","--stdio"},
	"vue": 	  	  {"vls"},
	"rust": 	  {"rust-analyzer"},
	"c": 	  	  {"clangd"},
	"c++": 	  	  {"clangd"},
	"scala": 	  {"metals", "-Xms1G -Xmx4G -Dmetals.ammoniteJvmProperties=metals.ammoniteJvmProperties=-Xmx4G"},
	"kotlin": 	  {"kotlin-language-server"},
	"java": 	  {"jdtls"},
}

func (this *LspClient) start(language string) bool {
	this.isReady = false

	// Getting the lsp command with args for a language:
	lspCmd, ok := langCommands[strings.ToLower(language)]
	if !ok || len(lspCmd) == 0 { return false }  // lang is not supported.

	this.lspCmd = lspCmd
	this.process = exec.Command( this.lspCmd[0], this.lspCmd[1:]...)

	var stdin, err = this.process.StdinPipe()
	if err != nil { fmt.Println(err) }
	this.stdin = stdin

	stdout, err := this.process.StdoutPipe()
	if err != nil { fmt.Println(err); return false }
	this.stdout = stdout

	// for debug, todo write logs to file
	//this.process.Stdout = os.Stdout
	//this.process.Stderr = os.Stderr

	err = this.process.Start()
	if err != nil { fmt.Println("An error occured: ", err) }

	this.responses = make(map[int]map[string]interface{})

	return true
}

func (this *LspClient) send(o interface{}) error {
	m, err := json.Marshal(o)
	if err != nil { return fmt.Errorf("error encoding JSON: %v", err) }

	message := fmt.Sprintf("Content-Length: %d\r\n\r\n%s", len(m), m)
	//fmt.Println("[send]", message)
	_, err = this.stdin.Write([]byte(message))
	if err != nil { return fmt.Errorf("error sending message: %v", err) }

	return nil
}

func (this *LspClient) init(dir string) {
	this.id = 0
	initializeRequest := InitializeRequest{
		ID: this.id, JSONRPC: "2.0",
		Method: "initialize",
		Params: InitializeParams{
			RootURI: "file://" + dir, RootPath: dir,
			WorkspaceFolders: []WorkspaceFolder{{ Name: "edgo", URI:  "file://" + dir }},
			Capabilities: capabilities,
			ClientInfo: ClientInfo{ Name: "edgo",Version: "1.0.0"},
		},
	}

	this.send(initializeRequest)
	time.Sleep(time.Millisecond * 30)
	go this.read_stdout(true)
	time.Sleep(time.Millisecond * 30)
	this.waitForResponseInMap(this.id)

	initializedRequest := InitializedRequest{
		JSONRPC: "2.0",
		Method:  "initialized",
		Params:  struct{}{},
	}
	this.send(initializedRequest)
	this.isReady = true
}

func (this *LspClient) didOpen(file string) {
	filecontent, err := os.ReadFile(file)
	if err != nil {
		fmt.Println("Error reading file:", err)
		return
	}

	didOpenRequest := DidOpenRequest{
		JSONRPC: "2.0",
		Method:  "textDocument/didOpen",
		Params: DidOpenParams{
			TextDocument: TextDocument{
				LanguageID: "go",
				Text:       string(filecontent),
				URI:        "file://" + file,
				Version:    1,
			},
		},
	}
	this.send(didOpenRequest)
}
func (this *LspClient) didChange(file string) {
	filecontent, _ := os.ReadFile(file)
	text := string(filecontent)

	didOpenRequest := TextDocumentDidChangeParams{
		TextDocument: VersionedTextDocumentIdentifier{
			URI:     "file://" + file,
			Version: 1,
		},
		ContentChanges: []TextDocumentContentChangeEvent{
			{
				Text: text,
			},
		},
	}

	this.send(didOpenRequest)
	time.Sleep(time.Millisecond * 10)
}
func (this *LspClient) didSave(file string) {
	request := BaseRequest{
		JSONRPC: "2.0",
		Method:  "textDocument/didSave",
		Params: Params{
			TextDocument: TextDocument{
				URI: "file://" + file,
			},
		},
	}

	this.send(request)
	time.Sleep(time.Millisecond * 10)
}

func (this *LspClient) references(file string, line int, character int) {
	this.id++
	referencesRequest := BaseRequest{
		ID:      this.id,
		JSONRPC: "2.0",
		Method:  "textDocument/references",
		Params: Params{
			TextDocument: TextDocument{
				URI: "file://" + file,
			},
			Position: Position{
				Line:      line,
				Character: character,
			},
		},
	}

	this.send(referencesRequest)
	time.Sleep(time.Millisecond * 10)
}
func (this *LspClient) hover(file string, line int, character int) (HoverResponse, error) {
	this.id++

	request := BaseRequest{
		ID: this.id, JSONRPC: "2.0", Method:  "textDocument/hover",
		Params: Params {
			TextDocument: TextDocument { URI: "file://" + file },
			Position: Position { Line: line, Character: character },
		},
	}

	this.send(request)
	_, jsonData := this.read_stdout(false)

	var response HoverResponse
	err := json.Unmarshal([]byte(jsonData), &response)
	//if err != nil { panic("Error parsing JSON:" + err.Error()) }
	return response, err
}
func (this *LspClient) signatureHelp(file string, line int, character int) (SignatureHelpResponse, error) {
	this.id++

	request := BaseRequest{
		ID: this.id, JSONRPC: "2.0", Method:  "textDocument/signatureHelp",
		Params: Params {
			TextDocument: TextDocument { URI: "file://" + file },
			Position: Position { Line: line, Character: character },
		},
	}

	this.send(request)
	_, jsonData := this.read_stdout(false)

	var response SignatureHelpResponse
	err := json.Unmarshal([]byte(jsonData), &response)
	//if err != nil { panic("Error parsing JSON:" + err.Error()) }
	return response, err
}
func (this *LspClient) definition(file string, line int, character int) {
	this.id++
	request := BaseRequest{
		ID:      this.id,
		JSONRPC: "2.0",
		Method:  "textDocument/definition",
		Params: Params{
			TextDocument: TextDocument{
				URI: "file://" + file,
			},
			Position: Position{
				Line:      line,
				Character: character,
			},
		},
	}

	this.send(request)
	time.Sleep(time.Millisecond * 1000)
}

func (this *LspClient) completion(file string, code string, line int, character int) (CompletionResponse, map[string]interface{}) {
	this.id++
	id := this.id

	request := BaseRequest{
		ID: id, JSONRPC: "2.0", Method:  "textDocument/completion",
		Params: Params{
			TextDocument: TextDocument { URI:  "file://" + file, Text: code },
			Position: Position { Line: line, Character: character },
			Context: Context { TriggerKind: 1 },
		},
	}

	this.send(request)
	js, jsonData := this.read_stdout(false)
	var completionResponse CompletionResponse
	err := json.Unmarshal([]byte(jsonData), &completionResponse)
	if err != nil { panic("Error parsing JSON:" + err.Error()) }
	return completionResponse, js
}

func (this *LspClient) waitForResponseInMap(id int) map[string]interface{} {
	/// wait for response for some id

	for {
		//this.responsesMutex.Lock()
		value, ok := this.responses[id]
		//this.responsesMutex.Lock()
		if ok {
			//this.responsesMutex.Lock()
			delete(this.responses, id)
			//this.responsesMutex.Lock()
			return value
		}
	}
}

func (this *LspClient) read_stdout(addtoMap bool) (map[string]interface{}, string) {
	const LEN_HEADER = "Content-Length: "

	reader := bufio.NewReader(this.stdout)
	var messageSize int
	var responseMustBeNext bool

	for {
		var line string
		var err error

		if messageSize != 0 && responseMustBeNext {
			buf := make([]byte, messageSize)
			_, err = io.ReadFull(reader, buf)
			if err != nil {
				//fmt.Println("Error reading from stdout:", err)
				continue
			}
			line = string(buf)
			messageSize = 0
			//fmt.Println("response", line)

			responseJSON := make(map[string]interface{})
			err = json.Unmarshal(buf, &responseJSON)
			if err != nil {
				//fmt.Println("Error parsing JSON response:", err)
				continue
			}

			if value, found := responseJSON["id"]; found {
				if id, ok := value.(float64); ok {
					if addtoMap {
						this.responsesMutex.Lock()
						this.responses[int(id)] = responseJSON
						this.responsesMutex.Unlock()
					}
					return responseJSON, line
				}
			}

		} else {
			line, err = reader.ReadString('\n')
			if err != nil {
				continue
			}
			//fmt.Println("line", line)
		}

		line = strings.TrimSuffix(line, "\r\n")

		if strings.HasPrefix(line, LEN_HEADER) {
			sizeStr := strings.TrimPrefix(line, LEN_HEADER)
			msize, _ := strconv.Atoi(sizeStr)
			messageSize = msize
			responseMustBeNext = false
			continue
		}

		if line == "" {
			responseMustBeNext = true
			continue
		}
	}
}
