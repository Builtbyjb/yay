package tui

import (
	"strings"

	hook "github.com/robotn/gohook"
)

func padRight(s string, width int) string {
	if len(s) >= width {
		return s[:width]
	}
	return s + strings.Repeat(" ", width-len(s))
}

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	if maxLen <= 3 {
		return s[:maxLen]
	}
	return s[:maxLen-3] + "..."
}

func formatBool(b bool) string {
	if b {
		return "true"
	}
	return "false"
}

func displayKey(h string) string {
	if h == "" {
		return "---"
	}
	return h
}

func indexOf(slice []string, val string) int {
	for i, s := range slice {
		if s == val {
			return i
		}
	}
	return 0
}

func RawcodeToString(rawcode int) string {
	return hook.RawcodetoKeychar(uint16(rawcode))
}

func VerifiedModifier(mod string) bool {
	return true
}
