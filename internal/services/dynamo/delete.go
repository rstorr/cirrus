package dynamo

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

type deleteCompleteMsg struct {
	deleted int
	err     error
}

func newDeleteConfirmInput() textinput.Model {
	ti := textinput.New()
	ti.Placeholder = "Type table name to confirm"
	ti.Focus()
	ti.Width = 50
	return ti
}

func (m Model) deleteAllItems() tea.Cmd {
	return func() tea.Msg {
		keySchema := m.tableKeys[m.selectedTable]
		deleted := 0

		// DynamoDB BatchWriteItem supports up to 25 items per batch
		batchSize := 25

		for i := 0; i < len(m.items); i += batchSize {
			end := i + batchSize
			if end > len(m.items) {
				end = len(m.items)
			}

			batch := m.items[i:end]

			// Build delete requests
			var writeRequests []types.WriteRequest
			for _, item := range batch {
				key := make(map[string]types.AttributeValue)
				key[keySchema.PartitionKey] = item[keySchema.PartitionKey]
				if keySchema.SortKey != "" {
					key[keySchema.SortKey] = item[keySchema.SortKey]
				}

				writeRequests = append(writeRequests, types.WriteRequest{
					DeleteRequest: &types.DeleteRequest{
						Key: key,
					},
				})
			}

			// Execute batch delete
			_, err := m.client.BatchWriteItem(context.Background(), &dynamodb.BatchWriteItemInput{
				RequestItems: map[string][]types.WriteRequest{
					m.selectedTable: writeRequests,
				},
			})

			if err != nil {
				return deleteCompleteMsg{
					deleted: deleted,
					err:     fmt.Errorf("failed to delete batch: %w", err),
				}
			}

			deleted += len(batch)
		}

		return deleteCompleteMsg{
			deleted: deleted,
			err:     nil,
		}
	}
}

func (m Model) updateDeleteConfirm(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "esc":
			m.state = stateTableList
			m.confirmInput.SetValue("")
			return m, nil

		case "enter":
			if m.confirmInput.Value() == m.selectedTable {
				m.state = stateDeleting
				m.deleteTotal = len(m.items)
				return m, m.deleteAllItems()
			}
			return m, nil
		}
	}

	m.confirmInput, cmd = m.confirmInput.Update(msg)
	return m, cmd
}
