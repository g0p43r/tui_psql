package browser

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"

	"github.com/g0p43r/tui_psql/internal/ui/styles"
)

func (m Model) historyView(width, height int) string {
	lines := []string{
		styles.Title.Render("SQL History"),
		styles.Subtitle.Render("Enter: reuse query  Esc: close"),
		"",
	}

	if len(m.history) == 0 {
		lines = append(lines, styles.Subtitle.Render("History is empty."))
	} else {
		maxRows := maxInt(1, m.historyViewportRows())
		end := minInt(len(m.history), m.historyOffset+maxRows)
		for i := m.historyOffset; i < end; i++ {
			entry := m.history[i]
			prefix := "  "
			if i == m.historySelected {
				prefix = "> "
			}
			state := "ERR"
			if entry.Success {
				state = "OK "
			}
			firstLine := strings.Split(entry.SQL, "\n")[0]
			row := fmt.Sprintf("%s[%s] (%s) %s", prefix, state, strings.ToUpper(string(entry.QueryType)), firstLine)
			if i == m.historySelected {
				row = lipgloss.NewStyle().Foreground(lipgloss.Color("212")).Bold(true).Render(row)
			}
			lines = append(lines, row)
		}
	}

	return lipgloss.NewStyle().
		Width(maxInt(40, width-8)).
		Height(maxInt(10, height-4)).
		Border(lipgloss.DoubleBorder()).
		BorderForeground(lipgloss.Color("212")).
		Padding(1, 2).
		Background(lipgloss.Color("235")).
		Render(strings.Join(lines, "\n"))
}
