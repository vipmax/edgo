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
	isStopped bool

	Breakpoints  map[string][]int

	message2chan  map[int]chan string
	otherMessages  chan string
	StdoutMessages chan string
	EventMessages  chan string

	id   int

	conntype string
	port int
	conn net.Conn
}

func (this *DapClient) Start(cmd string, args ...string) bool {
	Log.Info("starting dap. ", cmd, strings.Join(args," "))

	_, err := exec.LookPath(cmd)
	if err != nil { Log.Info("lsp not found ", cmd); return false }

	ctx, stop := signal.NotifyContext(context.Background(), os.Kill)
	this.cmd = exec.CommandContext(ctx, cmd, args...)
	this.stop = stop

	stdin, err := this.cmd.StdinPipe()
	if err != nil { Log.Info(err.Error()); return false }
	this.stdin = stdin

	stdout, err := this.cmd.StdoutPipe()
	if err != nil { Log.Info(err.Error()); return false }
	this.stdout = stdout

	errorsChannel := make(chan error)

	// starting lsp cmd async
	go func() {
		startError := this.cmd.Run()
		if startError != nil { errorsChannel <- startError }
		close(errorsChannel)
	}()

	// wait for start
	var end = false
	for !end {
		select {
		case startError := <- errorsChannel:
			if startError != nil {
				Log.Error("error starting dap." + startError.Error())
				this.isStopped = true
				return false
			}

		case <-time.After(time.Duration(100) * time.Millisecond):
			Log.Info("dap started.", cmd, strings.Join(args," "))
			end = true
		}
	}

	if this.conntype == "tcp" {
		conn, err := net.Dial("tcp", fmt.Sprintf("localhost:%d", this.port))
		if err != nil {
			Log.Error("Error connecting:", err.Error())
			return false
		}

		this.reader = bufio.NewReader(conn)
		this.conn = conn
	}

	if this.conntype == "stdio" {
		this.reader = bufio.NewReader(stdout)
	}

	this.Breakpoints = make(map[string][]int)

	this.message2chan = make(map[int]chan string)
	this.otherMessages = make(chan string)
	this.StdoutMessages = make(chan string)
	this.EventMessages = make(chan string)
	this.IsReady = true

	go this.readStdout()
	go this.readTcp()

	return true
}

func (this *DapClient) Stop() {
	if this.isStopped { return }
	this.stop()
	this.isStopped = true
}


func (this *DapClient) readStdout() {
	scanner := bufio.NewScanner(this.stdout)
	for scanner.Scan() {
		message := scanner.Text()
		this.StdoutMessages <- message
	}
	close(this.StdoutMessages)
}


