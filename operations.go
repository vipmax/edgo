package main

type Action string

const (
	Insert Action = "insert"
	Delete Action = "delete"
	Enter  Action = "enter"
	DeleteLine Action = "deleteline"
)

type Operation struct {
	action Action
	char   rune
	line   int
	column int
}