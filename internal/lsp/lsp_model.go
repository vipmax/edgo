package lsp

import (
	"github.com/goccy/go-json"
)

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
	Capabilities          Capabilities      `json:"capabilities,omitempty"`
	WorkDoneToken         string            `json:"workDoneToken,omitempty"`
}

type InitializeRequest struct {
	ID      int              `json:"id"`
	JSONRPC string           `json:"jsonrpc"`
	Method  string           `json:"method"`
	Params  InitializeParams `json:"params"`
}

type Position struct {
	Line      int `json:"line"`
	Character int `json:"character"`
}


type Context struct {
	IncludeDeclaration bool `json:"includeDeclaration,omitempty"`
	Only        []string `json:"only,omitempty"`
	TriggerKind int `json:"triggerKind,omitempty"`
}

type Params struct {
	TextDocument TextDocument `json:"textDocument"`
	Position     Position     `json:"position,omitempty"`
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
type ShutdownRequest struct {
	ID      int    `json:"id,omitempty"`
	JSONRPC string      `json:"jsonrpc"`
	Method  string      `json:"method"`
}
type ExitRequest struct {
	JSONRPC string      `json:"jsonrpc"`
	Method  string      `json:"method"`
}

type BaseRequest struct {
	ID      int    `json:"id,omitempty"`
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



type CompletionResponse struct {
	JSONRPC string            `json:"jsonrpc"`
	Result  CompletionResult  `json:"result"`
	ID      float64           `json:"id"`
}

type CompletionResponse2 struct {
	JSONRPC string            `json:"jsonrpc"`
	Items   []CompletionItem  `json:"result"`
	ID      float64           `json:"id"`
}

func unmarshal[T any](source []byte) (T, error) {
	var target T

	if err := json.Unmarshal(source, &target); err != nil {
		return target, err
	}

	return target,  nil
}

func (m *CompletionResponse) UnmarshalJSON(b []byte) error {

	type TempCompletionResponse CompletionResponse
	temp := &TempCompletionResponse{}

	// Try to unmarshal into CompletionResponse
	err := json.Unmarshal(b, temp)
	if err != nil {
		// If error, try to unmarshal into CompletionResponse2
		temp2 := &CompletionResponse2{}
		err2 := json.Unmarshal(b, temp2)
		if err2 != nil { return err }

		m.ID = temp2.ID
		m.JSONRPC = temp2.JSONRPC
		m.Result = CompletionResult{ false, temp2.Items}
	} else {
		*m = CompletionResponse(*temp)
	}

	return nil
}

type CompletionResult struct {
	IsIncomplete bool             `json:"isIncomplete"`
	Items        []CompletionItem `json:"items"`
}

type CompletionItem struct {
	Label            string        `json:"label"`
	Kind             float64       `json:"kind"`
	Detail           string        `json:"detail"`
	Preselect        bool          `json:"preselect"`
	SortText         string        `json:"sortText"`
	InsertText       string        `json:"insertText"`
	FilterText       string        `json:"filterText"`
	InsertTextFormat float64       `json:"insertTextFormat"`
	TextEdit         TextEdit      `json:"textEdit"`
}

type TextEdit struct {
	Range   Range  `json:"range"`
	Replace Range  `json:"replace"`
	Insert  Range  `json:"insert"`
	NewText string `json:"newText"`
}

type Range struct {
	Start PositionResponse `json:"start"`
	End   PositionResponse `json:"end"`
}

type PositionResponse struct {
	Line      float64 `json:"line"`
	Character float64 `json:"character"`
}


type Contents struct {
	Kind  string `json:"kind"`
	Value string `json:"value"`
}

type HoverResult struct {
	Contents Contents `json:"contents"`
	Range    Range    `json:"range"`
}

type HoverResponse struct {
	JSONRPC string      `json:"jsonrpc"`
	Result  HoverResult `json:"result"`
	ID      int         `json:"id"`
}


type Parameter struct {
	Label string `json:"label"`
}

type Signature struct {
	Label      string      `json:"label"`
	Parameters []Parameter `json:"parameters"`
}

type SignatureHelpResult struct {
	Signatures      []Signature `json:"signatures"`
	ActiveParameter int         `json:"activeParameter"`
}

type SignatureHelpResponse struct {
	JSONRPC string `json:"jsonrpc"`
	Result  SignatureHelpResult `json:"result"`
	ID      int    `json:"id"`
}


type Capabilities struct {
	CapabilitiesTextDocument CapabilitiesTextDocument `json:"textDocument"`
}

type CapabilitiesTextDocument struct {
	Hover              Hover              `json:"hover"`
	PublishDiagnostics PublishDiagnostics `json:"publishDiagnostics"`
	SignatureHelp      SignatureHelp      `json:"signatureHelp"`
	Completion         Completion         `json:"completion"`
}

type Hover struct {
	ContentFormat []string `json:"contentFormat"`
}

type PublishDiagnostics struct {
	RelatedInformation     bool `json:"relatedInformation"`
	VersionSupport         bool `json:"versionSupport"`
	CodeDescriptionSupport bool `json:"codeDescriptionSupport"`
	DataSupport            bool `json:"dataSupport"`
}

type SignatureHelp struct {
	SignatureInformation SignatureInformation `json:"signatureInformation"`
}

type SignatureInformation struct {
	DocumentationFormat []string `json:"documentationFormat"`
}

type Completion struct {
	CapabilitiesCompletionItem CapabilitiesCompletionItem `json:"completionItem"`
}

type CapabilitiesCompletionItem struct {
	ResolveProvider     bool               `json:"resolveProvider"`
	SnippetSupport      bool               `json:"snippetSupport"`
	InsertReplaceSupport bool              `json:"insertReplaceSupport"`
	LabelDetailsSupport bool               `json:"labelDetailsSupport"`
	ResolveSupport      ResolveSupport     `json:"resolveSupport"`
}

type ResolveSupport struct {
	Properties []string `json:"properties"`
}

var capabilities = Capabilities{
	CapabilitiesTextDocument: CapabilitiesTextDocument{
		Hover: Hover{
			ContentFormat: []string{"plaintext", "markdown"},
		},
		PublishDiagnostics: PublishDiagnostics{
			RelatedInformation:     false,
			VersionSupport:         false,
			CodeDescriptionSupport: true,
			DataSupport:            true,
		},
		SignatureHelp: SignatureHelp{
			SignatureInformation: SignatureInformation{
				DocumentationFormat: []string{"plaintext", "markdown"},
			},
		},
		Completion: Completion{
			CapabilitiesCompletionItem: CapabilitiesCompletionItem{
				ResolveProvider:      true,
				//SnippetSupport:       true,
				InsertReplaceSupport: true,
				LabelDetailsSupport:  true,
				ResolveSupport: ResolveSupport{
					Properties: []string{"documentation", "detail", "additionalTextEdits"},
				},
			},
		},
	},
}


type LspSettings struct {
	Langs []map[string]string `yaml:"langs"`
}

type CodeDescription struct {
	Href string `json:"href"`
}

type Diagnostic struct {
	Range            Range            `json:"range"`
	Severity         int              `json:"severity"`
	Code             interface{}      `json:"code"`
	CodeDescription  CodeDescription  `json:"codeDescription"`
	Source           string           `json:"source"`
	Message          string           `json:"message"`
}

type DiagnosticParams struct {
	Uri         string        `json:"uri"`
	Version     int           `json:"version"`
	Diagnostics []Diagnostic  `json:"diagnostics"`
}

type DiagnosticResponse struct {
	Jsonrpc string `json:"jsonrpc"`
	Method  string `json:"method"`
	Params  DiagnosticParams `json:"params"`
}


type DefinitionRequest struct {
	ID      int    `json:"id"`
	JSONRPC string `json:"jsonrpc"`
	Method  string `json:"method"`
	Params  DefinitionParams `json:"params"`
}

type DefinitionParams struct {
	TextDocument TextDocument `json:"textDocument"`
	Position     Position     `json:"position,omitempty"`
}


type DefinitionResult struct {
	URI   string `json:"uri"`
	Range Range  `json:"range"`
}

type DefinitionResponse struct {
	JSONRPC string   `json:"jsonrpc"`
	Result  []DefinitionResult `json:"result"`
	ID      int      `json:"id"`
}
type DefinitionResponse2 struct {
	JSONRPC string           `json:"jsonrpc"`
	Result  DefinitionResult `json:"result"`
	ID      int              `json:"id"`
}

func (m *DefinitionResponse) UnmarshalJSON(b []byte) error {
	type TempDefinitionResponse DefinitionResponse
	temp := &TempDefinitionResponse{}

	// Try to unmarshal into DefinitionResponse
	err := json.Unmarshal(b, temp)
	if err != nil {
		// If error, try to unmarshal into DefinitionResponse2
		temp2 := &DefinitionResponse2{}
		err2 := json.Unmarshal(b, temp2)
		if err2 != nil {
			return err
		}
		// If unmarshal into DefinitionResponse2 is successful, map fields to DefinitionResponse
		m.ID = temp2.ID
		m.JSONRPC = temp2.JSONRPC
		m.Result = []DefinitionResult{temp2.Result}
	} else {
		*m = DefinitionResponse(*temp)
	}

	return nil
}


type Character struct {
	Character int `json:"character"`
	Line      int `json:"line"`
}

type ChangeRange struct {
	Start Character `json:"start"`
	End   Character `json:"end"`
}

type ContentChange struct {
	Range ChangeRange  `json:"range"`
	Text  string `json:"text"`
	RangeLength int   `json:"rangeLength"`
}

type DidChangeParams struct {
	ContentChanges []ContentChange `json:"contentChanges"`
	TextDocument   TextDocument    `json:"textDocument"`
}

type DidChangeRequest struct {
	Jsonrpc string `json:"jsonrpc"`
	Method  string `json:"method"`
	Params  DidChangeParams `json:"params"`
}

type DidSaveParams struct {
	TextDocument   TextDocument    `json:"textDocument"`
}

type DidSaveRequest struct {
	Jsonrpc string `json:"jsonrpc"`
	Method  string `json:"method"`
	Params  DidSaveParams `json:"params"`
}

type ReferencesResponse struct {
	JSONRPC string  `json:"jsonrpc"`
	Result  []ReferencesRange `json:"result"`
	ID      int     `json:"id"`
}

type ReferencesRange struct {
	URI   string `json:"uri"`
	Range Span   `json:"range"`
}

type Span struct {
	Start Position `json:"start"`
	End   Position `json:"end"`
}



type PrepareRenameRequest struct {
	ID int `json:"id"`
	Jsonrpc string `json:"jsonrpc"`
	Method  string `json:"method"`
	Params  Params `json:"params"`
}


type PrepareRenameResponse struct {
	ID int `json:"id"`
	Jsonrpc string `json:"jsonrpc"`
	Result  struct {
		Range Range `json:"range"`
		Placeholder string `json:"placeholder"`
	} `json:"result"`
}


type RenameParams struct {
	NewName string `json:"newName"`
	Position Position `json:"position"`
	TextDocument TextDocument `json:"textDocument"`
}

type RenameRequest struct {
	ID int `json:"id"`
	Jsonrpc string `json:"jsonrpc"`
	Method  string `json:"method"`
	Params  RenameParams `json:"params"`
}

type RenameResponse struct {
	Jsonrpc string        `json:"jsonrpc"`
	Result  ChangesResult `json:"result"`
	ID      int           `json:"id"`
}

type ChangesResult struct {
	DocumentChanges []DocumentChange `json:"documentChanges"`
}

type DocumentChange struct {
	TextDocument TextDocument `json:"textDocument"`
	Edits        []Edit        `json:"edits"`
}


type Edit struct {
	Range   Range  `json:"range"`
	NewText string `json:"newText"`
}


type RequestRange struct {
	Start Position `json:"start"`
	End   Position `json:"end"`
}

type CodeActionParams struct {
	TextDocument TextDocument `json:"textDocument"`
	Range        RequestRange        `json:"range"`
	Context      Context      `json:"context"`
}

type CodeActionRequest struct {
	ID int `json:"id"`
	Jsonrpc string `json:"jsonrpc"`
	Method  string `json:"method"`
	Params  CodeActionParams `json:"params"`
}

type CodeActionResponse struct {
	Jsonrpc string  `json:"jsonrpc"`
	Result  []CodeActionResult `json:"result"`
	ID      int     `json:"id"`
}

type CodeActionResult struct {
	Title   string  `json:"title"`
	Kind    string  `json:"kind"`
	Edit    struct{} `json:"edit"`
	Command Command `json:"command"`
}

type Command struct {
	//Title     string     `json:"title"`
	Command   string     `json:"command"`
	Arguments []Argument `json:"arguments"`
}

type Argument struct {
	Fix   string `json:"Fix"`
	URI   string `json:"URI"`
	Range Range  `json:"Range"`
}

type CommandRequest struct {
	ID int `json:"id"`
	Jsonrpc string `json:"jsonrpc"`
	Method  string `json:"method"`
	Params  Command `json:"params"`
}


//type CommandResponse struct {
//	ID      int    `json:"id"`
//	Jsonrpc string `json:"jsonrpc"`
//	Method  string `json:"method"`
//	Params  EditParams `json:"params"`
//}
//
//type EditParams struct {
//	Edit DocumentEdit `json:"edit"`
//}
//
//type DocumentEdit struct {
//	DocumentChanges ChangesResult `json:"documentChanges"`
//}

type DocChange struct {
	TextDocument struct {
		Version int    `json:"version"`
		URI     string `json:"uri"`
	} `json:"textDocument"`
	Edits []Edit `json:"edits"`
}

type CommandResponse struct {
	Jsonrpc string `json:"jsonrpc"`
	Method  string `json:"method"`
	Params  struct {
		Edit struct {
			DocumentChanges []DocChange `json:"documentChanges"`
		} `json:"edit"`
	} `json:"params"`
	ID int `json:"id"`
}


type Applied struct {
	Applied bool  `json:"applied"`
}
type ApplyWorkspaceEditParams struct {
	Applied Applied  `json:"edit"`
}

type ApplyEditRequest struct {
	ID int `json:"id"`
	Jsonrpc string `json:"jsonrpc"`
	Result  Applied `json:"result"`
}