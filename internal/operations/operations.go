package operations

type Action string

const (
	Insert Action = "insert"
	Delete Action = "delete"
	Enter  Action = "enter"
	DeleteLine Action = "deleteline"
	MoveCursor Action = "movecursor"
)

type Operation struct {
	Action Action
	Char   rune
	Line   int
	Column int
}

type EditOperation []Operation


type CursorMove struct {
	Filename string
	Row int
	Col int
	Y int
	X int
}
