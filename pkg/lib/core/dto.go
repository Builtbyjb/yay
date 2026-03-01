package core

type App struct {
	Name     string
	Path     string
	IconPath string
}

type Setting struct {
	Id       int
	Name     string
	Path     string
	IconPath string
	HotKey   string
	Mode     string
	Enabled  bool
}

type Update struct {
	Id      int
	Hotkey  string
	Mode    string
	Enabled bool
}
