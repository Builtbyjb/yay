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

func displayMod(m string) string {
	if m == "" {
		return "---"
	}
	return m
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
		var mod, key string
		if s.HotKey != "" {
			keys := strings.Split(s.HotKey, "+")
			mod = keys[0]
			key = keys[1]
		}

		modelSettings = append(modelSettings, ModelSetting{
			Id:      s.Id,
			Name:    s.Name,
			Mod:     core.DisplayFromModifier(mod),
			Key:     key,
			Mode:    s.Mode,
			Enabled: s.Enabled,
		})
	}

	return modelSettings
}
