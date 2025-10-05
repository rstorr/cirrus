package dynamo

import (
	"cirrus/internal/app/nav"
	"cirrus/internal/messages"
	"cirrus/internal/services/dynamo/filter"
	"log"

	tea "github.com/charmbracelet/bubbletea"
)

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {

	// Handle column filter state separately
	if m.state == stateColumnFilter {
		return m.updateColumnFilter(msg)
	}

	if m.state == stateItemFilter {
		return m.updateItemFilter(msg)
	}

	switch msg := msg.(type) {

	case tea.WindowSizeMsg:
		m.Width = msg.Width
		m.Height = msg.Height
		return m, nil

	case tea.KeyMsg:
		return m.handleKeyPress(msg)

	// Handle async command responses
	case tablesLoadedMsg:
		return m.handleTablesLoaded(msg)

	case tableKeysLoadedMsg:
		return m.handleTableKeysLoaded(msg)

	case itemsLoadedMsg:
		return m.handleItemsLoaded(msg)

	case ColumnFilterSavedMsg:
		// Save the column preferences
		m.config.SetTableColumns(m.selectedTable, msg.Columns)
		if err := m.config.Save(); err != nil {
			m.err = err
			return m, func() tea.Msg {
				return messages.ToastMsg{
					Message: "Failed to save preferences",
					Level:   messages.ToastError,
				}
			}
		}

		if keys, ok := m.tableKeys[m.selectedTable]; ok {
			m.itemTable, _ = BuildItemTable(ItemTableParams{
				PartitionKey:    keys.PartitionKey,
				SortKey:         keys.SortKey,
				Items:           m.items,
				FilteredColumns: msg.Columns,
			})
		}
		m.state = stateItemList
		return m, func() tea.Msg {
			return messages.ToastMsg{
				Message: "Column preferences saved",
				Level:   messages.ToastSuccess,
			}
		}
	}

	return m, nil
}

func (m Model) handleRefresh() (tea.Model, tea.Cmd) {
	switch m.state {
	case stateTableList:
		m.state = stateLoading
		return m, m.loadTables()
	case stateItemList:
		if m.selectedTable != "" {
			m.state = stateLoading
			return m, m.loadItems(m.activeFilters)
		}
	}
	return m, nil
}

func (m Model) handleKeyPress(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if m.err != nil && msg.String() != "esc" && msg.String() != "enter" {
		return m, nil
	}

	if m.state == stateItemList {
		return m.handleItemListInput(msg)
	}

	switch msg.String() {
	case "q":
		if m.state != stateTableList {
			return m.handleBack()
		}
		return m, func() tea.Msg {
			return nav.BackToMenuMsg{}
		}

	case "esc":
		if m.err != nil {
			m.err = nil
			return m, nil
		}
		return m.handleBack()

	case "enter":
		return m.handleEnter()

	case "d", "delete":
		if m.state == stateItemDetail {
			m.state = stateConfirmDelete
		}
		return m, nil

	case "r":
		return m.handleRefresh()

	case "up", "k":
		if m.selectedIdx > 0 {
			m.selectedIdx--
		}
		return m, nil

	case "down", "j":
		return m.handleDown()

	case "y":
		if m.state == stateConfirmDelete {
			return m.confirmDelete()
		}
		return m, nil

	case "n":
		if m.state == stateConfirmDelete {
			m.state = stateItemDetail
		}
		return m, nil
	}

	return m, nil
}

func (m Model) handleItemListInput(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "f":
		// Enter filter mode (preserve existing filters)
		if m.itemFilter.Conditions == nil {
			m.itemFilter = filter.NewItemFilterModel()
		}

		savedFilters := m.config.GetFilterConditions(m.selectedTable)
		// Copy active filters to editor
		m.itemFilter.Conditions = append([]filter.FilterCondition{}, savedFilters...)
		m.state = stateItemFilter
		return m, nil

	case "c":
		savedCols := m.config.GetTableColumns(m.selectedTable)
		m.columnFilter = NewColumnFilterModel(m.allColumns, savedCols)
		m.state = stateColumnFilter
		return m, nil

	case "r":
		// Refresh with current filters
		m.state = stateLoading
		return m, m.loadItems(m.activeFilters)

	case "enter":
		cursor := m.itemTable.Cursor()
		log.Printf("cursor %d", cursor)
		if cursor >= 0 && cursor < len(m.items) {
			m.selectedIdx = cursor
			m.state = stateItemDetail
		}
		return m, nil

	case "q":
		return m.handleBack()
	}

	var cmd tea.Cmd
	m.itemTable, cmd = m.itemTable.Update(msg)
	return m, cmd
}

