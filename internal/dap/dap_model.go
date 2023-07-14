package dap

type Arguments struct {
	ClientID                     string `json:"clientID"`
	ClientName                   string `json:"clientName"`
	AdapterID                    string `json:"adapterID"`
	Locale                       string `json:"locale"`
	LinesStartAt1                bool   `json:"linesStartAt1"`
	ColumnsStartAt1              bool   `json:"columnsStartAt1"`
	PathFormat                   string `json:"pathFormat"`
	SupportsVariableType         bool   `json:"supportsVariableType,omitempty"`
	SupportsVariablePaging       bool   `json:"supportsVariablePaging,omitempty"`
	SupportsRunInTerminalRequest bool   `json:"supportsRunInTerminalRequest,omitempty"`
	SupportsMemoryReferences     bool   `json:"supportsMemoryReferences,omitempty"`
	SupportsProgressReporting    bool   `json:"supportsProgressReporting,omitempty"`
	SupportsInvalidatedEvent     bool   `json:"supportsInvalidatedEvent,omitempty"`
	SupportsMemoryEvent          bool   `json:"supportsMemoryEvent,omitempty"`
}

type InitializeRequest struct {
	Seq        int       `json:"seq"`
	RequestSeq int       `json:"request_seq"`
	Type       string    `json:"type"`
	Command    string    `json:"command"`
	Arguments  Arguments `json:"arguments"`
}

type Response struct {
	Seq     int    `json:"seq"`
	Type    string `json:"type"`
	Request int    `json:"request_seq"`
	Success bool   `json:"success"`
	Command string `json:"command"`
	Body    Body   `json:"body"`
}

type Body struct {
	SupportsConfigurationDoneRequest bool `json:"supportsConfigurationDoneRequest"`
	SupportsFunctionBreakpoints      bool `json:"supportsFunctionBreakpoints"`
	SupportsConditionalBreakpoints   bool `json:"supportsConditionalBreakpoints"`
	SupportsEvaluateForHovers        bool `json:"supportsEvaluateForHovers"`
	SupportsSetVariable              bool `json:"supportsSetVariable"`
	SupportsExceptionInfoRequest     bool `json:"supportsExceptionInfoRequest"`
	SupportsDelayedStackTraceLoading bool `json:"supportsDelayedStackTraceLoading"`
	SupportsLogPoints                bool `json:"supportsLogPoints"`
	SupportsDisassembleRequest       bool `json:"supportsDisassembleRequest"`
	SupportsClipboardContext         bool `json:"supportsClipboardContext"`
	SupportsSteppingGranularity      bool `json:"supportsSteppingGranularity"`
	SupportsInstructionBreakpoints   bool `json:"supportsInstructionBreakpoints"`
}

type LaunchRequest struct {
	Seq       int                    `json:"seq"`
	Type      string                 `json:"type"`
	Command   string                 `json:"command"`
	Arguments LaunchRequestArguments `json:"arguments"`
}

type LaunchRequestArguments struct {
	Name    string `json:"name"`
	Type    string `json:"type"`
	Request string `json:"request"`
	Mode    string `json:"mode"`
	Program string `json:"program"`
}

type LaunchResponse struct {
	Seq        int    `json:"seq"`
	Type       string `json:"type"`
	RequestSeq int    `json:"request_seq"`
	Success    bool   `json:"success"`
	Command    string `json:"command"`
}

type SetBreakpointsRequest struct {
	Seq       int                     `json:"seq"`
	Type      string                  `json:"type"`
	Command   string                  `json:"command"`
	Arguments SetBreakpointsArguments `json:"arguments"`
}

type SetBreakpointsArguments struct {
	Source      Source       `json:"source"`
	Breakpoints []Breakpoint `json:"breakpoints"`
	Lines       []int        `json:"lines"`
}

type Source struct {
	Name string `json:"name"`
	Path string `json:"path"`
}

type Breakpoint struct {
	Line int `json:"line"`
}

type SetBreakpointsResponse struct {
	Seq          int          `json:"seq"`
	Type         string       `json:"type"`
	RequestSeq   int          `json:"request_seq"`
	Success      bool         `json:"success"`
	Command      string       `json:"command"`
	ResponseBody ResponseBody `json:"body"`
}

type ResponseBody struct {
	Breakpoints []BreakpointResponse `json:"breakpoints"`
}

type BreakpointResponse struct {
	ID       int    `json:"id"`
	Verified bool   `json:"verified"`
	Source   Source `json:"source"`
	Line     int    `json:"line"`
}

