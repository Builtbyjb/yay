#include "keyevent.h"


extern int keyEventCallback(long long keycode, long long flags, long long eventType);

static CGEventRef eventCallback(CGEventTapProxy proxy, CGEventType type, CGEventRef event, void *refcon) {
    if (type == kCGEventKeyDown || type == kCGEventKeyUp || type == kCGEventFlagsChanged) {
        long long keycode = (long long)CGEventGetIntegerValueField(event, kCGKeyboardEventKeycode);
        long long flags = (long long)CGEventGetFlags(event);
        long long evType = (long long)type;
        int consumed = keyEventCallback(keycode, flags, evType);
        if (consumed) return NULL;
    }

    return event;
}

void startEventTap() {
    CGEventMask mask = (1 << kCGEventKeyDown) | ( 1 << kCGEventKeyUp) | (1 << kCGEventFlagsChanged);

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
