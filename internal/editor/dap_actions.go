package editor

import (
	dap "edgo/internal/dap"
	. "edgo/internal/logger"
	. "edgo/internal/utils"
	//"encoding/json"
	"fmt"
	. "github.com/gdamore/tcell"
	"github.com/goccy/go-json"
	"os"
	"sort"
	"strings"
	"time"
)

func (e *Editor) Breakpoint() {
	if e.Dap.Breakpoints == nil { e.Dap.Breakpoints = make(map[string][]int) }
	bps := e.Dap.Breakpoints[e.AbsoluteFilePath]
	line := e.Row + 1
	if !Contains(bps, line) {
		bps = append(bps, line)
	} else  {
		bps = FindAndRemove(bps, line)
	}
	sort.Ints(bps)
	e.Dap.Breakpoints[e.AbsoluteFilePath] = bps

	if e.Dap.IsStarted {
		e.Dap.SetBreakpoint(e.AbsoluteFilePath, -1)
	}
}

func (e *Editor) OnDebugStop() {
	e.Dap.Disconnect()
	e.Dap.Stop()
	e.DebugInfo.stopline = -1
	e.ProcessPanelHeight = 0
	e.ROWS = e.TERMINAL_HEIGHT

	go func() {
		time.Sleep(100*time.Millisecond)
		e.Dap = dap.DapClient{
			Lang: e.Dap.Lang,
			Conntype: "tcp",
			Port: e.Dap.Port + 1,
			Breakpoints: e.Dap.Breakpoints,
		}
	}()

}

func (e *Editor) OnDebug() {
	if (e.Lang == "" || e.langConf.Cmd == "") { return }

	if e.ProcessPanelHeight == 0 {
		e.ProcessPanelHeight = 10
		e.COLUMNS, e.ROWS = e.Screen.Size()
		e.ROWS -= e.ProcessPanelHeight
	}

	if !e.Dap.IsStarted {
		currentDir, _ := os.Getwd()
		e.ProcessPanelScroll = 0
		e.ProcessPanelSpacing = 2
		e.DebugInfo = DebugInfo{stopline: -1}
		e.ProcessContent = [][]rune{}

		var cmd = ""
		var runtype = "launch"

		if e.Lang == "go" {
			cmd = fmt.Sprintf("dlv dap --listen=127.0.0.1:%d --log=true --log-output=dap", e.Dap.Port)
		}

		if e.Lang == "python" {
			cmd = fmt.Sprintf("python3 -m debugpy --listen localhost:%d --wait-for-client %s", e.Dap.Port, e.AbsoluteFilePath)
			runtype = "attach"
			os.Setenv("PYTHONUNBUFFERED", "1")
		}

		split := strings.Split(cmd, " ")
		e.Dap.Start(split[0], split[1:]...)
		e.ProcessContent = append(e.ProcessContent, []rune(cmd))

		go e.processStdout()
		go e.processDapEvents()

		e.DrawEverything()
		e.Screen.Show()

		//time.Sleep(time.Second)
		e.Dap.Init(currentDir)
		//time.Sleep(time.Second)

		if runtype == "launch" {
			e.Dap.Launch(e.AbsoluteFilePath)
		}

		if runtype == "attach" {
			e.Dap.Attach(false)
		}

		//time.Sleep(time.Second)

		e.Dap.ConfigurationDone()
		e.Dap.SetAllBreakpoints()
		//e.Dap.Continue(1)
	}

	e.DrawEverything()
	e.Screen.Show()
}

func (e *Editor) processDapEvents() {
	for eventMessage := range e.Dap.EventMessages {
		stopped := dap.StoppedEvent{}
		err := json.Unmarshal([]byte(eventMessage), &stopped)
		if err != nil { Log.Error("Error parsing JSON:" + err.Error()); continue }

		if stopped.Event == "stopped" && stopped.ResponseBody.Reason == "breakpoint" {
			e.GetDebugInfo()

			if len(stopped.ResponseBody.HitBreakpointIDs) > 0 {
				id := stopped.ResponseBody.HitBreakpointIDs[0] - 1
				bps := e.Dap.Breakpoints[e.AbsoluteFilePath]
				if id >= len(bps) { continue }
				line := bps[id] - 1
				e.DebugInfo.stopline = line
			}

			e.DrawEverything()
			e.Screen.SetContent(e.FilesPanelWidth , e.DebugInfo.stopline, '▷', nil, StyleDefault)

			e.GetDebugInfo()
			e.DrawDebugPanel()
			e.Screen.Show()
		}
	}
}

