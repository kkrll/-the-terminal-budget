package tui

import (
	"fmt"
	"time"
)

func truncate(s string, length int) string {
	if len(s) <= length {
		return s
	}
	return s[:length-3] + "..."
}

func firstN(s string, n int) string {
	if len(s) > n {
		return s[:n]
	}
	return s
}

func formatTimeAgo(t time.Time) string {
	now := time.Now()
	duration := now.Sub(t)

	if duration < time.Minute {
		return "just now"
	} else if duration < time.Hour {
		minutes := int(duration.Minutes())
		if minutes == 1 {
			return "1 minute ago"
		}
		return fmt.Sprintf("%d minutes ago", minutes)
	} else if duration < 24*time.Hour {
		hours := int(duration.Hours())
		if hours == 1 {
			return "1 hour ago"
		}
		return fmt.Sprintf("%d hours ago", hours)
	} else {
		days := int(duration.Hours() / 24)
		if days == 1 {
			return "yesterday"
		}
		return fmt.Sprintf("%d days ago", days)
	}
}

func CleanSlates(m *model) {
	// Command input state
	m.commandInput = ""
	m.cursorPos = 0

	// Creation state
	m.creationInput = ""
	m.creationCursorPos = 0
	m.creationStep = 0
	m.selectedOption = 0
	m.isCustomInput = false

	// File selection state
	m.selectedFileIndex = 0
	m.isNewFile = false

	// Confirmation state
	m.confirmationMessage = ""
	m.confirmationAction = nil
	m.onConfirm = nil
	m.originScreen = greetingScreen

	// Error state
	m.err = nil
}
