package lsp

import (
	"bufio"
	"context"
	. "edgo/internal/logger"
	"fmt"
	"github.com/goccy/go-json"
	"io"
	"os"
	"os/exec"
	"os/signal"
	"strconv"
	"strings"
	"time"
)

type LspClient struct {
	Lang  string
	Cmd   *exec.Cmd
	stdin io.WriteCloser
	stdout io.ReadCloser
	stop   context.CancelFunc
	//reader    *textproto.Reader
	reader *bufio.Reader

	IsReady   bool
	isStopped bool

	message2chan          map[int]chan string
	completionMessages    chan string
	definitionMessages    chan string
	referencesMessages    chan string
	signatureHelpMessages chan string
	hoverMessages         chan string
	otherMessages         chan string
	DiagnosticsChannel    chan string

	id              int
	file2diagnostic map[string]DiagnosticParams
}



func (l *LspClient) Start(cmd string, args ...string) bool {
	Log.Info("starting lsp", cmd, strings.Join(args," "))

	_, err := exec.LookPath(cmd)
	if err != nil { Log.Info("lsp not found ", cmd); return false }

	ctx, stop := signal.NotifyContext(context.Background(), os.Kill)
	l.Cmd = exec.CommandContext(ctx, cmd, args...)
	l.stop = stop

	stdin, err := l.Cmd.StdinPipe()
	if err != nil { Log.Info(err.Error()); return false }
	l.stdin = stdin

	stdout, err := l.Cmd.StdoutPipe()
	if err != nil { Log.Info(err.Error()); return false }
	l.stdout = stdout

	errorsChannel := make(chan error)

	// starting lsp Cmd async
	go func() {
		startError := l.Cmd.Run()
		if startError != nil { errorsChannel <- startError }
		close(errorsChannel)
	}()

	// wait for start
	var end = false
	for !end {
		select {
		case startError := <- errorsChannel:
			if startError != nil {
				Log.Error("error starting lsp " + startError.Error())
				l.isStopped = true
				return false
			}

		case <-time.After(time.Duration(100) * time.Millisecond):
			Log.Info("lsp started success", cmd, strings.Join(args," "))
			end = true
		}
	}

	//l.reader = textproto.NewReader(bufio.NewReader(stdout))
	l.reader = bufio.NewReader(stdout)

	l.message2chan = make(map[int]chan string)
	l.completionMessages = make(chan string)
	l.referencesMessages = make(chan string)
	l.definitionMessages = make(chan string)
	l.signatureHelpMessages = make(chan string)
	l.hoverMessages = make(chan string)
	l.otherMessages = make(chan string)
	l.DiagnosticsChannel = make(chan string, 10)
	l.file2diagnostic = make(map[string]DiagnosticParams)
	l.IsReady = true

	go l.receiveLoop()

	return true
}

func (l *LspClient) Stop() {
	if l.isStopped { return }
	l.stop()
	l.isStopped = true
}


func (this *LspClient) send(o interface{})  {
	m, err := json.Marshal(o)
	if err != nil { panic(err) }

	Log.Info("->", string(m))
	message := fmt.Sprintf("Content-Length: %d\r\n\r\n%s", len(m), m)
	_, err = this.stdin.Write([]byte(message))
	if err != nil { Log.Error(err.Error()) }
}

//func (l *LspClient) receive() string {
//	headers, err := l.reader.ReadMIMEHeader()
//	if err != nil { fmt.Println(err); return "" }
//
//	length, err := strconv.Atoi(headers.Get("Content-Length"))
//	if err != nil { fmt.Println(err); return ""}
//
//	body := make([]byte, length)
//	if _, err := l.reader.R.Read(body); err != nil { fmt.Println(err); return "" }
//
//	return string(body)
//}

