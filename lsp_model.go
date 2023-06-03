package main

type ClientInfo struct {
	Name    string `json:"name,omitempty"`
	Version string `json:"version,omitempty"`
}
type WorkspaceFolder struct {
	Name string `json:"name"`
	URI  string `json:"uri"`
}

type InitializeParams struct {
	ProcessID             int               `json:"processId,omitempty"`
	RootPath              string            `json:"rootPath,omitempty"`
	RootURI               string            `json:"rootUri,omitempty"`
	WorkspaceFolders      []WorkspaceFolder `json:"workspaceFolders,omitempty"`
	ClientInfo            ClientInfo        `json:"clientInfo,omitempty"`
	Trace                 string            `json:"trace,omitempty"`
	InitializationOptions interface{}       `json:"initializationOptions,omitempty"`
	Capabilities          interface{}       `json:"capabilities"`
	WorkDoneToken         string            `json:"workDoneToken,omitempty"`
}

type InitializeRequest struct {
	ID     int              `json:"id"`
	Method string           `json:"method"`
	Params InitializeParams `json:"params"`
}

type Position struct {
	Line      int `json:"line"`
	Character int `json:"character"`
}

type Context struct {
	TriggerKind int `json:"triggerKind"`
}

type Params struct {
	TextDocument TextDocument `json:"textDocument"`
	Position     Position     `json:"position"`
	Context      Context      `json:"context,omitempty"`
}

type TextDocument struct {
	LanguageID string `json:"languageId,omitempty"`
	Text       string `json:"text,omitempty"`
	URI        string `json:"uri,omitempty"`
	Version    int    `json:"version,omitempty"`
}

type DidOpenParams struct {
	TextDocument TextDocument `json:"textDocument"`
}

type DidOpenRequest struct {
	JSONRPC string        `json:"jsonrpc"`
	Method  string        `json:"method"`
	Params  DidOpenParams `json:"params"`
}

type InitializedRequest struct {
	JSONRPC string      `json:"jsonrpc"`
	Method  string      `json:"method"`
	Params  interface{} `json:"params,omitempty"`
}

type BaseRequest struct {
	ID      int    `json:"id"`
	JSONRPC string `json:"jsonrpc"`
	Method  string `json:"method"`
	Params  Params `json:"params"`
}

type TextDocumentDidChangeParams struct {
	TextDocument   VersionedTextDocumentIdentifier  `json:"textDocument"`
	ContentChanges []TextDocumentContentChangeEvent `json:"contentChanges"`
}

type VersionedTextDocumentIdentifier struct {
	URI     string `json:"uri"`
	Version int    `json:"version"`
}

type TextDocumentContentChangeEvent struct {
	Text string `json:"text"`
}


type Item struct {
	Label             string         `json:"label"`
	Kind              int            `json:"kind"`
	Detail            string         `json:"detail"`
	Preselect         bool           `json:"preselect"`
	SortText          string         `json:"sortText"`
	FilterText        string         `json:"filterText"`
	InsertTextFormat  int            `json:"insertTextFormat"`
	TextEdit          TextEdit       `json:"textEdit"`
}

type TextEdit struct {
	Range   Range  `json:"range"`
	NewText string `json:"newText"`
}

type Range struct {
	Start Position `json:"start"`
	End   Position `json:"end"`
}


type Completion struct {
	IsIncomplete bool   `json:"isIncomplete"`
	Items        []Item `json:"items"`
}

