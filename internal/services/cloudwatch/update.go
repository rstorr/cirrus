package cloudwatch

import (
	"aws_tui/internal/app/nav"

	tea "github.com/charmbracelet/bubbletea"
)

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	if m.state == stateRipgrepInput {
		return m.updateRipgrepInput(msg)
	}

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.Width = msg.Width
		m.Height = msg.Height

		// Update viewport size (it's already initialized)
		m.viewport.Width = msg.Width
		m.viewport.Height = msg.Height - 12

		if m.state == stateLogStream {
			m.viewport.SetContent(m.renderLogs())
		}

		m.ready = true // Mark as ready after first resize
		return m, nil

	case tea.KeyMsg:
		if m.state == stateLogStream && m.isScrollKey(msg) {
			m.viewport, cmd = m.viewport.Update(msg)
			return m, cmd
		}
		return m.handleKeyPress(msg)

	case logGroupsLoadedMsg:
		if msg.err != nil {
			m.err = msg.err
			m.state = stateLogGroupList
		} else {
			m.logGroups = msg.groups
			m.selectedIdx = 0
			m.state = stateLogGroupList
		}
		return m, nil

	case filteredLogsMsg:
		if msg.err != nil {
			m.err = msg.err
			m.state = stateLogStream
			m.filteredView = false
		} else {
			m.viewport.SetContent(msg.output)
			m.viewport.GotoTop()
			m.state = stateLogStream
			m.filteredView = true
		}
		return m, nil

	case logEventsLoadedMsg:
		if msg.err != nil {
			m.err = msg.err
			m.state = stateLogGroupList
		} else {
			m.logEvents = msg.events
			m.state = stateLogStream
			m.viewport.SetContent(m.renderLogs())
			m.viewport.GotoBottom()

			m.allLogsText = m.renderLogs()
			m.viewport.SetContent(m.allLogsText)
			m.viewport.GotoBottom()
			m.filteredView = false
		}
		return m, nil
	}
	return m, nil
}

// Add this helper function
func (m Model) isScrollKey(msg tea.KeyMsg) bool {
	switch msg.String() {
	case "up", "down", "k", "j",
		"pgup", "pgdown",
		"home", "end",
		"ctrl+u", "ctrl+d",
		"ctrl+b", "ctrl+f":
		return true
	}
	return false
}

func (m Model) updateRipgrepInput(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "esc":
			m.state = stateLogStream
			return m, nil

		case "enter":
			pattern := m.ripgrepInput.Value()
			m.ripgrepInput.SetValue("")
			m.state = stateLoading
			return m, tea.Batch(
				m.filterWithRipgrep(pattern),
				m.spinner.Tick,
			)
		}
	}

	m.ripgrepInput, cmd = m.ripgrepInput.Update(msg)
	return m, cmd
}

func (m Model) handleKeyPress(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if m.err != nil {
		if msg.String() == "esc" || msg.String() == "enter" {
			m.err = nil
		}
		return m, nil
	}

	switch m.state {
	case stateLogGroupList:
		switch msg.String() {
		case "q":
			return m, func() tea.Msg {
				return nav.BackToMenuMsg{}
			}
		case "up", "k":
			if m.selectedIdx > 0 {
				m.selectedIdx--
			}
		case "down", "j":
			if m.selectedIdx < len(m.logGroups)-1 {
				m.selectedIdx++
			}
		case "enter":
			if len(m.logGroups) > 0 {
				m.currentGroup = *m.logGroups[m.selectedIdx].LogGroupName
				m.state = stateLoading
				return m, m.loadLogEvents(m.currentGroup)
			}
		case "r":
			m.state = stateLoading
			return m, m.loadLambdaLogGroups()
		}

	case stateLogStream:
		switch msg.String() {
		case "q", "esc":
			m.state = stateLogGroupList
			m.currentGroup = ""
			m.logEvents = nil
			return m, nil
		case "r":
			m.state = stateLoading
			return m, m.loadLogEvents(m.currentGroup)
		case "/": // ← Changed from 'f' to '/' (vim-style)
			m.state = stateRipgrepInput
			m.ripgrepInput.Focus()
			return m, nil
		case "c": // ← Clear filter
			if m.filteredView {
				m.filteredView = false
				m.viewport.SetContent(m.allLogsText)
				m.viewport.GotoBottom()
			}
			return m, nil
		case "1", "2", "3", "4", "5", "6", "7", "8", "9":
			// Quick switch to log group by number
			idx := int(msg.String()[0] - '1')
			if idx < len(m.logGroups) {
				m.currentGroup = *m.logGroups[idx].LogGroupName
				m.selectedIdx = idx
				m.state = stateLoading
				return m, m.loadLogEvents(m.currentGroup)
			}
		}
	}

	return m, nil
}
