package app

import (
	"aws_tui/internal/messages"
	"aws_tui/internal/styles"

	"github.com/charmbracelet/lipgloss"
)

var (
	toastInfoStyle = lipgloss.NewStyle().
			Background(lipgloss.Color("62")).
			Foreground(lipgloss.Color("230")).
			Padding(0, 2).
			Bold(true)

	toastSuccessStyle = lipgloss.NewStyle().
				Background(lipgloss.Color("42")).
				Foreground(lipgloss.Color("230")).
				Padding(0, 2).
				Bold(true)

	toastWarningStyle = lipgloss.NewStyle().
				Background(lipgloss.Color("214")).
				Foreground(lipgloss.Color("0")).
				Padding(0, 2).
				Bold(true)

	toastErrorStyle = lipgloss.NewStyle().
			Background(lipgloss.Color("196")).
			Foreground(lipgloss.Color("230")).
			Padding(0, 2).
			Bold(true)
)

func (m Model) View() string {
	if m.quitting {
		return "Goodbye! 👋\n"
	}

	var content string
	switch m.currentService {
	case ServiceMenu:
		content = m.renderMenu()
		content = m.centerContent(content)
	case ServiceDynamoDB:
		content = m.dynamoDBModel.View()
	case ServiceCloudWatchLogs:
		content = m.cloudWatchModel.View()
	default:
		content = "Unknown service"
	}

	// Overlay toast if showing
	if m.showToast {
		content = m.renderToast(content)
	}

	return content

}

func (m Model) renderMenu() string {
	menu := styles.MenuTitleStyle.Render("🔧 AWS TUI") + "\n\n"

	menu += styles.MenuItemStyle.Render("1. 📊 DynamoDB - Manage tables and items") + "\n"
	menu += styles.MenuItemStyle.Render("2. 📝 CloudWatch Logs - View Lambda logs") + "\n" // ← New

	menu += styles.HelpStyle.Render("Select a number • q: Quit")

	return styles.MenuBorderStyle.Render(menu)
}

func (m Model) centerContent(content string) string {
	// Apply centering
	centered := lipgloss.NewStyle().
		Width(m.width).
		Height(m.height).
		AlignHorizontal(lipgloss.Center).
		AlignVertical(lipgloss.Center).
		Render(content)

	return centered
}

func (m Model) renderToast(content string) string {
	var toastStyle lipgloss.Style
	var icon string

	switch m.toastLevel {
	case messages.ToastInfo:
		toastStyle = toastInfoStyle
		icon = "ℹ️"
	case messages.ToastSuccess:
		toastStyle = toastSuccessStyle
		icon = "✅"
	case messages.ToastWarning:
		toastStyle = toastWarningStyle
		icon = "⚠️"
	case messages.ToastError:
		toastStyle = toastErrorStyle
		icon = "❌"
	}

	toast := toastStyle.Render(icon + " " + m.toastMessage)

	// Position toast at top center
	toastWidth := lipgloss.Width(toast)
	leftPadding := (m.width - toastWidth) / 2
	if leftPadding < 0 {
		leftPadding = 0
	}

	positionedToast := lipgloss.NewStyle().
		MarginLeft(leftPadding).
		MarginTop(1).
		Render(toast)

	// Overlay on top of content
	return lipgloss.JoinVertical(lipgloss.Left, positionedToast, content)
}
