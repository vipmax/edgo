package dap

import (
	"bufio"
	"context"
	. "edgo/internal/logger"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

type DapClient struct {
	Lang   string
	cmd    *exec.Cmd
	stdin  io.WriteCloser
	stdout io.ReadCloser
	stop   context.CancelFunc
	reader *bufio.Reader

	IsReady   bool
	IsStopped bool
	IsStarted bool

	Breakpoints map[string][]int

	message2chan   map[int]chan string
	otherMessages  chan string
	StdoutMessages chan string
	EventMessages  chan string

	id int

	Conntype string
	Port     int
	conn     net.Conn
}

func (this *DapClient) Start(cmd string, args ...string) bool {
	Log.Info("starting dap ", cmd, strings.Join(args, " "))

	_, err := exec.LookPath(cmd)
	if err != nil { Log.Info("lsp not found ", cmd); return false }

	ctx, stop := signal.NotifyContext(context.Background(), os.Kill)
	this.cmd = exec.CommandContext(ctx, cmd, args...)
	this.cmd.Env = append(os.Environ(), "PYTHONUNBUFFERED=1")

	this.stop = stop


	stdin, err := this.cmd.StdinPipe()
	if err != nil { Log.Info(err.Error()); return false }
	this.stdin = stdin

	stdout, err := this.cmd.StdoutPipe()
	if err != nil { Log.Info(err.Error()); return false }
	this.stdout = stdout

	errorsChannel := make(chan error)

	// starting cmd async
	go func() {
		startError := this.cmd.Run()
		if startError != nil { errorsChannel <- startError }
		close(errorsChannel)
	}()

	// wait for start
	var end = false
	for !end {
		select {
		case startError := <-errorsChannel:
			if startError != nil {
				Log.Error("error starting dap." + startError.Error())
				this.IsStopped = true
				this.IsStarted = false
				return false
			}

		case <-time.After(time.Duration(100) * time.Millisecond):
			Log.Info("dap started.", cmd, strings.Join(args, " "))
			end = true

		}
	}

	if this.Conntype == "tcp" {
		time.Sleep(300 * time.Millisecond)
		conn, err := net.Dial("tcp", fmt.Sprintf("localhost:%d", this.Port))
		if err != nil {
			Log.Error("Error connecting:", err.Error())
			return false
		}

		this.reader = bufio.NewReader(conn)
		this.conn = conn
	}

	if this.Conntype == "stdio" {
		this.reader = bufio.NewReader(stdout)
	}

	if this.Breakpoints == nil {
		this.Breakpoints = make(map[string][]int)
	}

	this.message2chan = make(map[int]chan string)
	this.otherMessages = make(chan string)
	this.StdoutMessages = make(chan string)
	this.EventMessages = make(chan string)
	this.IsReady = true
	this.IsStarted = true

	go this.readStdout()
	go this.readTcp()

	return true
}

func (this *DapClient) Stop() {
	if this.IsStopped {
		return
	}
	this.stop()
	//close(this.StdoutMessages)
	//close(this.EventMessages)
	this.IsStopped = true
	this.IsStarted = false
	this.IsReady = false
}

func (this *DapClient) readStdout() {
	scanner := bufio.NewScanner(this.stdout)
	for scanner.Scan() {
		message := scanner.Text()
		if this.IsStopped {
			break
		}
		this.StdoutMessages <- message
	}
	close(this.StdoutMessages)
}

func (this *DapClient) send(o interface{}) {
	m, err := json.Marshal(o)
	if err != nil {
		panic(err)
	}

	Log.Info("->", string(m))
	message := fmt.Sprintf("Content-Length: %d\r\n\r\n%s", len(m), m)
	_, err = this.conn.Write([]byte(message))
	if err != nil {
		Log.Error(err.Error())
	}
}

//func (l *DapClient) receive() string {
//	reader := textproto.NewReader(l.reader)
//	headers, err :=  reader.ReadMIMEHeader()
//	if err != nil { fmt.Println(err); return "" }
//
//	length, err := strconv.Atoi(headers.Get("Content-Length"))
//	if err != nil { fmt.Println(err); return ""}
//
//	body := make([]byte, length)
//	if _, err := reader.R.Read(body); err != nil { fmt.Println(err); return "" }
//
//	return string(body)
//}

func (this *DapClient) receive() string {

	const LEN_HEADER = "Content-Length: "
	var messageSize int
	var responseMustBeNext bool
	var line string
	var err error

	for {

		if messageSize != 0 && responseMustBeNext {
			buf := make([]byte, messageSize)
			_, err = io.ReadFull(this.reader, buf)
			if err != nil {
				Log.Error(err.Error())
				continue
			}
			line = string(buf)
			messageSize = 0

			responseJSON := make(map[string]interface{})
			err = json.Unmarshal(buf, &responseJSON)
			if err != nil {
				Log.Error(err.Error())
				continue
			}

			return line

		} else {
			line, err = this.reader.ReadString('\n') // it stuck sometimes
			if err != nil {
				Log.Error(err.Error())
				break
			}
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

func (this *DapClient) readTcp() {
	for {
		if this.IsStopped { break }
		message := this.receive()
		Log.Info("<-", message)
		if this.IsStopped { break }

		responseJson := make(map[string]interface{})
		err := json.Unmarshal([]byte(message), &responseJson)
		if err != nil {
			Log.Error(err.Error())
			continue
		}

		if _, found := responseJson["event"]; found { // json has event
			this.EventMessages <- message
			continue
		}

		if value, found := responseJson["request_seq"]; found { // json has id
			if id, ok := value.(float64); ok {
				channel, foundRequest := this.message2chan[int(id)]
				if foundRequest {
					channel <- message
				} else {
					//skip message
				}
			}
		}
	}

	close(this.EventMessages)
}

func WaitForRequest[T any](channel chan string, timeout int) (T, error) {
	var response T
	var err error

	select {
	case jsonData := <-channel:
		err = json.Unmarshal([]byte(jsonData), &response)
		if err != nil {
			Log.Error("Error parsing JSON:" + err.Error())
		}

	case <-time.After(time.Duration(timeout) * time.Millisecond):
		err = fmt.Errorf("Timeout")
	}

	return response, err
}

func (this *DapClient) Init(dir string) bool {
	this.id = 1
	id := this.id

	initializeRequest := InitializeRequest{
		RequestSeq: id,
		Seq:     id,
		Type:    "request",
		Command: "initialize",
		Arguments: Arguments{
			ClientID:                     "edgo",
			ClientName:                   "edgo",
			AdapterID:                    this.Lang,
			Locale:                       "en",
			LinesStartAt1:                true,
			ColumnsStartAt1:              true,
			PathFormat:                   "path",
			//SupportsVariableType:         true,
			//SupportsVariablePaging:       true,
			//SupportsRunInTerminalRequest: true,
			//SupportsMemoryReferences:     true,
			//SupportsProgressReporting:    true,
			//SupportsInvalidatedEvent:     true,
			//SupportsMemoryEvent:          true,
		},
	}

	this.message2chan[id] = this.otherMessages
	this.send(initializeRequest)

	response, err := WaitForRequest[Response](this.otherMessages, 5000)
	if err != nil { Log.Error(err.Error()) }

	delete(this.message2chan, id)

	this.IsReady = true

	return response.Success
}

func (this *DapClient) Launch(program string) bool {

	this.id++
	id := this.id

	name := filepath.Base(program)

	launchRequest := LaunchRequest{
		Seq:     id, Type:    "request", Command: "launch",
		Arguments: LaunchRequestArguments{
			Name:    "Launch debug " + name,
			Type:    this.Lang,
			Request: "launch",
			Mode:    "debug",
			Program: program,
		},
	}

	this.message2chan[id] = this.otherMessages
	this.send(launchRequest)

	response, err := WaitForRequest[LaunchResponse](this.otherMessages, 3000)
	if err != nil { Log.Error(err.Error()) }

	delete(this.message2chan, id)
	return response.Success
}

func (this *DapClient) Attach(restart bool) bool {
	this.id++
	id := this.id

	request := AttachRequest{
		Seq: id, Type: "request", Command: "attach",
		Arguments: AttachArguments{Restart: restart},
	}

	this.message2chan[id] = this.otherMessages
	this.send(request)

	response, err := WaitForRequest[ContinueResponse](this.otherMessages, 1000)
	if err != nil { Log.Error(err.Error()) }

	delete(this.message2chan, id)
	return response.Success
}
func (this *DapClient) ConfigurationDone() bool {
	this.id++
	id := this.id

	request := AttachRequest{ Seq: id, Type: "request", Command: "configurationDone"}

	this.message2chan[id] = this.otherMessages
	this.send(request)

	response, err := WaitForRequest[ContinueResponse](this.otherMessages, 1000)
	if err != nil { Log.Error(err.Error()) }

	delete(this.message2chan, id)
	return response.Success
}

func (this *DapClient) SetBreakpoint(file string, line int) bool {

	this.id++
	id := this.id

	name := filepath.Base(file)

	bps := this.Breakpoints[file]
	if line != -1 {
		bps = append(bps, line)
	}

	this.Breakpoints[file] = bps

	breakpoints := []Breakpoint{}
	for _, b := range bps {
		breakpoints = append(breakpoints, Breakpoint{Line: b})
	}

	request := SetBreakpointsRequest{
		Seq: id, Type: "request", Command: "setBreakpoints",
		Arguments: SetBreakpointsArguments{
			Source:      Source{Name: name, Path: file},
			Breakpoints: breakpoints,
			Lines:       bps,
		},
	}

	this.message2chan[id] = this.otherMessages
	this.send(request)

	response, err := WaitForRequest[SetBreakpointsResponse](this.otherMessages, 1000)
	if err != nil { Log.Error(err.Error()) }

	delete(this.message2chan, id)
	return response.Success
}
func (this *DapClient) SetAllBreakpoints() {
	for file := range this.Breakpoints {
		this.SetBreakpoint(file, -1)
	}
}

func (this *DapClient) Continue(threadID int) bool {

	this.id++
	id := this.id

	request := ContinueRequest{
		Seq: id, Type: "request", Command: "continue",
		Arguments: ContinueArguments{threadID},
	}

	this.message2chan[id] = this.otherMessages
	this.send(request)

	response, err := WaitForRequest[ContinueResponse](this.otherMessages, 3000)
	if err != nil { Log.Error(err.Error()) }

	delete(this.message2chan, id)
	return response.Success
}

func (this *DapClient) Next(threadID int) bool {

	this.id++
	id := this.id

	request := ContinueRequest{
		Seq: id, Type: "request", Command: "next",
		Arguments: ContinueArguments{threadID},
	}

	this.message2chan[id] = this.otherMessages
	this.send(request)

	response, err := WaitForRequest[ContinueResponse](this.otherMessages, 3000)
	if err != nil { Log.Error(err.Error()) }

	delete(this.message2chan, id)
	return response.Success
}

func (this *DapClient) StepIn(threadID int) bool {

	this.id++
	id := this.id

	request := ContinueRequest{
		Seq: id, Type: "request", Command: "stepIn",
		Arguments: ContinueArguments{threadID},
	}

	this.message2chan[id] = this.otherMessages
	this.send(request)

	response, err := WaitForRequest[ContinueResponse](this.otherMessages, 3000)
	if err != nil { Log.Error(err.Error()) }

	delete(this.message2chan, id)
	return response.Success
}

func (this *DapClient) Pause(threadID int) bool {

	this.id++
	id := this.id

	request := ContinueRequest{
		Seq: id, Type: "request", Command: "pause",
		Arguments: ContinueArguments{threadID},
	}

	this.message2chan[id] = this.otherMessages
	this.send(request)

	response, err := WaitForRequest[ContinueResponse](this.otherMessages, 3000)
	if err != nil { Log.Error(err.Error()) }

	delete(this.message2chan, id)
	return response.Success
}

func (this *DapClient) Threads(threadID int) ThreadsResponse {

	this.id++
	id := this.id

	request := ContinueRequest{
		Seq: id, Type: "request", Command: "threads",
		Arguments: ContinueArguments{threadID},
	}

	this.message2chan[id] = this.otherMessages
	this.send(request)

	response, err := WaitForRequest[ThreadsResponse](this.otherMessages, 200)
	if err != nil { Log.Error(err.Error()) }

	delete(this.message2chan, id)
	return response
}

func (this *DapClient) Stacktrace(threadID int, levels int) StackTraceResponse {

	this.id++
	id := this.id

	request := StacktraceRequest{
		Seq: id, Type: "request", Command: "stackTrace",
		Arguments: StacktraceArguments{ThreadID: threadID, Levels: levels},
	}

	this.message2chan[id] = this.otherMessages
	this.send(request)

	response, err := WaitForRequest[StackTraceResponse](this.otherMessages, 200)
	if err != nil { Log.Error(err.Error()) }

	delete(this.message2chan, id)
	return response
}

func (this *DapClient) Scopes(frameId int) ScopesResponse {

	this.id++
	id := this.id

	request := ScopeRequest{
		Seq: id, Type: "request", Command: "scopes",
		Arguments: ScopeArguments{frameId},
	}

	this.message2chan[id] = this.otherMessages
	this.send(request)

	response, err := WaitForRequest[ScopesResponse](this.otherMessages, 200)
	if err != nil { Log.Error(err.Error()) }

	delete(this.message2chan, id)
	return response
}

func (this *DapClient) Variables(variablesReference int) VariablesResponse {

	this.id++
	id := this.id

	request := VariablesRequest{
		Seq: id, Type: "request", Command: "variables",
		Arguments: VariablesArguments{variablesReference},
	}

	this.message2chan[id] = this.otherMessages
	this.send(request)

	response, err := WaitForRequest[VariablesResponse](this.otherMessages, 200)
	if err != nil { Log.Error(err.Error()) }

	delete(this.message2chan, id)
	return response
}

func (this *DapClient) Disconnect() bool {
	this.id++
	id := this.id

	request := ContinueRequest{
		Seq: id, Type: "request", Command: "disconnect",
	}

	this.message2chan[id] = this.otherMessages
	this.send(request)

	response, err := WaitForRequest[ContinueResponse](this.otherMessages, 3000)
	if err != nil { Log.Error(err.Error()) }

	delete(this.message2chan, id)
	return response.Success
}
