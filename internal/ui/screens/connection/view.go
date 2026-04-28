package connection

import (
	"strings"

	"github.com/charmbracelet/lipgloss"

	"github.com/g0p43r/tui_psql/internal/ui/styles"
)

func (m Model) View() string {
	var lines []string
	lines = append(lines, styles.Title.Render("PostgreSQL Connection"))
	lines = append(lines, styles.Subtitle.Render("First TUI screen for connection setup"))
	lines = append(lines, "")

	formWidth := 56
	for _, f := range m.fields {
		line := lipgloss.JoinHorizontal(
			lipgloss.Top,
			styles.Label.Render(f.label),
			f.input.View(),
		)
		lines = append(lines, line)
	}

	lines = append(lines, "")
	lines = append(lines, m.renderStatus())
	lines = append(lines, "")
	lines = append(lines, styles.Help.Render("Tab/Shift+Tab: move  Enter: next  Ctrl+C: quit"))

	panel := styles.Panel.Width(formWidth).Render(strings.Join(lines, "\n"))
	content := styles.App.Render(panel)

	if m.width > 0 && m.height > 0 {
		return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, content)
	}

	return content
}
