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
	logger Logger
}

func (this *LspClient) start(language string, lspCmd []string) bool {
	this.isReady = false
	this.someMapMutex = sync.RWMutex{}

	_, lsperr := exec.LookPath(lspCmd[0])
	if lsperr != nil { this.logger.info("lsp not found:", lspCmd[0]); return false }

	this.process = exec.Command(lspCmd[0], lspCmd[1:]...)

	var stdin, err = this.process.StdinPipe()
	if err != nil { this.logger.info(err.Error()); return false  }
	this.stdin = stdin

	stdout, err := this.process.StdoutPipe()
	if err != nil { this.logger.info(err.Error()); return false }
	this.stdout = stdout

	this.reader = textproto.NewReader(bufio.NewReader(stdout))

	err = this.process.Start()
	if err != nil { this.logger.info(err.Error()); return false  }

	this.responsesMap = make(map[int]string)
	this.file2diagnostic = make(map[string]DiagnosticParams)

	return true
}

func (this *LspClient) send(o interface{})  {
	m, err := json.Marshal(o)
	if err != nil { panic(err) }
	this.logger.info("->", string(m))

	message := fmt.Sprintf("Content-Length: %d\r\n\r\n%s", len(m), m)
	_, err = this.stdin.Write([]byte(message))
	if err != nil {
		this.logger.error(err.Error())
	}
}

func (this *LspClient) receive() string {

	headers, err := this.reader.ReadMIMEHeader()
	if err != nil { /*panic(err)*/ return "" }

	length, err := strconv.Atoi(headers.Get("Content-Length"))
	if err != nil { /*panic(err)*/ return "" }

	body := make([]byte, length)
	if _, err := this.reader.R.Read(body); err != nil { /*panic(err)*/ return "" }

	message := string(body)
	this.logger.info("<-", message)

	return message
}


func (this *LspClient) receiveLoop(diagnosticUpdateChannel chan string) {
	for {
		//message := this.receive()
		_, message := this.read_stdout()
		this.logger.info("<-", message)

		if strings.Contains(message,"publishDiagnostics") {
			var dr DiagnosticResponse
			errp := json.Unmarshal([]byte(message), &dr)
			if errp != nil { this.logger.error(errp.Error()); continue }
			this.file2diagnostic[dr.Params.Uri] = dr.Params
			diagnosticUpdateChannel <- "update"
			continue
		}

		responseJSON := make(map[string]interface{})
		err := json.Unmarshal([]byte(message), &responseJSON)
		if err != nil { this.logger.error(err.Error()); continue }

		if value, found := responseJSON["id"]; found {
			if id, ok := value.(float64); ok {
				this.someMapMutex.Lock()
				this.responsesMap[int(id)] = message
				this.someMapMutex.Unlock()
			}
		}
	}
}


func (this *LspClient) init(dir string) {
	this.id = 0
	id := this.id

	initializeRequest := InitializeRequest{
		ID: id, JSONRPC: "2.0",
		Method: "initialize",
		Params: InitializeParams{
			RootURI: "file://" + dir, RootPath: dir,
			WorkspaceFolders: []WorkspaceFolder{{ Name: "edgo", URI:  "file://" + dir }},
			Capabilities: capabilities,
			ClientInfo: ClientInfo{ Name: "edgo",Version: "1.0.0"},
		},
	}

	this.send(initializeRequest)

	response := this.waitForResponse(id, 30000)

	if response == "" {
		this.logger.info("cant get initialize response from lsp server")
		this.isReady = false
		return
	}

	initializedRequest := InitializedRequest{
		JSONRPC: "2.0", Method:  "initialized", Params:  struct{}{},
	}
	this.send(initializedRequest)
	this.logger.info("lsp initialized ")
	this.isReady = true
}

func (this *LspClient) waitForResponse(id, ms int) string {
	start := time.Now()
	for {
		if time.Since(start) >= time.Millisecond * time.Duration(ms) { return "" }
		this.someMapMutex.Lock()
		value, ok := this.responsesMap[id]
		this.someMapMutex.Unlock()
		if ok {
			this.someMapMutex.Lock()
			delete(this.responsesMap, id)
			this.someMapMutex.Unlock()
			return value
		}
	}
}

func (this *LspClient) didOpen(file string, lang string) {
	filecontent, err := os.ReadFile(file)
	if err != nil { this.logger.error(err.Error());; return }


	didOpenRequest := DidOpenRequest{
		JSONRPC: "2.0",
		Method:  "textDocument/didOpen",
		Params: DidOpenParams{
			TextDocument: TextDocument{
				LanguageID: lang,
				Text:       string(filecontent),
				URI:        "file://" + file,
				Version:    1,
			},
		},
	}

	this.send(didOpenRequest)
}
func (this *LspClient) didChange(file string, startline, startcharacter,endline, endcharacter int, text string) {
	this.id++
	id := this.id

	didChangeRequest := DidChangeRequest{
		Jsonrpc: "2.0", Method:  "textDocument/didChange",
		Params: DidChangeParams{
			ContentChanges: []ContentChange{
				{
					Range: ChangeRange{
						Start: Character{Character: startcharacter, Line: startline},
						End:   Character{Character: endcharacter, Line: endline},
					},
					Text: text,
				},
			},
			TextDocument: TextDocument{
				URI:     "file://" + file,
				Version: id,
			},
		},
	}

	this.send(didChangeRequest)

	this.someMapMutex.Lock()
	delete(this.file2diagnostic,  "file://" + file)
	this.someMapMutex.Unlock()
}

