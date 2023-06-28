package main

import . "edgo/internal/highlighter"
import . "edgo/internal/logger"
import . "edgo/internal/editor"
import . "edgo/internal/config"

func main() {
	Log.Start()
	Conf := GetConfig()
	HighlighterGlobal.SetTheme(Conf.Theme)
	editor := Editor{}
	editor.Config = Conf
	editor.Start()
}