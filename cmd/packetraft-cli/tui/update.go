package tui

import (
	"time"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
)

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	case tea.WindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height
		m.gameList.SetSize(m.width-4, m.height-8)
		m.srvList.SetSize(m.width-4, m.height-8)
		return m, nil

	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spin, cmd = m.spin.Update(msg)
		return m, cmd

	case statusLoadedMsg:
		m.status = msg.status
		// populate game list
		items := make([]list.Item, len(m.status.Games))
		for i, g := range m.status.Games {
			items[i] = item{title: g.Name, desc: "Game", val: g.Name}
		}
		m.gameList.SetItems(items)
		m.screen = screenSelectGame
		return m, nil

	case serversLoadedMsg:
		// servers might be empty; still show "None"
		if len(msg.serverNames) == 0 {
			m.srvList.SetItems([]list.Item{item{title: "(no servers)", desc: "Proceed with empty server", val: ""}})
		} else {
			items := make([]list.Item, len(msg.serverNames))
			for i, s := range msg.serverNames {
				items[i] = item{title: s, desc: "Server", val: s}
			}
			m.srvList.SetItems(items)
		}
		m.screen = screenSelectServer
		return m, nil

	case configReadyMsg:
		m.config = msg.config
		m.remoteAddr = msg.remoteAddr
		m.replacedConfig = msg.replacedConf
		m.tmpConfPath = msg.tmpPath
		m.screen = screenGenerate
		m.localPort = msg.listenPort
		// auto-start forwarder, then show final running screen
		return m, tea.Batch(m.startForwarder(m.remoteAddr))

	case forwarderStartedMsg:
		m.screen = screenRunning
		m.startedAt = time.Now()
		return m, nil

	case errMsg:
		m.err = msg.err
		m.screen = screenError
		return m, nil

	case tea.KeyMsg:
		switch m.screen {

		case screenSelectGame:
			switch msg.String() {
			case "enter":
				if it, ok := m.gameList.SelectedItem().(item); ok {
					m.selectedGame = it.val
					return m, tea.Batch(m.spin.Tick, m.loadServers(m.selectedGame))
				}
			case "q", "esc", "ctrl+c":
				return m, tea.Quit
			}
			var cmd tea.Cmd
			m.gameList, cmd = m.gameList.Update(msg)
			return m, cmd

		case screenSelectServer:
			switch msg.String() {
			case "enter":
				if it, ok := m.srvList.SelectedItem().(item); ok {
					m.selectedServer = it.val
					m.screen = screenGenerate
					return m, tea.Batch(m.spin.Tick, m.genConfig(m.selectedGame, m.selectedServer))
				}
			case "backspace":
				m.screen = screenSelectGame
				return m, nil
			case "q", "esc", "ctrl+c":
				return m, tea.Quit
			}
			var cmd tea.Cmd
			m.srvList, cmd = m.srvList.Update(msg)
			return m, cmd

		case screenGenerate:
			// generating: allow quit or back
			switch msg.String() {
			case "q", "esc", "ctrl+c":
				return m, tea.Quit
			}
			return m, nil

		case screenRunning:
			switch msg.String() {
			case "c":
				// No clipboard by default; keep as no-op placeholder
				return m, nil
			case "q", "esc", "ctrl+c":
				return m, tea.Quit
			}
			return m, nil

		case screenError:
			switch msg.String() {
			case "q", "esc", "ctrl+c":
				return m, tea.Quit
			case "backspace":
				m.screen = screenSelectGame
				return m, nil
			}
			return m, nil
		}
	}
	return m, nil
}
