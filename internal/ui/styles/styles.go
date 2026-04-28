package styles

import "github.com/charmbracelet/lipgloss"

var (
	App = lipgloss.NewStyle().
		Padding(1, 2)

	Title = lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("212"))

	Subtitle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("245"))

	Label = lipgloss.NewStyle().
		Width(12).
		Foreground(lipgloss.Color("110"))

	Panel = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("240")).
		Padding(1, 2)

	Help = lipgloss.NewStyle().
		Foreground(lipgloss.Color("242"))
)