func (this *DapClient) send(o interface{}) {
	m, err := json.Marshal(o)
	if err != nil { panic(err) }

	Log.Info("->", string(m))
	message := fmt.Sprintf("Content-Length: %d\r\n\r\n%s", len(m), m)
	_, err = this.conn.Write([]byte(message))
	if err != nil { Log.Error(err.Error()) }
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
			if err != nil { Log.Error(err.Error()); continue }
			line = string(buf)
			messageSize = 0

			responseJSON := make(map[string]interface{})
			err = json.Unmarshal(buf, &responseJSON)
			if err != nil { Log.Error(err.Error()); continue }

			return line

		} else {
			line, err = this.reader.ReadString('\n') // it stuck sometimes
			if err != nil { Log.Error(err.Error()); continue }
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

func (this *DapClient) readTcp() {
	for {
		message := this.receive()
		Log.Info("<-", message)

		responseJson := make(map[string]interface{})
		err := json.Unmarshal([]byte(message), &responseJson)
		if err != nil { Log.Error(err.Error()); continue }

		if _, found := responseJson["event"]; found { // json has event
			this.EventMessages <- message
			continue
		}

		if value, found := responseJson["request_seq"]; found { // json has id
			if id, ok := value.(float64); ok {
				channel, foundRequest := this.message2chan[int(id)]
				if foundRequest {
					channel <- message
				} else  {
					//skip message
				}
			}
		}
	}
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

func (this *DapClient) Init(dir string) {
	this.id = 1
	id := this.id

	initializeRequest := InitializeRequest{
		Seq:     id,
		Type:    "request",
		Command: "initialize",
		Arguments: Arguments{
			ClientID:                   "edgo",
			ClientName:                 "edgo",
			AdapterID:                  this.Lang,
			Locale:                     "en",
			LinesStartAt1:              true,
			ColumnsStartAt1:            true,
			PathFormat:                 "path",
			SupportsVariableType:       true,
			SupportsVariablePaging:     true,
			SupportsRunInTerminalRequest: true,
			SupportsMemoryReferences:   true,
			SupportsProgressReporting:  true,
			SupportsInvalidatedEvent:   true,
			SupportsMemoryEvent:        true,
		},
	}


	this.message2chan[id] = this.otherMessages
	this.send(initializeRequest)

	response, err := WaitForRequest[Response](this.otherMessages, 3000)

	delete(this.message2chan, id)

	if err != nil || response.Success != true {
		Log.Error("initialize response from dap server error")
		this.IsReady = false
		return
	}
	
	this.IsReady = true
}

func (this *DapClient) Launch(program string) bool {
	
	this.id++
	id := this.id
		
	launchRequest := LaunchRequest{
		Seq:     id,
		Type:    "request",
		Command: "launch",
		Arguments: LaunchRequestArguments{
			Name:    "Launch debug " + program,
			Type:    this.Lang,
			Request: "launch",
			Mode:    "debug",
			Program: program,
		},
	}
	
	this.message2chan[id] = this.otherMessages
	this.send(launchRequest)

	response, err := WaitForRequest[LaunchResponse](this.otherMessages, 3000)

	delete(this.message2chan, id)

	if err != nil || response.Success != true {
		Log.Error("launch response from dap server error")
		this.IsReady = false
		return false
	}

	return true
}

func (this *DapClient) SetBreakpoint(file string, line int) bool {
	
	this.id++
	id := this.id

	name := filepath.Base(file)

	bps := this.Breakpoints[file]
	bps = append(bps, line)
	this.Breakpoints[file] = bps

	request := SetBreakpointsRequest{
		Seq: id, Type:    "request", Command: "setBreakpoints",
		Arguments: SetBreakpointsArguments{
			Source: Source{ Name: name,  Path: file },
			Breakpoints: []Breakpoint{
				{Line: line},
			},
			Lines: []int{line},
		},
	}
	
	this.message2chan[id] = this.otherMessages
	this.send(request)

	response, err := WaitForRequest[SetBreakpointsResponse](this.otherMessages, 3000)

	delete(this.message2chan, id)

	if err != nil || response.Success != true {
		Log.Error("launch response from dap server error")
		this.IsReady = false
	}

	return response.Success
}

func (this *DapClient) Continue(threadID int) bool {
	
	this.id++
	id := this.id
	
	request := ContinueRequest{
		Seq: id, Type: "request", Command: "continue",
		Arguments: ContinueArguments{ threadID },
	}
	
	this.message2chan[id] = this.otherMessages
	this.send(request)

	response, err := WaitForRequest[ContinueResponse](this.otherMessages, 3000)

	delete(this.message2chan, id)

	if err != nil || response.Success != true {
		Log.Error("launch response from dap server error")
		this.IsReady = false
	}

	return response.Success
}

func (this *DapClient) Next(threadID int) bool {

	this.id++
	id := this.id

	request := ContinueRequest{
		Seq: id, Type: "request", Command: "next",
		Arguments: ContinueArguments{ threadID },
	}

	this.message2chan[id] = this.otherMessages
	this.send(request)

	response, err := WaitForRequest[ContinueResponse](this.otherMessages, 3000)

	delete(this.message2chan, id)

	if err != nil || response.Success != true {
		Log.Error("launch response from dap server error")
		this.IsReady = false
	}

	return response.Success
}
func (this *DapClient) StepIn(threadID int) bool {

	this.id++
	id := this.id

	request := ContinueRequest{
		Seq: id, Type: "request", Command: "stepIn",
		Arguments: ContinueArguments{ threadID },
	}

	this.message2chan[id] = this.otherMessages
	this.send(request)

	response, err := WaitForRequest[ContinueResponse](this.otherMessages, 3000)

	delete(this.message2chan, id)

	if err != nil || response.Success != true {
		Log.Error("launch response from dap server error")
		this.IsReady = false
	}

	return response.Success
}

