package dynamo

import (
	"aws_tui/internal/styles"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

type ColumnFilterModel struct {
	availableColumns []string
	selectedColumns  map[string]bool
	cursorIdx        int
	Width            int
	Height           int
}

type ColumnFilterSavedMsg struct {
	Columns []string
}

func NewColumnFilterModel(available []string, selected []string) ColumnFilterModel {
	selectedMap := make(map[string]bool)

	// If we have saved preferences, use them
	if len(selected) > 0 {
		for _, col := range selected {
			selectedMap[col] = true
		}
	} else {
		// By default, select all columns
		for _, col := range available {
			selectedMap[col] = true
		}
	}

	return ColumnFilterModel{
		availableColumns: available,
		selectedColumns:  selectedMap,
		cursorIdx:        0,
	}
}

func (m ColumnFilterModel) Init() tea.Cmd {
	return nil
}

func (m ColumnFilterModel) Update(msg tea.Msg) (ColumnFilterModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.Width = msg.Width
		m.Height = msg.Height

	case tea.KeyMsg:
		switch msg.String() {
		case "up", "k":
			if m.cursorIdx > 0 {
				m.cursorIdx--
			}
		case "down", "j":
			if m.cursorIdx < len(m.availableColumns)-1 {
				m.cursorIdx++
			}
		case " ", "enter":
			// Toggle selection
			col := m.availableColumns[m.cursorIdx]
			m.selectedColumns[col] = !m.selectedColumns[col]
		case "a":
			// Select all
			for _, col := range m.availableColumns {
				m.selectedColumns[col] = true
			}
		case "n":
			// Select none
			for _, col := range m.availableColumns {
				m.selectedColumns[col] = false
			}
		}
	}

	return m, nil
}

func (m ColumnFilterModel) View() string {
	var b strings.Builder

	b.WriteString(styles.TitleStyle.Render("🔍 Column Filter"))
	b.WriteString("\n\n")
	b.WriteString("Select columns to display:\n\n")

	for i, col := range m.availableColumns {
		cursor := "  "
		if i == m.cursorIdx {
			cursor = "▶ "
		}

		checkbox := "☐"
		if m.selectedColumns[col] {
			checkbox = "☑"
		}

		line := cursor + checkbox + " " + col

		if i == m.cursorIdx {
			b.WriteString(styles.SelectedStyle.Render(line))
		} else {
			b.WriteString(line)
		}
		b.WriteString("\n")
	}

	b.WriteString("\n")
	b.WriteString(
		styles.HelpStyle.Render(
			"↑/↓: Navigate • Space/Enter: Toggle • a: All • n: None • s: Save • Esc: Cancel",
		),
	)

	return b.String()
}

func (m ColumnFilterModel) GetSelectedColumns() []string {
	selected := make([]string, 0)
	for _, col := range m.availableColumns {
		if m.selectedColumns[col] {
			selected = append(selected, col)
		}
	}
	return selected
}
