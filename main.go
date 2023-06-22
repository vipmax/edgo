package main

var logger = Logger{}
var editor = Editor{}
var lsp = LspClient{}
var highlighter = Highlighter{}

func main() {

	config := GetConfig()
	logger.start()

	editor.config = config
	highlighter.setTheme(config.Theme)

	editor.start()
}