func (this *LspClient) receive() string {

	const LEN_HEADER = "Content-Length: "
	var messageSize int
	var responseMustBeNext bool
	var line string
	var err error

	for !this.isStopped {

		if messageSize != 0 && responseMustBeNext {
			buf := make([]byte, messageSize)
			_, err = io.ReadFull(this.reader, buf)
			if err != nil { Log.Error(err.Error()); continue }
			line = string(buf)
			messageSize = 0

			responseJSON := make(map[string]interface{})
			err = json.Unmarshal(buf, &responseJSON)
			if err != nil { Log.Error(err.Error()); continue }

			method, methodFound := responseJSON["method"]
			if methodFound && method.(string) == "textDocument/publishDiagnostics" {
				var dr DiagnosticResponse
				err = json.Unmarshal(buf, &dr)
				if err != nil { Log.Error(err.Error()); continue }
				return line
			}

			if value, idFound := responseJSON["id"]; idFound {
				if _, ok := value.(float64); ok {
					return line
				}
			}

		} else {
			line, err = this.reader.ReadString('\n') // it stuck sometimes
			if err != nil { Log.Error("[445 lsp]", err.Error()); continue }
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
	return ""
}


func (l *LspClient) receiveLoop() {
	for !l.isStopped {
		message := l.receive()
		Log.Info("<-", message)

		if strings.Contains(message,"publishDiagnostics") {
			var dr DiagnosticResponse
			err := json.Unmarshal([]byte(message), &dr)
			if err != nil { Log.Error(err.Error()); continue }
			l.file2diagnostic[dr.Params.Uri] = dr.Params
			l.DiagnosticsChannel <- message
			continue
		}
		if strings.Contains(message,"workspace/applyEdit") {
			l.otherMessages <- message
			continue
		}

		responseJSON := make(map[string]interface{})
		err := json.Unmarshal([]byte(message), &responseJSON)
		if err != nil { Log.Error(err.Error()); continue }

		if value, found := responseJSON["id"]; found { // json has id
			if id, ok := value.(float64); ok {
				channel, foundRequest := l.message2chan[int(id)]
				if foundRequest {
					channel <- message
				} else  {
					//skip message
				}
			}
		}
	}
}

func (this *LspClient) GetDiagnostic(filename string) (DiagnosticParams, bool) {
	d, found := this.file2diagnostic[filename]
	return  d, found
}

func WaitForRequest[T any](channel chan string, timeout int) (T, error) {
	var response T
	var err error

	select {
	case jsonData := <- channel:
		err = json.Unmarshal([]byte(jsonData), &response)
		if err != nil { Log.Error("Error parsing JSON:" + err.Error()) }

	case <-time.After(time.Duration(timeout) * time.Millisecond):
		err = fmt.Errorf("Timeout")
	}

	return response, err
}


func (l *LspClient) Init(dir string) {
	l.id = 0
	id := l.id

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

	l.message2chan[id] = l.otherMessages
	l.send(initializeRequest)

	response, err := WaitForRequest[interface{}](l.otherMessages, 3000)

	delete(l.message2chan, id)

	if response == "" || err != nil {
		Log.Info("cant get initialize response from lsp server")
		l.IsReady = false
		return
	}

	initializedRequest := InitializedRequest{
		JSONRPC: "2.0", Method:  "initialized", Params:  struct{}{},
	}
	l.send(initializedRequest)

	Log.Info("lsp initialized")
	l.IsReady = true
}

func (this *LspClient) DidOpen(file string, text string) {
	filecontent, err := os.ReadFile(file)
	if err != nil { Log.Error(err.Error()); return }

	didOpenRequest := DidOpenRequest{
		JSONRPC: "2.0",  Method:  "textDocument/didOpen",
		Params: DidOpenParams{
			TextDocument: TextDocument{
				LanguageID: this.Lang,
				Text:       string(filecontent),
				URI:        "file://" + file,
				Version:    1,
			},
		},
	}

	this.send(didOpenRequest)
}

func (this *LspClient) DidClose(file string) {
	request := DidOpenRequest{
		JSONRPC: "2.0",  Method:  "textDocument/didClose",
		Params: DidOpenParams{
			TextDocument: TextDocument{
				LanguageID: this.Lang,
				URI:        "file://" + file,
				Version:    1,
			},
		},
	}

	this.send(request)
}


func (this *LspClient) Hover(file string, line int, character int) (HoverResponse, error) {
	this.id++
	id := this.id

	request := BaseRequest{
		ID: id, JSONRPC: "2.0", Method:  "textDocument/hover",
		Params: Params {
			TextDocument: TextDocument { URI: "file://" + file },
			Position: Position { Line: line, Character: character },
		},
	}

	this.message2chan[id] = this.hoverMessages
	this.send(request)

	response, err := WaitForRequest[HoverResponse](this.hoverMessages, 1000)

	delete(this.message2chan, id)
	return response, err
}

func (this *LspClient) Completion(file string, line int, character int) (CompletionResponse, error) {
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

	this.message2chan[id] = this.completionMessages
	this.send(request)

	response, err := WaitForRequest[CompletionResponse](this.completionMessages, 1000)

	delete(this.message2chan, id)
	return response, err
}

func (this *LspClient) Definition(file string, line int, character int) (DefinitionResponse, error) {
	this.id++
	id := this.id

	request := DefinitionRequest{
		ID: this.id, JSONRPC: "2.0", Method:  "textDocument/definition",
		Params: DefinitionParams {
			TextDocument: TextDocument{ URI: "file://" + file },
			Position: Position{ Line: line, Character: character },
		},
	}

	this.message2chan[id] = this.definitionMessages
	this.send(request)

	response, err := WaitForRequest[DefinitionResponse](this.definitionMessages, 1000)

	delete(this.message2chan, id)
	return response, err
}

func (this *LspClient) SignatureHelp(file string, line int, character int) (SignatureHelpResponse, error) {
	this.id++
	id := this.id

	request := BaseRequest{
		ID: id, JSONRPC: "2.0", Method:  "textDocument/signatureHelp",
		Params: Params {
			TextDocument: TextDocument { URI: "file://" + file },
			Position: Position { Line: line, Character: character },
		},
	}

	this.message2chan[id] = this.signatureHelpMessages
	this.send(request)

	response, err := WaitForRequest[SignatureHelpResponse](this.signatureHelpMessages, 1000)

	delete(this.message2chan, id)
	return response, err
}

func (this *LspClient) References(file string, line int, character int) (ReferencesResponse, error) {
	this.id++
	id := this.id

	request := BaseRequest{
		ID: id, JSONRPC: "2.0", Method:  "textDocument/references",
		Params: Params{
			TextDocument: TextDocument{ URI: "file://" + file },
			Position: Position{ Line: line, Character: character },
			Context: Context{ IncludeDeclaration: false },
		},
	}

	this.message2chan[id] = this.referencesMessages
	this.send(request)

	response, err := WaitForRequest[ReferencesResponse](this.referencesMessages, 3000)

	delete(this.message2chan, id)
	return response, err
}


func (this *LspClient) PrepareRename(file string, line int, character int) (PrepareRenameResponse, error) {
	this.id++
	id := this.id

	request := PrepareRenameRequest {
		ID: id, Jsonrpc: "2.0", Method:  "textDocument/prepareRename",
		Params: Params{
			TextDocument: TextDocument { URI:  "file://" + file },
			Position: Position { Line: line, Character: character },
		},
	}

	this.message2chan[id] = this.otherMessages
	this.send(request)

	response, err := WaitForRequest[PrepareRenameResponse](this.otherMessages, 10000)

	delete(this.message2chan, id)
	return response, err
}

func (this *LspClient) Rename(file string, newname string, line int, character int) (RenameResponse, error) {
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

	this.message2chan[id] = this.otherMessages
	this.send(request)

	response, err := WaitForRequest[RenameResponse](this.otherMessages, 10000)

	delete(this.message2chan, id)
	return response, err
}

func (this *LspClient) CodeAction(file string, spc int, spl int, epc int, epl int) (CodeActionResponse, error) {
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

	this.message2chan[id] = this.otherMessages
	this.send(request)

	response, err := WaitForRequest[CodeActionResponse](this.otherMessages, 10000)

	delete(this.message2chan, id)
	return response, err
}

func (this *LspClient) Command(command Command) (CommandResponse, error) {
	this.id++
	id := this.id

	request := CommandRequest {
		ID: id,  Jsonrpc: "2.0", Method: "workspace/executeCommand",
		Params: command,
	}

	this.send(request)

	jsonData := <- this.otherMessages

	var response CommandResponse
	err := json.Unmarshal([]byte(jsonData), &response)
	if err != nil { Log.Error("Error parsing JSON:" + err.Error()) }
	return response, err
}


func (this *LspClient) ApplyEdit(key int) {
	request := ApplyEditRequest {
		ID: key,  Jsonrpc: "2.0",
		Result:  Applied { true } ,
	}
	this.send(request)
}
