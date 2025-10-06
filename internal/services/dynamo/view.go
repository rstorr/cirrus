package dynamo

import (
	"cirrus/internal/styles"
	"fmt"
	"sort"
	"strings"

	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/charmbracelet/lipgloss"
)

func (m Model) View() string {
	if m.err != nil {
		return m.renderError()
	}

	var content string
	switch m.state {
	case stateLoading:
		content = m.renderLoading()
	case stateTableList:
		content = m.renderTableList()
		content = m.centerContent(content)
	case stateItemList:
		content = m.renderItemListTable()
	case stateItemDetail:
		content = m.renderItemDetail()
	case stateColumnFilter:
		content = m.columnFilter.View()
	case stateItemFilter:
		content = m.itemFilter.View()
	case stateDeleteConfirm:
		return m.renderDeleteConfirm()
	case stateDeleting:
		return m.renderDeleting()
	}
	return content
}

func (m Model) centerContent(content string) string {
	return lipgloss.Place(
		m.Width,
		m.Height,
		lipgloss.Center,
		lipgloss.Center,
		content,
	)
}

func (m Model) renderError() string {
	var b strings.Builder
	b.WriteString(styles.ErrorStyle.Render("❌ Error") + "\n\n")
	b.WriteString(m.err.Error() + "\n\n")
	b.WriteString(styles.HelpStyle.Render("Press Enter or Esc to continue"))
	return b.String()
}

func (m Model) renderLoading() string {
	return styles.LoadingStyle.Render("⏳ Loading...")
}

func (m Model) renderTableList() string {
	var b strings.Builder

	b.WriteString(styles.TitleStyle.Render("DynamoDB Tables"))
	b.WriteString("\n\n")

	if len(m.tables) == 0 {
		b.WriteString("No tables found\n")
	} else {
		for i, table := range m.tables {
			if i == m.selectedIdx {
				b.WriteString(styles.SelectedStyle.Render("▶ " + table))
			} else {
				b.WriteString("  " + table)
			}
			b.WriteString("\n")
		}
	}

	b.WriteString("\n")
	b.WriteString(
		styles.HelpStyle.Render(
			"↑/↓: Navigate • Enter: Select • r: Refresh • e: Empty Table • q: Back to Menu",
		),
	)

	return b.String()
}

func (m Model) renderItemListTable() string {
	var b strings.Builder

	b.WriteString(styles.TitleStyle.Render(fmt.Sprintf("Items in %s", m.selectedTable)))
	b.WriteString("\n")

	// Show active filters badge
	if len(m.activeFilters) > 0 {
		filterBadgeStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("0")).
			Background(lipgloss.Color("10")).
			Bold(true).
			Padding(0, 1)

		var filters []string
		for _, f := range m.activeFilters {
			filters = append(filters, fmt.Sprintf("%s %s %s", f.Column, f.Operator, f.Value))
		}

		badge := filterBadgeStyle.Render(fmt.Sprintf("🔍 %d filter(s) active", len(m.activeFilters)))
		detail := styles.HelpStyle.Render(strings.Join(filters, " AND "))
		b.WriteString(badge + " " + detail)
		b.WriteString("\n")
	}

	b.WriteString(fmt.Sprintf("Total items: %d\n", len(m.items)))
	b.WriteString("\n")
	b.WriteString(m.itemTable.View())
	b.WriteString("\n\n")

	help := "↑/↓: Navigate • Enter: View Details • Delete: Delete Item • c: Column Filter • "
	if len(m.activeFilters) > 0 {
		help += "f: Edit Filters "
	} else {
		help += "f: Add Filters • "
	}
	help += "r: Refresh • q: Back"

	b.WriteString(styles.HelpStyle.Render(help))

	return b.String()
}

