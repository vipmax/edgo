package dap

import (
	//"github.com/goccy/go-json"
)

type Arguments struct {
	ClientID                   string `json:"clientID"`
	ClientName                 string `json:"clientName"`
	AdapterID                  string `json:"adapterID"`
	Locale                     string `json:"locale"`
	LinesStartAt1              bool   `json:"linesStartAt1"`
	ColumnsStartAt1            bool   `json:"columnsStartAt1"`
	PathFormat                 string `json:"pathFormat"`
	SupportsVariableType       bool   `json:"supportsVariableType"`
	SupportsVariablePaging     bool   `json:"supportsVariablePaging"`
	SupportsRunInTerminalRequest bool   `json:"supportsRunInTerminalRequest"`
	SupportsMemoryReferences   bool   `json:"supportsMemoryReferences"`
	SupportsProgressReporting  bool   `json:"supportsProgressReporting"`
	SupportsInvalidatedEvent   bool   `json:"supportsInvalidatedEvent"`
	SupportsMemoryEvent        bool   `json:"supportsMemoryEvent"`
}

type InitializeRequest struct {
	Seq       int       `json:"seq"`
	Type      string    `json:"type"`
	Command   string    `json:"command"`
	Arguments Arguments `json:"arguments"`
}

type Response struct {
	Seq      int    `json:"seq"`
	Type     string `json:"type"`
	Request  int    `json:"request_seq"`
	Success  bool   `json:"success"`
	Command  string `json:"command"`
	Body     Body   `json:"body"`
}

type Body struct {
	SupportsConfigurationDoneRequest   bool `json:"supportsConfigurationDoneRequest"`
	SupportsFunctionBreakpoints        bool `json:"supportsFunctionBreakpoints"`
	SupportsConditionalBreakpoints     bool `json:"supportsConditionalBreakpoints"`
	SupportsEvaluateForHovers          bool `json:"supportsEvaluateForHovers"`
	SupportsSetVariable                bool `json:"supportsSetVariable"`
	SupportsExceptionInfoRequest       bool `json:"supportsExceptionInfoRequest"`
	SupportsDelayedStackTraceLoading   bool `json:"supportsDelayedStackTraceLoading"`
	SupportsLogPoints                  bool `json:"supportsLogPoints"`
	SupportsDisassembleRequest         bool `json:"supportsDisassembleRequest"`
	SupportsClipboardContext           bool `json:"supportsClipboardContext"`
	SupportsSteppingGranularity        bool `json:"supportsSteppingGranularity"`
	SupportsInstructionBreakpoints     bool `json:"supportsInstructionBreakpoints"`
}

type LaunchRequest struct {
	Seq       int                  `json:"seq"`
	Type      string               `json:"type"`
	Command   string               `json:"command"`
	Arguments LaunchRequestArguments `json:"arguments"`
}

type LaunchRequestArguments struct {
	Name     string `json:"name"`
	Type     string `json:"type"`
	Request  string `json:"request"`
	Mode     string `json:"mode"`
	Program  string `json:"program"`
}

type LaunchResponse struct {
	Seq          int    `json:"seq"`
	Type         string `json:"type"`
	RequestSeq   int    `json:"request_seq"`
	Success      bool   `json:"success"`
	Command      string `json:"command"`
}

type SetBreakpointsRequest struct {
	Seq       int                       `json:"seq"`
	Type      string                    `json:"type"`
	Command   string                    `json:"command"`
	Arguments SetBreakpointsArguments `json:"arguments"`
}

type SetBreakpointsArguments struct {
	Source      Source            `json:"source"`
	Breakpoints []Breakpoint      `json:"breakpoints"`
	Lines       []int             `json:"lines"`
}

type Source struct {
	Name string `json:"name"`
	Path string `json:"path"`
}

type Breakpoint struct {
	Line int `json:"line"`
}


type SetBreakpointsResponse struct {
	Seq         int               `json:"seq"`
	Type        string            `json:"type"`
	RequestSeq  int               `json:"request_seq"`
	Success     bool              `json:"success"`
	Command     string            `json:"command"`
	ResponseBody ResponseBody     `json:"body"`
}

type ResponseBody struct {
	Breakpoints []BreakpointResponse `json:"breakpoints"`
}

type BreakpointResponse struct {
	ID     int          `json:"id"`
	Verified bool       `json:"verified"`
	Source Source       `json:"source"`
	Line   int          `json:"line"`
}

type ContinueRequest struct {
	Seq       int                `json:"seq"`
	Type      string             `json:"type"`
	Command   string             `json:"command"`
	Arguments ContinueArguments `json:"arguments"`
}

type ContinueArguments struct {
	ThreadID int `json:"threadId"`
}

type ContinueResponse struct {
	Seq         int               `json:"seq"`
	Type        string            `json:"type"`
	RequestSeq  int               `json:"request_seq"`
	Success     bool              `json:"success"`
	Command     string            `json:"command"`
	ResponseBody ContinueBody      `json:"body"`
}

type ContinueBody struct {
	AllThreadsContinued bool      `json:"allThreadsContinued"`
}

type StoppedEvent struct {
	Seq         int               `json:"seq"`
	Type        string            `json:"type"`
	Event       string            `json:"event"`
	ResponseBody StoppedBody      `json:"body"`
}

type StoppedBody struct {
	Reason            string    `json:"reason"`
	ThreadID          int       `json:"threadId"`
	AllThreadsStopped bool      `json:"allThreadsStopped"`
	HitBreakpointIDs  []int     `json:"hitBreakpointIds"`
}


