package main

var logger = Logger{}
var editor = Editor{}
var lsp = LspClient{}
var highlighter = Highlighter{}

func main() {
	logger.start()
	config := GetConfig()
	editor.config = config
	highlighter.setTheme(config.Theme)
	editor.start()
}