package dynamo

import (
	"aws_tui/internal/services/dynamo/filter"
	"context"
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	tea "github.com/charmbracelet/bubbletea"
)

// Message types
type tablesLoadedMsg struct {
	tables []string
	err    error
}

type tableKeysLoadedMsg struct {
	tableName string
	keys      TableKeySchema
	err       error
}

type itemsLoadedMsg struct {
	items []map[string]types.AttributeValue
	err   error
}

type itemDeletedMsg struct {
	err error
}

func (m Model) loadTables() tea.Cmd {
	return func() tea.Msg {
		result, err := m.client.ListTables(context.TODO(), &dynamodb.ListTablesInput{})
		if err != nil {
			return tablesLoadedMsg{err: err}
		}

		var filteredTables []string
		for _, table := range result.TableNames {
			if strings.HasPrefix(table, "dev-cot") {
				filteredTables = append(filteredTables, table)
			}
		}

		return tablesLoadedMsg{tables: filteredTables}
	}
}

func (m Model) loadTableKeys(tableName string) tea.Cmd {
	return func() tea.Msg {
		result, err := m.client.DescribeTable(context.TODO(), &dynamodb.DescribeTableInput{
			TableName: aws.String(tableName),
		})
		if err != nil {
			return tableKeysLoadedMsg{tableName: tableName, err: err}
		}

		keys := TableKeySchema{}
		for _, key := range result.Table.KeySchema {
			if key.KeyType == types.KeyTypeHash {
				keys.PartitionKey = *key.AttributeName
			} else if key.KeyType == types.KeyTypeRange {
				keys.SortKey = *key.AttributeName
			}
		}

		return tableKeysLoadedMsg{
			tableName: tableName,
			keys:      keys,
		}
	}
}

func (m *Model) loadItems(filters []filter.FilterCondition) tea.Cmd {
	return func() tea.Msg {
		input := &dynamodb.ScanInput{
			TableName: aws.String(m.selectedTable),
		}

		// Build FilterExpression from conditions
		if len(filters) > 0 {
			filterExpr, exprAttrNames, exprAttrValues := buildFilterExpression(filters)
			input.FilterExpression = aws.String(filterExpr)
			input.ExpressionAttributeNames = exprAttrNames
			input.ExpressionAttributeValues = exprAttrValues
		}

		result, err := m.client.Scan(context.TODO(), input)
		if err != nil {
			return itemsLoadedMsg{err: err}
		}

		return itemsLoadedMsg{
			items: result.Items,
		}
	}
}

func buildFilterExpression(
	filters []filter.FilterCondition,
) (string, map[string]string, map[string]types.AttributeValue) {
	var expressions []string
	exprAttrNames := make(map[string]string)
	exprAttrValues := make(map[string]types.AttributeValue)

	for i, filter := range filters {
		nameKey := fmt.Sprintf("#attr%d", i)
		valueKey := fmt.Sprintf(":val%d", i)

		exprAttrNames[nameKey] = filter.Column
		exprAttrValues[valueKey] = &types.AttributeValueMemberS{Value: filter.Value}

		var expr string
		switch filter.Operator {
		case "==":
			expr = fmt.Sprintf("%s = %s", nameKey, valueKey)
		case "!=":
			expr = fmt.Sprintf("%s <> %s", nameKey, valueKey)
		case "contains":
			expr = fmt.Sprintf("contains(%s, %s)", nameKey, valueKey)
		case "startswith":
			expr = fmt.Sprintf("begins_with(%s, %s)", nameKey, valueKey)
		case "endswith":
			// DynamoDB doesn't have endswith, need to use contains as fallback
			expr = fmt.Sprintf("contains(%s, %s)", nameKey, valueKey)
		}

		expressions = append(expressions, expr)
	}

	filterExpr := strings.Join(expressions, " AND ")
	return filterExpr, exprAttrNames, exprAttrValues
}

func (m Model) deleteItem(tableName string, item map[string]types.AttributeValue) tea.Cmd {
	return func() tea.Msg {
		keySchema := m.tableKeys[tableName]
		key := make(map[string]types.AttributeValue)

		if val, ok := item[keySchema.PartitionKey]; ok {
			key[keySchema.PartitionKey] = val
		}
		if keySchema.SortKey != "" {
			if val, ok := item[keySchema.SortKey]; ok {
				key[keySchema.SortKey] = val
			}
		}

		_, err := m.client.DeleteItem(context.TODO(), &dynamodb.DeleteItemInput{
			TableName: aws.String(tableName),
			Key:       key,
		})
		return itemDeletedMsg{err: err}
	}
}
