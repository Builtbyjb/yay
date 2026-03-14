package core

import "database/sql"

type App struct {
	Name    string
	BinName string
	Path    string
}

type Setting struct {
	Id      int
	Name    string
	BinName string
	Path    string
	HotKey  sql.NullString
	Mode    string
	Enabled bool
}

type Update struct {
	Id      int
	Hotkey  string
	Mode    string
	Enabled bool
}
