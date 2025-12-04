package tui

import (
	"context"
	"time"

	"github.com/b00tkitism/packetraft/api"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
)

type Model struct {
	ctx    context.Context
	cancel context.CancelFunc

	apiClient *api.API

	// data
	status         api.Status
	selectedGame   string
	selectedServer string

	remoteAddr     string
	localPort      string
	config         string
	replacedConfig string
	tmpConfPath    string

	// ui state
	screen    Screen
	gameList  list.Model
	srvList   list.Model
	spin      spinner.Model
	err       error
	width     int
	height    int
	startedAt time.Time
}

func NewModel(ctx context.Context, apiClient *api.API) Model {
	cctx, cancel := context.WithCancel(ctx)

	// game list
	gl := list.New([]list.Item{}, list.NewDefaultDelegate(), 0, 0)
	gl.Title = "Select a game"
	gl.SetShowStatusBar(false)
	gl.SetFilteringEnabled(true)

	// server list
	sl := list.New([]list.Item{}, list.NewDefaultDelegate(), 0, 0)
	sl.Title = "Select a server"
	sl.SetShowStatusBar(false)
	sl.SetFilteringEnabled(true)

	sp := spinner.New()
	sp.Spinner = spinner.Dot

	return Model{
		ctx:       cctx,
		cancel:    cancel,
		apiClient: apiClient,
		screen:    screenLoading,
		gameList:  gl,
		srvList:   sl,
		spin:      sp,
	}
}

func (m Model) Init() tea.Cmd {
	return tea.Batch(m.spin.Tick, m.loadStatus())
}
