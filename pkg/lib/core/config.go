package core

import "runtime"

// Canonical modifier names understood by the operation
// (these should match whatever your underlying hotkey engine expects)
const (
	ModifierShift = "shift"
	ModifierAlt   = "alt"
	ModifierCtrl  = "ctrl"
	ModifierMeta  = "meta" // Command on macOS, Super/Win on Linux/Windows
)

// Available modes for the mode column
var AvailableModes = []string{"default", "fullscreen", "desktop"}

// Human-facing modifiers per OS (what you show in the TUI)

// macOS: standard system symbols
var AvailableModifiersMacos = []string{
	"⇧ Shift",
	"⌥ Option",
	"⌃ Control",
	"⌘ Command",
}

// Linux: you can keep this text-only if fonts are limited
var AvailableModifiersLinux = []string{
	"Shift",
	"Alt",
	"Ctrl",
	"Super",
}

// Windows: with icons for each modifier key
var AvailableModifiersWindows = []string{
	"⭡ Shift", // Shift
	"⎇ Alt",   // Alt
	"⌃ Ctrl",  // Control
	"⊞ Win",   // Windows / Super key
}

// Mapping from display strings to internal canonical modifier names

var modifierDisplayToInternalMacos = map[string]string{
	"⇧ Shift":   ModifierShift,
	"⌥ Option":  ModifierAlt,
	"⌃ Control": ModifierCtrl,
	"⌘ Command": ModifierMeta,
}

var internalToModifierDisplayMacos = map[string]string{
	ModifierShift: "⇧ Shift",
	ModifierAlt:   "⌥ Option",
	ModifierCtrl:  "⌃ Control",
	ModifierMeta:  "⌘ Command",
}

var modifierDisplayToInternalLinux = map[string]string{
	"Shift": ModifierShift,
	"Alt":   ModifierAlt,
	"Ctrl":  ModifierCtrl,
	"Super": ModifierMeta,
}

// Note: keys must exactly match the strings in AvailableModifiersWindows
var modifierDisplayToInternalWindows = map[string]string{
	"⭡ Shift": ModifierShift,
	"⎇ Alt":   ModifierAlt,
	"⌃ Ctrl":  ModifierCtrl,
	"⊞ Win":   ModifierMeta,
}

func ModifierFromDisplay(display string) string {
	switch runtime.GOOS {
	case "darwin":
		return modifierDisplayToInternalMacos[display]
	case "linux":
		return modifierDisplayToInternalLinux[display]
	case "windows":
		return modifierDisplayToInternalWindows[display]
	default:
		return display // fallback, or handle explicitly
	}
}

func DisplayFromModifier(modifier string) string {
	switch runtime.GOOS {
	case "darwin":
		return internalToModifierDisplayMacos[modifier]
	case "linux":
		return modifier // Linux display is same as internal
	case "windows":
		return modifier // Windows display is same as internal
	default:
		return modifier // fallback, or handle explicitly
	}
}