func (e *Editor) processStdout() {
	for line := range e.Dap.StdoutMessages {

		e.ProcessContent = append(e.ProcessContent, []rune(line))

		if len(e.ProcessContent) > e.ProcessPanelHeight {
			if e.ProcessPanelScroll >= len(e.ProcessContent)-e.ProcessPanelHeight-1 {
				e.ProcessPanelScroll = len(e.ProcessContent) - e.ProcessPanelHeight + 1 // focusing
				e.ProcessPanelScroll = Max(0, e.ProcessPanelScroll)
			}
		}

		e.DrawProcessPanel()
		e.Screen.Show()

		if e.Dap.IsStopped {
			if len(e.ProcessContent) > e.ProcessPanelHeight { // focusing
				e.ProcessPanelScroll = len(e.ProcessContent) - e.ProcessPanelHeight + 1
			}
			e.DrawProcessPanel()
			e.Screen.Show()
			break
		}
	}
}

type DebugInfo struct {
	StackFrames     []dap.StackFrame
	Scopes          []dap.Scope
	Variables       []dap.Variable
	stopline        int
}

func (e *Editor) DrawDebugPanel() {
	// clean files panel
	for i := 0; i < e.ROWS; i++ {
		for j := 0; j < e.FilesPanelWidth-1; j++ {
			e.Screen.SetContent(j, i, ' ', nil, StyleDefault)
		}
	}

	for i, variable := range e.DebugInfo.Variables {
		label := []rune(variable.Name + " = " + variable.Value)
		if len(label) > e.FilesPanelWidth {label = label[:e.FilesPanelWidth] }

		for j := 0; j < len(label); j++ {
			e.Screen.SetContent(j, i, label[j], nil, StyleDefault)
		}
	}
}

func (e *Editor) GetDebugInfo()  {
	stackTraceResponse := e.Dap.Stacktrace(1, 1)

	if stackTraceResponse.ResponseBody.StackFrames == nil ||
		len(stackTraceResponse.ResponseBody.StackFrames) == 0 {
		return
	}

	e.DebugInfo.StackFrames = stackTraceResponse.ResponseBody.StackFrames
	frameid := stackTraceResponse.ResponseBody.StackFrames[0].ID

	e.DebugInfo.stopline = stackTraceResponse.ResponseBody.StackFrames[0].Line - 1

	scopesResponse := e.Dap.Scopes(frameid)
	e.DebugInfo.Scopes = scopesResponse.ResponseBody.Scopes

	if scopesResponse.ResponseBody.Scopes != nil && len(scopesResponse.ResponseBody.Scopes) > 0 {
		scope := scopesResponse.ResponseBody.Scopes[0]

		variablesResponse := e.Dap.Variables(scope.VariablesReference)
		e.DebugInfo.Variables = variablesResponse.ResponseBody.Variables
	}

}



func (e *Editor) OnDebugKeyHandle(key Key, ev *EventKey, threadId int) {
	keyrune := ev.Rune()

	if key == KeyRune && keyrune == 'c' {
		e.Dap.Continue(threadId)
		e.DebugInfo.stopline = -1
	}

	if key == KeyRune && keyrune == 'n' {
		e.Dap.Next(threadId)
	}
	if key == KeyRune && keyrune == 'b' {
		e.Breakpoint()
		e.DrawEverything()
		e.Screen.Show()
	}
	if key == KeyRune && keyrune == 's' {
		e.Dap.StepIn(threadId)
	}
	if key == KeyRune && keyrune == 'q' {
		e.OnDebugStop()
	}
	if key == KeyRune && keyrune == 'p' {
		e.Dap.Pause(threadId)
	}
}

