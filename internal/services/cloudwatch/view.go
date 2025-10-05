package cloudwatch

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
)

var (
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("63")).
			MarginBottom(1)

	selectedStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("170")).
			Bold(true)

	errorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("196")).
			Bold(true)

	helpStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("241")).
			MarginTop(1)

	timestampStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("241"))

	logGroupStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("86"))

	numberStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("33")).
			Bold(true)
)

func (m Model) View() string {
	if m.err != nil {
		return m.renderError()
	}

	switch m.state {
	case stateLoading:
		return "⏳ Loading logs..."
	case stateLogGroupList:
		return m.renderLogGroupList()
	case stateLogStream:
		return m.renderLogStream()
	case stateRipgrepInput:
		return m.renderRipgrepInput()
	}
	return ""
}

func (m Model) renderRipgrepInput() string {
	var b strings.Builder

	functionName := strings.TrimPrefix(m.currentGroup, "/aws/lambda/")
	title := fmt.Sprintf("🔍 Filter Logs: %s", functionName)
	b.WriteString(titleStyle.Render(title))
	b.WriteString("\n\n")

	b.WriteString("Enter ripgrep pattern:\n\n")
	b.WriteString(m.ripgrepInput.View())
	b.WriteString("\n\n")

	b.WriteString(helpStyle.Render("Examples: ERROR | 'status.*500' | '\\b(ERROR|WARN)\\b'"))
	b.WriteString("\n")
	b.WriteString(helpStyle.Render("Enter: Search • Esc: Cancel"))

	return b.String()
}

func (m Model) renderError() string {
	var b strings.Builder
	b.WriteString(errorStyle.Render("❌ Error") + "\n\n")
	b.WriteString(m.err.Error() + "\n\n")
	b.WriteString(helpStyle.Render("Press Enter or Esc to continue"))
	return b.String()
}

func (m Model) renderLogGroupList() string {
	var b strings.Builder

	b.WriteString(titleStyle.Render("📝 Lambda Log Groups (Last 10 minutes)"))
	b.WriteString("\n\n")

	if len(m.logGroups) == 0 {
		b.WriteString("No Lambda log groups found\n")
	} else {
		for i, group := range m.logGroups {
			// Extract Lambda function name from log group
			name := *group.LogGroupName
			functionName := strings.TrimPrefix(name, "/aws/lambda/")

			number := numberStyle.Render(fmt.Sprintf("[%d] ", i+1))

			if i == m.selectedIdx {
				b.WriteString(selectedStyle.Render("▶ " + number + functionName))
			} else {
				b.WriteString("  " + number + functionName)
			}
			b.WriteString("\n")
		}
	}

	b.WriteString("\n")
	b.WriteString(
		helpStyle.Render(
			"↑/↓: Navigate • Enter: View Logs • 1-9: Quick Switch • r: Refresh • q: Back",
		),
	)

	return b.String()
}

func (m Model) renderLogStream() string {
	var b strings.Builder

	functionName := strings.TrimPrefix(m.currentGroup, "/aws/lambda/")
	title := fmt.Sprintf("📋 Logs: %s (Last 10 minutes)", functionName)
	b.WriteString(titleStyle.Render(title))
	b.WriteString("\n")

	if len(m.logEvents) == 0 {
		b.WriteString("No log events in the last 10 minutes\n")
	} else {
		b.WriteString(fmt.Sprintf("Total events: %d\n\n", len(m.logEvents)))
		b.WriteString(m.viewport.View())
	}

	b.WriteString("\n")
	help := "↑/↓: Scroll • 1-9: Switch Lambda • r: Refresh • Esc: Back to List"
	b.WriteString(helpStyle.Render(help))

	return b.String()
}

func (m Model) renderLogs() string {
	var b strings.Builder

	functionName := strings.TrimPrefix(m.currentGroup, "/aws/lambda/")
	title := fmt.Sprintf("📋 Logs: %s (Last 10 minutes)", functionName)
	b.WriteString(titleStyle.Render(title))
	b.WriteString("\n")

	for _, event := range m.logEvents {
		if event.Message == nil {
			continue
		}

		message := strings.TrimSpace(*event.Message)
		timestamp := time.UnixMilli(*event.Timestamp).Format("15:04:05.000")

		line := fmt.Sprintf("%s %s", timestampStyle.Render(timestamp), message)
		b.WriteString(line)
		b.WriteString("\n")
	}

	return b.String()
}
