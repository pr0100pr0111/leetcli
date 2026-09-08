package main

import (
	"log"
	"os"

	"leetcli/internal/api"
	"leetcli/internal/cli"
	"leetcli/internal/config"
	"leetcli/internal/service"
	"leetcli/internal/ui"

	tea "github.com/charmbracelet/bubbletea"
)

func main() {
	if len(os.Args) > 1 {
		os.Exit(cli.Run(os.Args[1:]))
	}

	cfg, err := config.Load()
	if err != nil {
		log.Fatal("create ~/.leetcli.yaml first")
	}

	client := api.NewClient()
	service := service.NewProfileService(client, cfg.Username)

	model := ui.NewModel(service, cfg)

	p := tea.NewProgram(model, tea.WithAltScreen())

	if _, err := p.Run(); err != nil {
		log.Fatal(err)
	}
}
