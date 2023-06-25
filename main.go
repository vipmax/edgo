package main

var Log = Logger{}
var editor = Editor{}
var Lsp = LspClient{}
var Highlight = Highlighter{}

func main() {
	Log.Start()
	config := GetConfig()
	editor.config = config
	Highlight.SetTheme(config.Theme)
	editor.Start()
}