type ContinueRequest struct {
	Seq       int               `json:"seq"`
	Type      string            `json:"type"`
	Command   string            `json:"command"`
	Arguments ContinueArguments `json:"arguments"`
}

type ContinueArguments struct {
	ThreadID int `json:"threadId"`
}

type ContinueResponse struct {
	Seq          int          `json:"seq"`
	Type         string       `json:"type"`
	RequestSeq   int          `json:"request_seq"`
	Success      bool         `json:"success"`
	Command      string       `json:"command"`
	ResponseBody ContinueBody `json:"body"`
}

type ContinueBody struct {
	AllThreadsContinued bool `json:"allThreadsContinued"`
}

type StoppedEvent struct {
	Seq          int         `json:"seq"`
	Type         string      `json:"type"`
	Event        string      `json:"event"`
	ResponseBody StoppedBody `json:"body"`
}

type StoppedBody struct {
	Reason            string `json:"reason"`
	ThreadID          int    `json:"threadId"`
	AllThreadsStopped bool   `json:"allThreadsStopped"`
	HitBreakpointIDs  []int  `json:"hitBreakpointIds"`
}

type ThreadsResponse struct {
	Seq          int         `json:"seq"`
	Type         string      `json:"type"`
	RequestSeq   int         `json:"request_seq"`
	Success      bool        `json:"success"`
	Command      string      `json:"command"`
	ResponseBody ThreadsBody `json:"body"`
}

type ThreadsBody struct {
	Threads []Thread `json:"threads"`
}

type Thread struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

type StacktraceRequest struct {
	Seq       int                 `json:"seq"`
	Type      string              `json:"type"`
	Command   string              `json:"command"`
	Arguments StacktraceArguments `json:"arguments"`
}

type StacktraceArguments struct {
	ThreadID int `json:"threadId"`
	Levels   int `json:"levels"`
}

type StackTraceResponse struct {
	Seq          int            `json:"seq"`
	Type         string         `json:"type"`
	RequestSeq   int            `json:"request_seq"`
	Success      bool           `json:"success"`
	Command      string         `json:"command"`
	ResponseBody StackTraceBody `json:"body"`
}

type StackTraceBody struct {
	StackFrames []StackFrame `json:"stackFrames"`
	TotalFrames int          `json:"totalFrames"`
}

type StackFrame struct {
	ID                    int    `json:"id"`
	Name                  string `json:"name"`
	Source                Source `json:"source"`
	Line                  int    `json:"line"`
	Column                int    `json:"column"`
	InstructionPointerRef string `json:"instructionPointerReference"`
}

type ScopeRequest struct {
	Seq       int            `json:"seq"`
	Type      string         `json:"type"`
	Command   string         `json:"command"`
	Arguments ScopeArguments `json:"arguments"`
}

type ScopeArguments struct {
	FramedId int `json:"frameId"`
}

type ScopesResponse struct {
	Seq          int        `json:"seq"`
	Type         string     `json:"type"`
	RequestSeq   int        `json:"request_seq"`
	Success      bool       `json:"success"`
	Command      string     `json:"command"`
	ResponseBody ScopesBody `json:"body"`
}

type ScopesBody struct {
	Scopes []Scope `json:"scopes"`
}

type Scope struct {
	Name               string `json:"name"`
	VariablesReference int    `json:"variablesReference"`
	Expensive          bool   `json:"expensive"`
	Source             Source `json:"source"`
}

type VariablesRequest struct {
	Seq       int                `json:"seq"`
	Type      string             `json:"type"`
	Command   string             `json:"command"`
	Arguments VariablesArguments `json:"arguments"`
}

type VariablesArguments struct {
	VariablesReference int `json:"variablesReference"`
}

type VariablesResponse struct {
	Seq          int           `json:"seq"`
	Type         string        `json:"type"`
	RequestSeq   int           `json:"request_seq"`
	Success      bool          `json:"success"`
	Command      string        `json:"command"`
	ResponseBody VariablesBody `json:"body"`
}

type VariablesBody struct {
	Variables []Variable `json:"variables"`
}

type Variable struct {
	Name               string           `json:"name"`
	Value              string           `json:"value"`
	Type               string           `json:"type"`
	PresentationHint   PresentationHint `json:"presentationHint"`
	EvaluateName       string           `json:"evaluateName"`
	VariablesReference int              `json:"variablesReference"`
}

type PresentationHint struct {
	// Add relevant fields if available
}

type AttachRequest struct {
	Seq       int             `json:"seq"`
	Type      string          `json:"type"`
	Command   string          `json:"command"`
	Arguments AttachArguments `json:"arguments"`
}

type AttachArguments struct {
	Restart bool `json:"restart"`
}