func (this *LspClient) didSave(file string) {
	request := DidSaveRequest{
		Jsonrpc: "2.0", Method:  "textDocument/didSave",
		Params: DidSaveParams{
			TextDocument: TextDocument{ URI: "file://" + file },
		},
	}

	this.send(request)
}

func (this *LspClient) references(file string, line int, character int) (ReferencesResponse, error) {
	this.id++
	id := this.id

	referencesRequest := BaseRequest{
		ID:      id,
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

	jsonData := this.waitForResponse(id,10000)
	if jsonData == "" { this.logger.error("cant get hover response from lsp server") }

	var response ReferencesResponse
	err := json.Unmarshal([]byte(jsonData), &response)
	if err != nil { this.logger.error("Error parsing JSON:" + err.Error()) }
	return response, err

}

func (this *LspClient) hover(file string, line int, character int) (HoverResponse, error) {
	this.id++
	id := this.id

	request := BaseRequest{
		ID: id, JSONRPC: "2.0", Method:  "textDocument/hover",
		Params: Params {
			TextDocument: TextDocument { URI: "file://" + file },
			Position: Position { Line: line, Character: character },
		},
	}

	this.send(request)

	jsonData := this.waitForResponse(id,1000)
	if jsonData == "" { this.logger.error("cant get hover response from lsp server") }

	var response HoverResponse
	err := json.Unmarshal([]byte(jsonData), &response)
	if err != nil { this.logger.error("Error parsing JSON:" + err.Error()) }
	return response, err
}
func (this *LspClient) signatureHelp(file string, line int, character int) (SignatureHelpResponse, error) {
	this.id++
	id := this.id

	request := BaseRequest{
		ID: id, JSONRPC: "2.0", Method:  "textDocument/signatureHelp",
		Params: Params {
			TextDocument: TextDocument { URI: "file://" + file },
			Position: Position { Line: line, Character: character },
		},
	}

	this.send(request)

	jsonData := this.waitForResponse(id,1000)
	if jsonData == "" { this.logger.error("cant get signature help response from lsp server") }

	var response SignatureHelpResponse
	err := json.Unmarshal([]byte(jsonData), &response)
	if err != nil { this.logger.error("Error parsing JSON:" + err.Error()) }
	return response, err
}

func (this *LspClient) definition(file string, line int, character int) (DefinitionResponse, error) {
	this.id++
	id := this.id

	request := DefinitionRequest{
		ID: this.id, JSONRPC: "2.0", Method:  "textDocument/definition",
		Params: DefinitionParams {
			TextDocument: TextDocument{ URI: "file://" + file },
			Position: Position{ Line: line, Character: character, },
		},
	}

	this.send(request)

	jsonData := this.waitForResponse(id,1000)
	if jsonData == "" { this.logger.error("cant get definition response from lsp server") }

	var response DefinitionResponse
	err := json.Unmarshal([]byte(jsonData), &response)
	if err != nil { this.logger.error("Error parsing JSON:" + err.Error()) }
	return response, err
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


	jsonData := this.waitForResponse(id,1000)
	if jsonData == "" { this.logger.error("cant get completion response from lsp server") }


	var completionResponse CompletionResponse
	err := json.Unmarshal([]byte(jsonData), &completionResponse)
	if err != nil { this.logger.error("Error parsing JSON:" + err.Error()) }
	return completionResponse, err
}

func (this *LspClient) read_stdout() (map[string]interface{}, string) {
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
			if err != nil { this.logger.error(err.Error()); continue }
			line = string(buf)
			messageSize = 0
			//fmt.Println("response", line)

			responseJSON := make(map[string]interface{})
			err = json.Unmarshal(buf, &responseJSON)
			if err != nil { this.logger.error(err.Error()); continue }

			method, found := responseJSON["method"];
			if found && method.(string) == "textDocument/publishDiagnostics" {
				var dr DiagnosticResponse
				err := json.Unmarshal(buf, &dr)
				if err != nil {  this.logger.error(err.Error()); continue }

				return responseJSON, line
			}

			if value, found := responseJSON["id"]; found {
				if _, ok := value.(float64); ok {
					return responseJSON, line
				}
			}

		} else {
			line, err = reader.ReadString('\n') // it stuck sometimes
			if err != nil { this.logger.error(err.Error()); continue }
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