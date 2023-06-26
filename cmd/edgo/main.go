package main

import . "edgo/internal/highlighter"
import . "edgo/internal/logger"
import . "edgo/internal/editor"
import . "edgo/internal/config"

func main() {
	Log.Start()
	Conf := GetConfig()
	EditorGlobal.Config = Conf
	HighlighterGlobal.SetTheme(Conf.Theme)
	EditorGlobal.Start()
}