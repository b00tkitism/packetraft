package tui

import (
	"github.com/b00tkitism/packetraft/api"
	"github.com/charmbracelet/lipgloss"
)

// ---------- UI styles ----------
var (
	titleStyle  = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#7D56F4"))
	subtleStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#6B7280"))
	okStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("#04B575")).Bold(true)
	errStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("#FF5F87")).Bold(true)
	codeStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("#F5F5F5")).Background(lipgloss.Color("#3C3C3C")).Padding(0, 1)
	sectionPad  = lipgloss.NewStyle().Padding(1, 2)
)

// ---------- list item ----------
type item struct {
	title string
	desc  string
	val   string
}

func (i item) Title() string       { return i.title }
func (i item) Description() string { return i.desc }
func (i item) FilterValue() string { return i.val }

// ---------- app state ----------
type Screen int

const (
	screenLoading Screen = iota
	screenSelectGame
	screenSelectServer
	screenGenerate
	screenRunning
	screenError
)

type statusLoadedMsg struct {
	status api.Status
}

type serversLoadedMsg struct {
	serverNames []string
}

type configReadyMsg struct {
	config       string
	listenPort   string
	remoteAddr   string
	replacedConf string
	tmpPath      string
}

type forwarderStartedMsg struct{}

type errMsg struct{ err error }
