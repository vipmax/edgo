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
	//stopped        bool
	process        *exec.Cmd      // The underlying process running the LSP server.
	lang2stdin     map[string]io.WriteCloser
	lang2stdout    map[string]io.ReadCloser
	lang2isReady   map[string]bool

	//stdin          io.WriteCloser // The standard input pipe for sending data to the LSP server.
	//stdout         io.ReadCloser
	responsesMap   map[int]string
	someMapMutex   sync.RWMutex
	//isReady   	   bool
	id        	   int
	file2diagnostic map[string]DiagnosticParams
	
	lang string
}

func (this *LspClient) start(language string, lspCmd []string) bool {
	//this.isReady = false
	//this.stopped = true
	this.someMapMutex = sync.RWMutex{}

	if this.lang2stdin == nil { this.lang2stdin = make(map[string]io.WriteCloser) }
	if this.lang2stdout == nil { this.lang2stdout = make(map[string]io.ReadCloser) }
	if this.lang2isReady == nil { this.lang2isReady = make(map[string]bool) }

	this.lang2isReady[language] = false

	_, lsperr := exec.LookPath(lspCmd[0])
	if lsperr != nil { logger.info("lsp not found:", lspCmd[0]); return false }

	this.process = exec.Command(lspCmd[0], lspCmd[1:]...)

	var stdin, err = this.process.StdinPipe()
	if err != nil { logger.info(err.Error()); return false  }
	this.lang2stdin[language] = stdin
	//this.stdin = stdin

	stdout, err := this.process.StdoutPipe()
	if err != nil { logger.info(err.Error()); return false }
	this.lang2stdout[language] = stdout
	//this.stdout = stdout

	err = this.process.Start()
	if err != nil { logger.info(err.Error()); return false  }

	//this.stopped = false
	this.responsesMap = make(map[int]string)
	this.file2diagnostic = make(map[string]DiagnosticParams)

	return true
}

func (this *LspClient) stop()  {
	// doesnt work
	if this.process == nil { return }

	//this.shutdown()
	this.exit()
	//this.stopped = true
	logger.info("successfully stopped lsp")
}

func (this *LspClient) send(o interface{})  {
	m, err := json.Marshal(o)
	if err != nil { panic(err) }
	logger.info("->", string(m))

	message := fmt.Sprintf("Content-Length: %d\r\n\r\n%s", len(m), m)
	stdin := this.lang2stdin[this.lang]
	_, err = stdin.Write([]byte(message))
	if err != nil {
		logger.error(err.Error())
	}
}


func (this *LspClient) receiveLoop(diagnosticUpdateChannel chan string, language string) {
	for {
		_, message := this.readStdout(language)
		logger.info("<-", message)

		if strings.Contains(message,"publishDiagnostics") {
			var dr DiagnosticResponse
			errp := json.Unmarshal([]byte(message), &dr)
			if errp != nil {
				logger.error(errp.Error()); continue
			}
			this.file2diagnostic[dr.Params.Uri] = dr.Params
			diagnosticUpdateChannel <- "update"
			continue
		}

		responseJSON := make(map[string]interface{})
		err := json.Unmarshal([]byte(message), &responseJSON)
		if err != nil {
			logger.error(err.Error()); continue
		}

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
		logger.info("cant get initialize response from lsp server")
		this.lang2isReady[this.lang] = false
		return
	}

	initializedRequest := InitializedRequest{
		JSONRPC: "2.0", Method:  "initialized", Params:  struct{}{},
	}
	this.send(initializedRequest)
	logger.info("lsp initialized ")
	//this.isReady = true
	this.lang2isReady[this.lang] = true
}

func (this *LspClient) shutdown() {
	this.id++
	id := this.id
	shutdownRequest := ShutdownRequest{
		ID: id, JSONRPC: "2.0", Method:  "shutdown",
	}
	this.send(shutdownRequest)
	//response := this.waitForResponse(id, 30000)
	//logger.info("shutdown ", response)
}
func (this *LspClient) exit() {
	shutdownRequest := ExitRequest{
		JSONRPC: "2.0", Method:  "exit",
	}
	this.send(shutdownRequest)
	logger.info("exit")
}

