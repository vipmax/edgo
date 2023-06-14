package main

import (
	"bufio"
	"fmt"
	"github.com/goccy/go-json"
	"io"
	"net/textproto"
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
	reader 		   *textproto.Reader
	responsesMap   map[int]string
	someMapMutex   sync.RWMutex
	isReady   	   bool
	id        	   int
	file2diagnostic map[string]DiagnosticParams
}

var langCommands = map[string][]string{
	"go":         {"gopls"},
	"python":     {"pyright-langserver", "--stdio"},
	"typescript": {"typescript-language-server", "--stdio"},
	"javascript": {"typescript-language-server", "--stdio"},
	"html": 	  {"vscode-html-language-server","--stdio"},
	"vue": 	  	  {"vls"},
	"rust": 	  {"rust-analyzer"},
	"c": 	  	  {"clangd"},
	"c++": 	  	  {"clangd"},
	"scala": 	  {"metals", "-Dmetals.ammoniteJvmProperties=metals.ammoniteJvmProperties=-Xmx4G"},
	"kotlin": 	  {"kotlin-language-server"},
	"java": 	  {"jdtls"},
}

func (this *LspClient) start(language string) bool {
	this.isReady = false
	this.someMapMutex = sync.RWMutex{}

	// Getting the lsp command with args for a language:
	lspCmd, ok := langCommands[strings.ToLower(language)]
	if !ok || len(lspCmd) == 0 { return false }  // lang is not supported.

	_, lsperr := exec.LookPath(lspCmd[0])
	if lsperr != nil { fmt.Println("lsp %s not found", lspCmd[0]); return false }

	this.process = exec.Command(lspCmd[0], lspCmd[1:]...)

	var stdin, err = this.process.StdinPipe()
	if err != nil { fmt.Println(err) }
	this.stdin = stdin

	stdout, err := this.process.StdoutPipe()
	if err != nil { fmt.Println(err); return false }
	this.stdout = stdout

	this.reader = textproto.NewReader(bufio.NewReader(stdout))

	// for debug, todo write logs to file
	//this.process.Stdout = os.Stdout
	//this.process.Stderr = os.Stderr

	err = this.process.Start()
	if err != nil { fmt.Println("An error occured: ", err) }

	this.responsesMap = make(map[int]string)
	this.file2diagnostic = make(map[string]DiagnosticParams)

	return true
}

func (this *LspClient) send(o interface{})  {
	m, err := json.Marshal(o)
	if err != nil { panic(err) }

	message := fmt.Sprintf("Content-Length: %d\r\n\r\n%s", len(m), m)
	//fmt.Println("[send]", message)
	_, err = this.stdin.Write([]byte(message))
	if err != nil { panic(err) }
}

func (this *LspClient) receive() string {

	headers, err := this.reader.ReadMIMEHeader()
	if err != nil { /*panic(err)*/ return "" }

	length, err := strconv.Atoi(headers.Get("Content-Length"))
	if err != nil { /*panic(err)*/ return "" }

	body := make([]byte, length)
	if _, err := this.reader.R.Read(body); err != nil { /*panic(err)*/ return "" }

	return string(body)
}


func (this *LspClient) receiveLoop(diagnosticUpdateChannel chan string) {
	go func() {
		for {
			//message := this.receive()
			_, message := this.read_stdout(false, false)

			if strings.Contains(message,"textDocument/publishDiagnostics") {
				var dr DiagnosticResponse
				errp := json.Unmarshal([]byte(message), &dr)
				if errp != nil { /*panic(errp)*/ continue }
				this.file2diagnostic[dr.Params.Uri] = dr.Params
				diagnosticUpdateChannel <- "update"
				continue
			}

			responseJSON := make(map[string]interface{})
			err := json.Unmarshal([]byte(message), &responseJSON)
			if err != nil {  /*fmt.Println("Error parsing JSON response:", err)*/ continue }

			if value, found := responseJSON["id"]; found {
				if id, ok := value.(float64); ok {
					this.someMapMutex.Lock()
					this.responsesMap[int(id)] = message
					this.someMapMutex.Unlock()
				}
			}
		}
	}()
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
	this.receive()

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
	if err != nil { fmt.Println("Error reading file:", err); return }

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

	//jsonString := this.receive()
	//
	//var dr DiagnosticResponse
	//errp := json.Unmarshal([]byte(jsonString), &dr)
	//if errp != nil { /*panic(errp)*/ return }
	//this.file2diagnostic[dr.Params.Uri] = dr.Params
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

	start := time.Now()
	var jsonData string
	for {
		if time.Since(start) >= time.Second {
			break
		}
		this.someMapMutex.Lock()
		value, ok := this.responsesMap[this.id]
		this.someMapMutex.Unlock()
		if ok {
			jsonData = value
			this.someMapMutex.Lock()
			delete(this.responsesMap, this.id)
			this.someMapMutex.Unlock()
			break
		}
	}

	//jsonData := this.receive()
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

	start := time.Now()
	var jsonData string
	for {
		if time.Since(start) >= time.Second {
			break
		}
		this.someMapMutex.Lock()
		value, ok := this.responsesMap[this.id]
		this.someMapMutex.Unlock()
		if ok {
			jsonData = value
			this.someMapMutex.Lock()
			delete(this.responsesMap, this.id)
			this.someMapMutex.Unlock()
			break
		}
	}

	//jsonData := this.receive()
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

func (this *LspClient) completion(file string, line int, character int) (CompletionResponse, error) {
	this.id++
	id := this.id

	request := BaseRequest{
		ID: id, JSONRPC: "2.0", Method:  "textDocument/completion",
		Params: Params{
			TextDocument: TextDocument { URI:  "file://" + file },
			Position: Position { Line: line, Character: character },
			Context: Context { TriggerKind: 1 },
		},
	}

	this.send(request)

	start := time.Now()
	var jsonData string
	for {
		if time.Since(start) >= time.Second {
			break
		}
		this.someMapMutex.Lock()
		value, ok := this.responsesMap[this.id]
		this.someMapMutex.Unlock()
		if ok {
			jsonData = value
			this.someMapMutex.Lock()
			delete(this.responsesMap, this.id)
			this.someMapMutex.Unlock()
			break
		}
	}

	//jsonData := this.receive()
	var completionResponse CompletionResponse
	err := json.Unmarshal([]byte(jsonData), &completionResponse)
	//if err != nil { panic("Error parsing JSON:" + err.Error()) }
	return completionResponse, err
}

func (this *LspClient) read_stdout(addtoMap bool, isDiagnostic bool) (map[string]interface{}, string) {
	//start := time.Now()

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
				panic(err)
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

			method, found := responseJSON["method"];
			if found && method.(string) == "textDocument/publishDiagnostics" {
				var dr DiagnosticResponse
				err := json.Unmarshal(buf, &dr)
				if err != nil { panic(err) }

				return responseJSON, line
			}

			if value, found := responseJSON["id"]; found {
				if _, ok := value.(float64); ok {
					return responseJSON, line
				}
			}

		} else {
			line, err = reader.ReadString('\n') // it stuck sometimes
			if err != nil {
				//panic(err)
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