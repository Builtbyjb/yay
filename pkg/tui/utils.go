package tui

import (
	"strings"

	"github.com/Builtbyjb/yay/pkg/lib/core"
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

func mapToModelSetting(settings []core.Setting) []ModelSetting {
	modelSettings := []ModelSetting{}

	for _, s := range settings {
		modelSettings = append(modelSettings, ModelSetting{
			Id:      s.Id,
			Name:    s.Name,
			HotKey:  s.HotKey,
			Mode:    s.Mode,
			Enabled: s.Enabled,
		})
	}

	return modelSettings
}
