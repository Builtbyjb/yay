package darwin

/*
#cgo LDFLAGS: -framework ApplicationServices
#include <ApplicationServices/ApplicationServices.h>

extern int keyEventCallback(long long keycode, long long flags, long long eventType);

static inline CGEventRef eventCallback(CGEventTapProxy proxy, CGEventType type, CGEventRef event, void *refcon) {
    if (type == kCGEventKeyDown || type == kCGEventKeyUp || type == kCGEventFlagsChanged) {
        long long keycode = (long long)CGEventGetIntegerValueField(event, kCGKeyboardEventKeycode);
        long long flags   = (long long)CGEventGetFlags(event);
        long long evType  = (long long)type;
        int consumed = keyEventCallback(keycode, flags, evType);
        if (consumed) {
            return NULL;
        }
    }
    return event;
}

static inline void startEventTap() {
    CGEventMask mask = (1 << kCGEventKeyDown) | (1 << kCGEventKeyUp) | (1 << kCGEventFlagsChanged);

    CFMachPortRef tap = CGEventTapCreate(
        kCGSessionEventTap,
        kCGHeadInsertEventTap,
        kCGEventTapOptionDefault,
        mask,
        eventCallback,
        NULL
    );

    if (!tap) {
        return;
    }

    CFRunLoopSourceRef runLoopSource = CFMachPortCreateRunLoopSource(kCFAllocatorDefault, tap, 0);
    CFRunLoopAddSource(CFRunLoopGetCurrent(), runLoopSource, kCFRunLoopCommonModes);
    CGEventTapEnable(tap, true);

    CFRunLoopRun();
}
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

// StartEventTap starts the global event tap. It blocks forever (runs a
// CFRunLoop), so call it in a goroutine.
func StartEventTap() {
	C.startEventTap()
}
