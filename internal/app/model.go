package app

import (
	"cirrus/internal/messages"
	"cirrus/internal/services/cloudwatch"
	"cirrus/internal/services/dynamo"

	"github.com/aws/aws-sdk-go-v2/service/cloudwatchlogs"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	tea "github.com/charmbracelet/bubbletea"
)

// ServiceType represents different AWS services
type ServiceType int

const (
	ServiceMenu ServiceType = iota
	ServiceDynamoDB
	ServiceCloudWatchLogs
)

// Model is the root application model
type Model struct {
	width          int
	height         int
	currentService ServiceType

	// Child models
	dynamoDBModel   tea.Model
	cloudWatchModel tea.Model

	// Toast notification
	toastMessage string
	toastLevel   messages.ToastLevel
	showToast    bool

	quitting bool
}

// NewModel creates a new root model
func NewModel(dynamoClient *dynamodb.Client, logsClient *cloudwatchlogs.Client) Model {
	return Model{
		currentService:  ServiceMenu,
		dynamoDBModel:   dynamo.NewModel(dynamoClient),
		cloudWatchModel: cloudwatch.NewModel(logsClient),
	}
}

func (m Model) Init() tea.Cmd {
	return nil
}
