package app

import (
	"aws_tui/internal/app/nav"
	"aws_tui/internal/messages"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

// Custom message types for navigation
type switchToServiceMsg ServiceType

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		// Forward window size to active child model
		return m.forwardToChildModel(msg)

	case tea.KeyMsg:
		// Global keybindings
		switch msg.String() {
		case "ctrl+c":
			m.quitting = true
			return m, tea.Quit

		case "q":
			// Quit from menu, otherwise handled by child
			if m.currentService == ServiceMenu {
				m.quitting = true
				return m, tea.Quit
			}
		}

		// Service selection from menu
		if m.currentService == ServiceMenu {
			return m.handleMenuInput(msg)
		}

		// Forward to child model
		return m.forwardToChildModel(msg)

	case messages.ToastMsg: // ← New: Handle toast
		m.toastMessage = msg.Message
		m.toastLevel = msg.Level
		m.showToast = true
		return m, tea.Tick(3*time.Second, func(t time.Time) tea.Msg {
			return messages.ClearToastMsg{}
		})

	case messages.ClearToastMsg:
		m.showToast = false
		return m, nil

	case switchToServiceMsg:
		m.currentService = ServiceType(msg)
		// Initialize the selected service
		switch m.currentService {
		case ServiceDynamoDB:
			var cmd tea.Cmd
			m.dynamoDBModel, cmd = m.dynamoDBModel.Update(nil)
			// Trigger Init for the child model
			initCmd := m.dynamoDBModel.Init()
			return m, tea.Batch(cmd, initCmd)
		}
		return m, nil

	case nav.BackToMenuMsg:
		m.currentService = ServiceMenu
		return m, nil
	}

	// Forward all other messages to the active child model
	return m.forwardToChildModel(msg)
}

func (m Model) handleMenuInput(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "1":
		return m, func() tea.Msg {
			return switchToServiceMsg(ServiceDynamoDB)
		}
	case "2":
		return m, func() tea.Msg {
			return switchToServiceMsg(ServiceCloudWatchLogs) // ← New
		}
	}
	return m, nil
}

func (m Model) forwardToChildModel(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch m.currentService {
	case ServiceDynamoDB:
		m.dynamoDBModel, cmd = m.dynamoDBModel.Update(msg)
	case ServiceCloudWatchLogs: // ← New
		m.cloudWatchModel, cmd = m.cloudWatchModel.Update(msg)

	default:
		switch msg := msg.(type) {
		case tea.WindowSizeMsg:
			m.cloudWatchModel, cmd = m.cloudWatchModel.Update(msg)
			m.dynamoDBModel, cmd = m.dynamoDBModel.Update(msg)
		}
	}

	return m, cmd
}
