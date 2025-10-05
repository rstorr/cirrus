package filter

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type FilterCondition struct {
	Column   string `json:"column"`
	Operator string `json:"operator"` // ==, !=, contains, startswith
	Value    string `json:"value"`
}

type ItemFilterModel struct {
	columnInput   textinput.Model
	operatorInput textinput.Model
	valueInput    textinput.Model
	focusIndex    int
	Conditions    []FilterCondition
}

func NewItemFilterModel() ItemFilterModel {
	colInput := textinput.New()
	colInput.Placeholder = "Column name (e.g., type)"
	colInput.Focus()
	colInput.Width = 30

	opInput := textinput.New()
	opInput.Placeholder = "Operator (==, !=, contains)"
	opInput.Width = 20

	valInput := textinput.New()
	valInput.Placeholder = "Value (e.g., consumer)"
	valInput.Width = 30

	return ItemFilterModel{
		columnInput:   colInput,
		operatorInput: opInput,
		valueInput:    valInput,
		focusIndex:    0,
	}
}

func (m ItemFilterModel) Update(msg tea.Msg) (ItemFilterModel, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "tab", "shift+tab":
			// Cycle through inputs
			if msg.String() == "tab" {
				m.focusIndex = (m.focusIndex + 1) % 3
			} else {
				m.focusIndex--
				if m.focusIndex < 0 {
					m.focusIndex = 2
				}
			}

			m.columnInput.Blur()
			m.operatorInput.Blur()
			m.valueInput.Blur()

			switch m.focusIndex {
			case 0:
				m.columnInput.Focus()
			case 1:
				m.operatorInput.Focus()
			case 2:
				m.valueInput.Focus()
			}

			return m, nil

		case "enter":
			// Add condition
			if m.columnInput.Value() != "" &&
				m.operatorInput.Value() != "" &&
				m.valueInput.Value() != "" {
				m.Conditions = append(m.Conditions, FilterCondition{
					Column:   m.columnInput.Value(),
					Operator: m.operatorInput.Value(),
					Value:    m.valueInput.Value(),
				})

				// Clear inputs
				m.columnInput.SetValue("")
				m.operatorInput.SetValue("")
				m.valueInput.SetValue("")
				m.focusIndex = 0
				m.columnInput.Focus()
				m.operatorInput.Blur()
				m.valueInput.Blur()
			}
			return m, nil

		case "backspace":
			// Remove last condition if inputs are empty
			if m.columnInput.Value() == "" &&
				m.operatorInput.Value() == "" &&
				m.valueInput.Value() == "" &&
				len(m.Conditions) > 0 {
				m.Conditions = m.Conditions[:len(m.Conditions)-1]
				return m, nil
			}
		}
	}

	// Update active input
	switch m.focusIndex {
	case 0:
		m.columnInput, cmd = m.columnInput.Update(msg)
	case 1:
		m.operatorInput, cmd = m.operatorInput.Update(msg)
	case 2:
		m.valueInput, cmd = m.valueInput.Update(msg)
	}

	return m, cmd
}

func (m ItemFilterModel) View() string {
	var b strings.Builder

	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("12"))

	b.WriteString(titleStyle.Render("🔍 Filter Items"))
	b.WriteString("\n\n")

	// Show active conditions
	if len(m.Conditions) > 0 {
		b.WriteString("Active Filters:\n")
		for i, cond := range m.Conditions {
			filterStyle := lipgloss.NewStyle().
				Foreground(lipgloss.Color("10")).
				Bold(true)
			b.WriteString(fmt.Sprintf(
				"  %d. %s\n",
				i+1,
				filterStyle.Render(
					fmt.Sprintf("%s %s %s", cond.Column, cond.Operator, cond.Value),
				),
			))
		}
		b.WriteString("\n")
	}

	// Input form
	b.WriteString("Add Filter Condition:\n\n")
	b.WriteString(fmt.Sprintf("Column:   %s\n", m.columnInput.View()))
	b.WriteString(fmt.Sprintf("Operator: %s\n", m.operatorInput.View()))
	b.WriteString(fmt.Sprintf("Value:    %s\n", m.valueInput.View()))

	b.WriteString("\n")
	helpStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("241"))
	b.WriteString(
		helpStyle.Render("Operators: == (equals), != (not equals), contains, startswith, endswith"),
	)
	b.WriteString("\n")
	b.WriteString(
		helpStyle.Render(
			"Tab: Next field • Enter: Add condition • Backspace: Remove last • Ctrl+S: Apply • Esc: Cancel",
		),
	)

	return b.String()
}