func (m Model) handleBack() (tea.Model, tea.Cmd) {
	switch m.state {
	case stateItemDetail:
		m.state = stateItemList
		return m, nil

	case stateItemList:
		m.state = stateTableList
		m.selectedTable = ""
		m.items = nil
		m.selectedIdx = 0
		return m, nil

	case stateConfirmDelete:
		m.state = stateItemDetail
		return m, nil

	case stateLoading:
		m.state = m.previousState
		return m, nil
	}

	return m, nil
}

func (m Model) handleItemsLoaded(msg itemsLoadedMsg) (tea.Model, tea.Cmd) {
	if msg.err != nil {
		m.err = msg.err
		m.state = stateTableList
	} else {
		m.items = msg.items
		m.selectedIdx = 0
		m.state = stateItemList

		// Build table
		if keys, ok := m.tableKeys[m.selectedTable]; ok {
			m.itemTable, m.allColumns = BuildItemTable(ItemTableParams{
				PartitionKey: keys.PartitionKey, SortKey: keys.SortKey, Items: m.items,
			})
		}

	}
	return m, nil
}

// Message handlers
func (m Model) handleTablesLoaded(msg tablesLoadedMsg) (tea.Model, tea.Cmd) {
	if msg.err != nil {
		m.err = msg.err
		m.state = stateTableList
	} else {
		m.tables = msg.tables
		m.selectedIdx = 0
		m.state = stateTableList
	}
	return m, nil
}

func (m Model) handleTableKeysLoaded(msg tableKeysLoadedMsg) (tea.Model, tea.Cmd) {
	if msg.err != nil {
		m.err = msg.err
		m.state = stateTableList
	} else {
		m.tableKeys[msg.tableName] = msg.keys
		return m, m.loadItems(m.activeFilters)
	}
	return m, nil
}

func (m Model) handleEnter() (tea.Model, tea.Cmd) {
	switch m.state {
	case stateTableList:
		if len(m.tables) > 0 && m.selectedIdx < len(m.tables) {
			m.selectedTable = m.tables[m.selectedIdx]
			m.state = stateLoading
			return m, m.loadTableKeys(m.selectedTable)
		}

	case stateConfirmDelete:
		return m.confirmDelete()
	}

	return m, nil
}

func (m Model) handleDown() (tea.Model, tea.Cmd) {
	switch m.state {
	case stateTableList:
		if m.selectedIdx < len(m.tables)-1 {
			m.selectedIdx++
		}
	case stateItemList:
		if m.selectedIdx < len(m.items)-1 {
			m.selectedIdx++
		}
	}
	return m, nil
}

func (m Model) confirmDelete() (tea.Model, tea.Cmd) {
	if len(m.items) > 0 && m.selectedIdx < len(m.items) {
		m.state = stateLoading
		m.previousState = stateItemList
		return m, m.deleteItem(m.selectedTable, m.items[m.selectedIdx])
	}
	return m, nil
}

func (m Model) updateColumnFilter(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.Width = msg.Width
		m.Height = msg.Height
		m.columnFilter.Width = msg.Width
		m.columnFilter.Height = msg.Height

	case tea.KeyMsg:
		switch msg.String() {
		case "esc":
			m.state = stateItemList
			return m, nil
		case "s":
			// Save and apply filter
			m.state = stateItemList
			selected := m.columnFilter.GetSelectedColumns()
			return m, func() tea.Msg {
				return ColumnFilterSavedMsg{Columns: selected}
			}
		default:
			var cmd tea.Cmd
			m.columnFilter, cmd = m.columnFilter.Update(msg)
			return m, cmd
		}
	}

	var cmd tea.Cmd
	m.columnFilter, cmd = m.columnFilter.Update(msg)
	return m, cmd
}

func (m Model) updateItemFilter(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "esc":
			m.state = stateItemList
			return m, nil

		case "ctrl+s":
			// Apply filters and reload from DynamoDB
			m.config.SetFilterConditions(m.selectedTable, m.itemFilter.Conditions)

			if err := m.config.Save(); err != nil {
				m.err = err
				return m, func() tea.Msg {
					return messages.ToastMsg{
						Message: "Failed to save preferences",
						Level:   messages.ToastError,
					}
				}
			}

			m.activeFilters = m.itemFilter.Conditions
			m.state = stateLoading
			return m, tea.Batch(
				m.loadItems(m.activeFilters),
			)

		case "ctrl+x":
			// Clear filters and reload all items
			m.activeFilters = nil
			m.itemFilter.Conditions = nil
			m.state = stateLoading
			return m, tea.Batch(
				m.loadItems(nil),
			)
		}
	}

	var cmd tea.Cmd
	m.itemFilter, cmd = m.itemFilter.Update(msg)
	return m, cmd
}
