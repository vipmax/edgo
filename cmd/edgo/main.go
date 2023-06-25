package main

import . "edgo/internal/highlighter"
import . "edgo/internal/logger"
import . "edgo/internal/editor"
import . "edgo/internal/config"



func main() {
	Log.Start()
	config := GetConfig()
	EditorGlobal.Config = config
	HighlighterGlobal.SetTheme(config.Theme)
	EditorGlobal.Start()
}