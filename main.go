package main

import (
	"log"

	"leetcli/internal/api"
	"leetcli/internal/config"
	"leetcli/internal/service"
	"leetcli/internal/ui"

	tea "github.com/charmbracelet/bubbletea"
)

func main() {

	cfg, err := config.Load()
	if err != nil {
		log.Fatal("create ~/.leetcli.yaml first")
	}

	client := api.NewClient()
	service := service.NewProfileService(client, cfg.Username)

	model := ui.NewModel(service)

	p := tea.NewProgram(model, tea.WithAltScreen())

	if err := p.Start(); err != nil {
		log.Fatal(err)
	}
}