package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/b00tkitism/packetraft/api"
	"github.com/b00tkitism/packetraft/cmd/packetraft-cli/tui"
	tea "github.com/charmbracelet/bubbletea"
)

func main() {
	// signal-aware context
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, os.Kill, syscall.SIGTERM)
	defer stop()

	apiClient := api.NewAPI()

	p := tea.NewProgram(tui.NewModel(ctx, apiClient), tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		log.Fatal(err)
	}
}
