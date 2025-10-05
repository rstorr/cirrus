package messages

import (
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

// Toast notification levels
type ToastLevel int

const (
	ToastInfo ToastLevel = iota
	ToastSuccess
	ToastWarning
	ToastError
)

// ToastMsg is a global message for showing toast notifications
type ToastMsg struct {
	Message string
	Level   ToastLevel
}

// ClearToastMsg clears the current toast
type ClearToastMsg struct{}

// ShowToast creates a toast message command
func ShowToast(message string, level ToastLevel) tea.Cmd {
	return func() tea.Msg {
		return ToastMsg{
			Message: message,
			Level:   level,
		}
	}
}

// ShowToastWithDuration creates a toast that auto-clears after duration
func ShowToastWithDuration(message string, level ToastLevel, duration time.Duration) tea.Cmd {
	return tea.Batch(
		func() tea.Msg {
			return ToastMsg{
				Message: message,
				Level:   level,
			}
		},
		tea.Tick(duration, func(t time.Time) tea.Msg {
			return ClearToastMsg{}
		}),
	)
}

// BackToMenuMsg signals the child service wants to return to main menu
type BackToMenuMsg struct{}

// BackToMenu creates a command to return to main menu
func BackToMenu() tea.Msg {
	return BackToMenuMsg{}
}
