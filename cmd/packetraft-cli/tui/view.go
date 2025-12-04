package tui

import (
	"fmt"
	"strings"
)

func (m Model) View() string {
	switch m.screen {
	case screenLoading:
		return sectionPad.Render(titleStyle.Render("Loading status…") + m.spin.View())
	case screenSelectGame:
		header := titleStyle.Render("PacketRaft") + "\n" + subtleStyle.Render("Select a game and press Enter")
		return sectionPad.Render(header + "\n\n" + m.gameList.View())
	case screenSelectServer:
		header := titleStyle.Render("Servers for: "+m.selectedGame) + "\n" + subtleStyle.Render("Select a server and press Enter (Backspace to go back)")
		return sectionPad.Render(header + "\n\n" + m.srvList.View())
	case screenGenerate:
		return sectionPad.Render(titleStyle.Render("Preparing config & forwarders… ") + m.spin.View())
	case screenRunning:
		var b strings.Builder
		fmt.Fprintf(&b, "%s\n", titleStyle.Render("Ready"))
		if !strings.Contains(m.tmpConfPath, "WireGuard") {
			fmt.Fprintf(&b, "%s %s\n", okStyle.Render("Temp config:"), codeStyle.Render(m.tmpConfPath))
		} else {
			fmt.Fprintf(&b, "%s %s\n", okStyle.Render("Temp config:"), codeStyle.Render(m.tmpConfPath))
			fmt.Fprintf(&b, "%s\n", okStyle.Render("Config file appended to WireGuard configuration folder!"))
		}
		fmt.Fprintf(&b, "%s %s\n", subtleStyle.Render("Local forward:"), "127.0.0.1:"+m.localPort+" → "+m.remoteAddr)

		fmt.Fprintf(&b, "\n%s", titleStyle.Render("You MUST route these IPs through your default interface before enabling Wireguard:"))
		fmt.Fprintf(&b, "\n%s (%s)\n", okStyle.Render(" - "+strings.Split(m.remoteAddr, ":")[0]), okStyle.Render("real Wireguard endpoint"))

		fmt.Fprintf(&b, "\n%s\n", subtleStyle.Render("Keys: q to quit"))
		return sectionPad.Render(b.String())
	case screenError:
		var b strings.Builder
		fmt.Fprintf(&b, "%s\n", errStyle.Render("Error"))
		if m.err != nil {
			fmt.Fprintf(&b, "%s\n", m.err.Error())
		}
		fmt.Fprintf(&b, "\n%s\n", subtleStyle.Render("Press Backspace to retry or q to quit"))
		return sectionPad.Render(b.String())
	default:
		return ""
	}
}
