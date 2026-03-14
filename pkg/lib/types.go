package lib

type KeyEvent struct {
	Keycode   uint16
	Flags     uint64
	EventType int
}

const (
	EventKeyDown      = 10
	EventKeyUp        = 11
	EventFlagsChanged = 12
)

type CKeyMsg struct {
	Event KeyEvent
}
