package main

import (
	. "edgo/internal/config"
	. "edgo/internal/editor"
	. "edgo/internal/highlighter"
	. "edgo/internal/logger"
)

func main() {
	Log.Start()
	Conf := GetConfig()
	HighlighterGlobal.SetTheme(Conf.Theme)
	editor := Editor{}
	editor.Config = Conf

	//defer func() {
	//	if r := recover(); r != nil {
	//		editor.Exit()
	//		errMsg := fmt.Sprintf("Recovered from panic. Error: %v\n", r)
	//		stackTrace := make([]byte, 4096)
	//		stackSize := runtime.Stack(stackTrace, false)
	//		fmt.Printf("%s\nStack Trace:\n%s\n", errMsg, stackTrace[:stackSize])
	//	}
	//}()

	editor.Start()
}