func (this *LspClient) waitForResponse(id, ms int) string {
	start := time.Now()
	for {
		if time.Since(start) >= time.Millisecond * time.Duration(ms) { return "" }
		this.someMapMutex.RLock()
		value, ok := this.responsesMap[id]
		this.someMapMutex.RUnlock()
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
	if err != nil { logger.error(err.Error()); return }


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
func (this *LspClient) didChange(file string, version int, startline, startcharacter,endline, endcharacter int, text string) {
	//this.id++
	//id := this.id

	didChangeRequest := DidChangeRequest{
		Jsonrpc: "2.0", Method:  "textDocument/didChange",
		Params: DidChangeParams{
			ContentChanges: []ContentChange{
				{
					Range: ChangeRange{
						Start: Character{Character: startcharacter, Line: startline},
						End:   Character{Character: endcharacter, Line: endline},
					},
					RangeLength: 0,
					Text: text,
				},
			},
			TextDocument: TextDocument{
				URI:     "file://" + file,
				Version: version,
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
	if jsonData == "" { logger.error("cant get hover response from lsp server") }

	var response ReferencesResponse
	err := json.Unmarshal([]byte(jsonData), &response)
	if err != nil { logger.error("Error parsing JSON:" + err.Error()) }
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
	if jsonData == "" { logger.error("cant get hover response from lsp server") }

	var response HoverResponse
	err := json.Unmarshal([]byte(jsonData), &response)
	if err != nil { logger.error("Error parsing JSON:" + err.Error()) }
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
	if jsonData == "" { logger.error("cant get signature help response from lsp server") }

	var response SignatureHelpResponse
	err := json.Unmarshal([]byte(jsonData), &response)
	if err != nil { logger.error("Error parsing JSON:" + err.Error()) }
	return response, err
}

func (this *LspClient) definition(file string, line int, character int) (DefinitionResponse, error) {
	this.id++
	id := this.id

	request := DefinitionRequest{
		ID: this.id, JSONRPC: "2.0", Method:  "textDocument/definition",
		Params: DefinitionParams {
			TextDocument: TextDocument{ URI: "file://" + file },
			Position: Position{ Line: line, Character: character },
		},
	}

	this.send(request)

	jsonData := this.waitForResponse(id,1000)
	if jsonData == "" { logger.error("cant get definition response from lsp server") }

	var response DefinitionResponse
	err := json.Unmarshal([]byte(jsonData), &response)
	if err != nil { logger.error("Error parsing JSON:" + err.Error()) }
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
	if jsonData == "" { logger.error("cant get completion response from lsp server") }


	var completionResponse CompletionResponse
	err := json.Unmarshal([]byte(jsonData), &completionResponse)
	if err != nil { logger.error("Error parsing JSON:" + err.Error()) }
	return completionResponse, err
}

func (this *LspClient) prepareRename(file string, line int, character int) (PrepareRenameResponse, error) {
	this.id++
	id := this.id

	request := PrepareRenameRequest {
		ID: id, Jsonrpc: "2.0", Method:  "textDocument/prepareRename",
		Params: Params{
			TextDocument: TextDocument { URI:  "file://" + file },
			Position: Position { Line: line, Character: character },
		},
	}

	this.send(request)

	jsonData := this.waitForResponse(id,1000)
	if jsonData == "" { logger.error("cant get rename response from lsp server") }

	var response PrepareRenameResponse
	err := json.Unmarshal([]byte(jsonData), &response)
	if err != nil { logger.error("Error parsing JSON:" + err.Error()) }
	return response, err
}

func (this *LspClient) rename(file string, newname string, line int, character int) (RenameResponse, error) {
	this.id++
	id := this.id

	request := RenameRequest{
		ID: id,  Jsonrpc: "2.0", Method:  "textDocument/rename",
		Params: RenameParams {
			NewName: newname,
			Position: Position { Line: line, Character: character },
			TextDocument: TextDocument { URI:  "file://" + file },
		},
	}

	this.send(request)

	jsonData := this.waitForResponse(id,1000)
	if jsonData == "" { logger.error("cant get rename response from lsp server") }

	var response RenameResponse
	err := json.Unmarshal([]byte(jsonData), &response)
	if err != nil { logger.error("Error parsing JSON:" + err.Error()) }
	return response, err
}

func (this *LspClient) codeAction(file string, spc int, spl int, epc int, epl int) (CodeActionResponse, error) {
	this.id++
	id := this.id

	request := CodeActionRequest {
		ID: id,  Jsonrpc: "2.0", Method: "textDocument/codeAction",
		Params: CodeActionParams {
			TextDocument: TextDocument { URI:  "file://" + file },
			Context: Context{ Only: []string{"refactor"}, TriggerKind: 1 },
			Range: RequestRange{
				Start: Position{ Line: spl, Character: spc},
				End: Position{ Line: epl, Character: epc},
			},
		},
	}

	this.send(request)

	jsonData := this.waitForResponse(id,1000)
	if jsonData == "" { logger.error("cant get rename response from lsp server") }

	var response CodeActionResponse
	err := json.Unmarshal([]byte(jsonData), &response)
	if err != nil { logger.error("Error parsing JSON:" + err.Error()) }
	return response, err
}

func (this *LspClient) command(command Command) (CommandResponse, error) {
	this.id++
	id := this.id

	request := CommandRequest {
		ID: id,  Jsonrpc: "2.0", Method: "workspace/executeCommand",
		Params: command,
	}

	this.send(request)

	jsonData := ""
	ok := false; key := 0
	start := time.Now()

	for {
		if time.Since(start) >= time.Millisecond * time.Duration(3000) { break }
		this.someMapMutex.RLock()
		for k, value := range this.responsesMap {
			if strings.Contains(value, "workspace/applyEdit") {
				jsonData = value
				ok = true
				key = k
				break
			}
		}
		this.someMapMutex.RUnlock()

		if ok {
			this.someMapMutex.Lock()
			delete(this.responsesMap, key)
			this.someMapMutex.Unlock()
			break
		}
	}

	//jsonData := this.waitForResponse(id,1000)
	if jsonData == "" { logger.error("cant get rename response from lsp server") }

	var response CommandResponse
	err := json.Unmarshal([]byte(jsonData), &response)
	if err != nil { logger.error("Error parsing JSON:" + err.Error()) }
	return response, err
}

func (this *LspClient) applyEdit(key int) {
	//this.id++
	//id := this.id
	request := ApplyEditRequest {
		ID: key,  Jsonrpc: "2.0",
		Result:  Applied { true } ,
	}

	this.send(request)

}

func (this *LspClient) readStdout(language string) (map[string]interface{}, string) {
	//start := time.Now()

	const LEN_HEADER = "Content-Length: "
	stdout := this.lang2stdout[language]
	reader := bufio.NewReader(stdout)

	var messageSize int
	var responseMustBeNext bool

	for {
		var line string
		var err error

		if messageSize != 0 && responseMustBeNext {
			buf := make([]byte, messageSize)
			_, err = io.ReadFull(reader, buf)
			if err != nil { logger.error(err.Error()); continue }
			line = string(buf)
			messageSize = 0
			//fmt.Println("response", line)

			responseJSON := make(map[string]interface{})
			err = json.Unmarshal(buf, &responseJSON)
			if err != nil {
				logger.error(err.Error()); continue
			}

			method, found := responseJSON["method"]
			if found && method.(string) == "textDocument/publishDiagnostics" {
				var dr DiagnosticResponse
				err := json.Unmarshal(buf, &dr)
				if err != nil {
					logger.error("[432 lsp]", err.Error()); continue
				}

				return responseJSON, line
			}

			if value, found := responseJSON["id"]; found {
				if _, ok := value.(float64); ok {
					return responseJSON, line
				}
			}

		} else {
			line, err = reader.ReadString('\n') // it stuck sometimes
			//if  err != nil && err.Error() == "EOF" {
			//	logger.error("[445 lsp] ", err.Error());
			//	break
			//}
			if err != nil {
				//if err.Error() == "read |0: file already closed" {
				//	break
				//}
				logger.error("[445 lsp]", err.Error()); continue
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

func (this *LspClient) IsLangReady(language string) bool {
	ready, found := this.lang2isReady[language]
	if !found { return false } else { return ready }
}