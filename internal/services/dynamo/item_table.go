package dynamo

import (
	"fmt"

	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/lipgloss"
)

type ItemTableParams struct {
	PartitionKey    string
	SortKey         string
	Items           []map[string]types.AttributeValue
	FilteredColumns []string
}

func BuildItemTable(
	p ItemTableParams,
) (table.Model, []string) {
	var allColumns []string

	if len(p.Items) == 0 {
		return table.Model{}, allColumns
	}

	// Extract all unique columns from items
	columns := extractColumns(p)
	rows := buildRows(columns, p.Items)

	t := table.New(
		table.WithColumns(columns),
		table.WithRows(rows),
		table.WithFocused(true),
		table.WithHeight(min(len(rows), 50)),
	)

	s := table.DefaultStyles()
	s.Header = s.Header.
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("240")).
		BorderBottom(true).
		Bold(false)
	s.Selected = s.Selected.
		Foreground(lipgloss.Color("229")).
		Background(lipgloss.Color("57")).
		Bold(false)

	t.SetStyles(s)

	allColumns = extractAllColumns(p.Items)

	return t, allColumns
}

func extractAllColumns(items []map[string]types.AttributeValue) []string {
	columnSet := make(map[string]bool)

	for _, item := range items {
		for k := range item {
			columnSet[k] = true
		}
	}

	columns := make([]string, 0, len(columnSet))
	for col := range columnSet {
		columns = append(columns, col)
	}

	return columns
}

func extractColumns(
	p ItemTableParams,
) []table.Column {

	// If we have saved preferences, use them
	if len(p.FilteredColumns) > 0 {
		columns := make([]table.Column, 0, len(p.FilteredColumns))
		for _, colName := range p.FilteredColumns {
			// Only include columns that actually exist in the data
			exists := false
			for _, item := range p.Items {
				if _, ok := item[colName]; ok {
					exists = true
					break
				}
			}
			if exists {
				width := calculateColumnWidth(colName, p.Items)
				columns = append(columns, table.Column{Title: colName, Width: width})
			}
		}
		return columns
	}

	// Otherwise, extract all columns with primary keys first
	columnSet := make(map[string]bool)

	columnSet[p.PartitionKey] = true
	if p.SortKey != "" {
		columnSet[p.SortKey] = true
	}

	for _, item := range p.Items {
		for k := range item {
			columnSet[k] = true
		}
	}

	columns := make([]table.Column, 0, len(columnSet))

	// Add primary keys first
	columns = append(
		columns,
		table.Column{Title: p.PartitionKey, Width: calculateColumnWidth(p.PartitionKey, p.Items)},
	)
	if p.SortKey != "" {
		columns = append(
			columns,
			table.Column{Title: p.SortKey, Width: calculateColumnWidth(p.SortKey, p.Items)},
		)
		delete(columnSet, p.SortKey)
	}
	delete(columnSet, p.PartitionKey)

	// Add remaining columns
	for col := range columnSet {
		columns = append(
			columns,
			table.Column{Title: col, Width: calculateColumnWidth(col, p.Items)},
		)
	}

	return columns
}

func buildRows(
	columns []table.Column,
	items []map[string]types.AttributeValue,
) []table.Row {
	rows := make([]table.Row, len(items))

	for i, item := range items {
		row := make(table.Row, len(columns))
		for j, col := range columns {
			if val, ok := item[col.Title]; ok {
				row[j] = formatAttributeValueCompact(val)
			} else {
				row[j] = "-"
			}
		}
		rows[i] = row
	}

	return rows
}

// Format attribute value for compact display (table view)
func formatAttributeValueCompact(av types.AttributeValue) string {
	switch v := av.(type) {
	case *types.AttributeValueMemberS:
		if len(v.Value) > 70 {
			return v.Value[:67] + "..."
		}
		return v.Value
	case *types.AttributeValueMemberN:
		return v.Value
	case *types.AttributeValueMemberBOOL:
		if v.Value {
			return "✓"
		}
		return "✗"
	case *types.AttributeValueMemberNULL:
		return "∅"
	case *types.AttributeValueMemberL:
		return fmt.Sprintf("[%d]", len(v.Value))
	case *types.AttributeValueMemberM:
		return fmt.Sprintf("{%d}", len(v.Value))
	case *types.AttributeValueMemberSS:
		return fmt.Sprintf("Set<%d>", len(v.Value))
	case *types.AttributeValueMemberNS:
		return fmt.Sprintf("NumSet<%d>", len(v.Value))
	case *types.AttributeValueMemberBS:
		return fmt.Sprintf("BinSet<%d>", len(v.Value))
	case *types.AttributeValueMemberB:
		return fmt.Sprintf("Binary<%d bytes>", len(v.Value))
	default:
		return "unknown"
	}
}

func calculateColumnWidth(columnName string, items []map[string]types.AttributeValue) int {
	maxWidth := len(columnName) // Start with column name length

	// Sample first 10 items to determine width
	sampleSize := min(10, len(items))
	for i := 0; i < sampleSize; i++ {
		if val, ok := items[i][columnName]; ok {
			formatted := formatAttributeValueCompact(val)
			if len(formatted) > maxWidth {
				maxWidth = len(formatted)
			}
		}
	}

	// Set min/max bounds
	minWidth := 15
	maxWidthLimit := 1000
	setWidth := 0

	if maxWidth < minWidth {
		setWidth = minWidth
	} else if maxWidth > maxWidthLimit {
		setWidth = maxWidthLimit
	} else {
		setWidth = maxWidth + 2 // Add padding
	}

	return setWidth
}
