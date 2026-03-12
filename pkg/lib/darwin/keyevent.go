package darwin

/*
#cgo LDFLAGS: -framework ApplicationServices
#include <keyevent.h>
*/
import "C"

import (
	"sync"
)

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

var (
	keyHandler   func(KeyEvent) bool // returns true to consume the event
	keyHandlerMu sync.RWMutex
)

// SetKeyHandler registers a function that is called synchronously for every
// key event. Return true to consume (swallow) the event, false to pass it
// through to other applications.
func SetKeyHandler(handler func(KeyEvent) bool) {
	keyHandlerMu.Lock()
	defer keyHandlerMu.Unlock()
	keyHandler = handler
}

// cgo directive
//
//export keyEventCallback
func keyEventCallback(keycode C.longlong, flags C.longlong, eventType C.longlong) C.int {
	ev := KeyEvent{
		Keycode:   uint16(keycode),
		Flags:     uint64(flags),
		EventType: int(eventType),
	}

	keyHandlerMu.RLock()
	handler := keyHandler
	keyHandlerMu.RUnlock()

	if handler != nil && handler(ev) {
		return 1 // consume
	}
	return 0 // pass through
}

func StartEventTap() {
	C.startEventTap()
}
