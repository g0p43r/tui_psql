package main

import (
	"log"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/g0p43r/tui_psql/internal/app"
)

func main() {
	p := tea.NewProgram(app.New(), tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		log.Fatal(err)
	}
}
