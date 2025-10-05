package dynamo

import (
	"aws_tui/internal/config"
	"aws_tui/internal/services/dynamo/filter"

	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
)

// View states
type viewState int

const (
	stateTableList viewState = iota
	stateItemList
	stateItemDetail
	stateConfirmDelete
	stateLoading
	stateColumnFilter
	stateItemFilter
)

// Model represents the DynamoDB child model
type Model struct {
	client        *dynamodb.Client
	config        *config.Config
	state         viewState
	previousState viewState

	// Data
	tables        []string
	selectedTable string
	tableKeys     map[string]TableKeySchema
	selectedItem  map[string]types.AttributeValue
	items         []map[string]types.AttributeValue
	allColumns    []string // All available columns

	// Filters (applied server-side)
	activeFilters []filter.FilterCondition
	itemFilter    filter.ItemFilterModel

	// UI components
	itemTable    table.Model
	columnFilter ColumnFilterModel
	selectedIdx  int
	err          error

	// Dimensions
	Width  int
	Height int
}

// TableKeySchema holds primary key information
type TableKeySchema struct {
	PartitionKey string
	SortKey      string
}

// NewModel creates a new DynamoDB model
func NewModel(client *dynamodb.Client) Model {
	cfg, err := config.LoadConfig()
	if err != nil {
		// If config fails to load, create a new one
		cfg = config.NewConfig()
	}

	return Model{
		client:    client,
		config:    cfg,
		state:     stateTableList,
		tableKeys: make(map[string]TableKeySchema),
	}
}

// Init initializes the DynamoDB model
func (m Model) Init() tea.Cmd {
	return m.loadTables()
}
