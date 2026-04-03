#include "keyevent.h"
#include <stdio.h>

extern int keyEventCallback(long long keycode, long long flags, long long eventType);

// Store the tap reference so we can re-enable it
static CFMachPortRef eventTap = NULL;

static CGEventRef eventCallback(CGEventTapProxy proxy, CGEventType type, CGEventRef event, void *refcon) {
    // Re-enable the tap if macOS disabled it due to timeout
    if (type == kCGEventTapDisabledByTimeout || type == kCGEventTapDisabledByUserInput) {
        fprintf(stderr, "Event tap was disabled (type=%d), re-enabling...\n", (int)type);
        fflush(stderr);
        if (eventTap) {
            CGEventTapEnable(eventTap, true);
        }
        return event;
    }

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

    eventTap = CGEventTapCreate(
        kCGSessionEventTap,
        kCGHeadInsertEventTap,
        kCGEventTapOptionDefault,
        mask,
        eventCallback,
        NULL
    );

    if (!eventTap) {
        fprintf(stderr, "ERROR: CGEventTapCreate failed. "
            "Make sure Accessibility is enabled for this binary in "
            "System Settings > Privacy & Security > Accessibility\n");
        fflush(stderr);
        return;
    }

    // fprintf(stdout, "Event tap created successfully, listening for keyboard events...\n");
    fflush(stdout);

    CFRunLoopSourceRef runLoopSource = CFMachPortCreateRunLoopSource(kCFAllocatorDefault, eventTap, 0);
    CFRunLoopAddSource(CFRunLoopGetCurrent(), runLoopSource, kCFRunLoopCommonModes);
    CGEventTapEnable(eventTap, true);

    CFRunLoopRun();
}