func (m Model) renderItemDetail() string {
	if m.selectedIdx >= len(m.items) {
		return "Invalid item selection"
	}

	var b strings.Builder
	item := m.items[m.selectedIdx]

	title := fmt.Sprintf("📋 Item Details - %s", m.selectedTable)
	b.WriteString(styles.TitleStyle.Render(title))
	b.WriteString("\n\n")

	// Get sorted keys for consistent display
	keys := make([]string, 0, len(item))
	for k := range item {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	// Highlight primary keys
	tableKeys := m.tableKeys[m.selectedTable]

	for _, k := range keys {
		v := item[k]

		// Mark primary keys
		keyDisplay := k
		if k == tableKeys.PartitionKey {
			keyDisplay = k + " 🔑"
		} else if k == tableKeys.SortKey {
			keyDisplay = k + " 🗝️"
		}

		b.WriteString(styles.KeyStyle.Render(keyDisplay))
		b.WriteString("\n")
		b.WriteString(formatAttributeValueDetailed(v, 1))
		b.WriteString("\n\n")
	}

	b.WriteString(styles.HelpStyle.Render("d: Delete • Esc: Back to List"))

	return styles.BoxStyle.Render(b.String())
}

// Format attribute value for detailed display
func formatAttributeValueDetailed(av types.AttributeValue, indent int) string {
	prefix := strings.Repeat("  ", indent)

	switch v := av.(type) {
	case *types.AttributeValueMemberS:
		return styles.ValueStyle.Render(fmt.Sprintf("%s%s %s",
			prefix,
			styles.TypeStyle.Render("(String)"),
			v.Value))

	case *types.AttributeValueMemberN:
		return styles.ValueStyle.Render(fmt.Sprintf("%s%s %s",
			prefix,
			styles.TypeStyle.Render("(Number)"),
			v.Value))

	case *types.AttributeValueMemberBOOL:
		return styles.ValueStyle.Render(fmt.Sprintf("%s%s %t",
			prefix,
			styles.TypeStyle.Render("(Boolean)"),
			v.Value))

	case *types.AttributeValueMemberNULL:
		return styles.ValueStyle.Render(fmt.Sprintf("%s%s",
			prefix,
			styles.TypeStyle.Render("(Null)")))

	case *types.AttributeValueMemberL:
		var b strings.Builder
		b.WriteString(styles.TypeStyle.Render(fmt.Sprintf("%s(List - %d items)\n", prefix, len(v.Value))))
		for i, item := range v.Value {
			b.WriteString(fmt.Sprintf("%s  [%d] ", prefix, i))
			b.WriteString(formatAttributeValueDetailed(item, indent+1))
			b.WriteString("\n")
		}
		return b.String()

	case *types.AttributeValueMemberM:
		var b strings.Builder
		b.WriteString(styles.TypeStyle.Render(fmt.Sprintf("%s(Map - %d fields)\n", prefix, len(v.Value))))

		// Sort keys for consistent display
		keys := make([]string, 0, len(v.Value))
		for k := range v.Value {
			keys = append(keys, k)
		}
		sort.Strings(keys)

		for _, key := range keys {
			val := v.Value[key]
			b.WriteString(styles.KeyStyle.Render(fmt.Sprintf("%s  %s:", prefix, key)))
			b.WriteString("\n")
			b.WriteString(formatAttributeValueDetailed(val, indent+1))
			b.WriteString("\n")
		}
		return b.String()

	case *types.AttributeValueMemberSS:
		var b strings.Builder
		b.WriteString(styles.TypeStyle.Render(fmt.Sprintf("%s(String Set - %d items)\n", prefix, len(v.Value))))
		for i, item := range v.Value {
			b.WriteString(styles.ValueStyle.Render(fmt.Sprintf("%s  - %s", prefix, item)))
			if i < len(v.Value)-1 {
				b.WriteString("\n")
			}
		}
		return b.String()

	case *types.AttributeValueMemberNS:
		var b strings.Builder
		b.WriteString(styles.TypeStyle.Render(fmt.Sprintf("%s(Number Set - %d items)\n", prefix, len(v.Value))))
		for i, item := range v.Value {
			b.WriteString(styles.ValueStyle.Render(fmt.Sprintf("%s  - %s", prefix, item)))
			if i < len(v.Value)-1 {
				b.WriteString("\n")
			}
		}
		return b.String()

	case *types.AttributeValueMemberBS:
		return styles.TypeStyle.Render(fmt.Sprintf("%s(Binary Set - %d items)",
			prefix, len(v.Value)))

	case *types.AttributeValueMemberB:
		return styles.TypeStyle.Render(fmt.Sprintf("%s(Binary - %d bytes)",
			prefix, len(v.Value)))

	default:
		return fmt.Sprintf("%sunknown type", prefix)
	}
}

func (m Model) renderDeleteConfirm() string {
	var b strings.Builder

	warningStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("196")).
		Padding(1, 0)

	b.WriteString(warningStyle.Render("⚠️  EMPTY TABLE"))
	b.WriteString("\n\n")

	infoStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("241"))

	b.WriteString(
		infoStyle.Render(fmt.Sprintf("You are about to delete ALL %d items from:", len(m.items))),
	)
	b.WriteString("\n\n")

	tableStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("12"))

	b.WriteString(tableStyle.Render(m.selectedTable))
	b.WriteString("\n\n")

	b.WriteString(warningStyle.Render("THIS ACTION CANNOT BE UNDONE!"))
	b.WriteString("\n\n")

	b.WriteString(infoStyle.Render("Type the table name to confirm:"))
	b.WriteString("\n\n")
	b.WriteString(m.confirmInput.View())
	b.WriteString("\n\n")

	keyBindStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("205")).
		Bold(true)

	b.WriteString(keyBindStyle.Render("Enter"))
	b.WriteString(infoStyle.Render(": Confirm deletion • "))

	b.WriteString(keyBindStyle.Render("Esc"))
	b.WriteString(infoStyle.Render(": Cancel"))

	return b.String()
}

func (m Model) renderDeleting() string {
	var b strings.Builder

	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("12")).
		Padding(1, 0)

	b.WriteString(titleStyle.Render("🗑️  Deleting Items..."))
	b.WriteString("\n\n")

	infoStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("241"))

	b.WriteString(
		infoStyle.Render(fmt.Sprintf("Deleting %d items from %s", m.deleteTotal, m.selectedTable)),
	)
	b.WriteString("\n\n")

	b.WriteString("⏳ Please wait...")

	return b.String()
}
