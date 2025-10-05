package cloudwatch

import (
	"github.com/aws/aws-sdk-go-v2/service/cloudwatchlogs"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatchlogs/types"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type viewState int

const (
	stateLogGroupList viewState = iota
	stateLogStream
	stateLoading
	stateRipgrepInput
)

type Model struct {
	client *cloudwatchlogs.Client
	state  viewState

	// Data
	logGroups    []types.LogGroup
	selectedIdx  int
	currentGroup string
	logEvents    []types.FilteredLogEvent
	allLogsText  string

	// AWS
	env string

	// UI
	viewport viewport.Model
	ready    bool
	err      error
	spinner  spinner.Model

	// Dimensions
	Width  int
	Height int

	ripgrepInput textinput.Model // ← New
	filteredView bool
}

func NewModel(client *cloudwatchlogs.Client, env string) Model {

	rgInput := textinput.New()
	rgInput.Placeholder = "ripgrep pattern (e.g., ERROR|WARN)"
	rgInput.Focus()
	rgInput.CharLimit = 100
	rgInput.Width = 50

	sp := spinner.New()
	sp.Spinner = spinner.Dot
	sp.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))

	vp := viewport.New(160, 80) // ← Initialize with default size
	vp.YPosition = 0

	return Model{
		client:       client,
		state:        stateLogGroupList,
		spinner:      sp,
		viewport:     vp, // ← Add initialized viewport
		ripgrepInput: rgInput,
		env:          env,
	}
}

func (m Model) Init() tea.Cmd {
	return tea.Batch(
		m.loadLambdaLogGroups(),
		m.spinner.Tick,
	)
}
