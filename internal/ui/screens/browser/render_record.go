package browser

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"

	"github.com/g0p43r/tui_psql/internal/ui/styles"
)

func (m Model) expandedRecordView(width, height int) string {
	var lines []string
	lines = append(lines, styles.Title.Render("Record Details"))
	lines = append(lines, "")

	if m.previewTable != nil {
		lines = append(lines, styles.Subtitle.Render(fmt.Sprintf("%s.%s", m.previewTable.Schema, m.previewTable.Name)))
		lines = append(lines, "")
	}

	row, ok := m.SelectedPreviewRow()
	if !ok {
		lines = append(lines, styles.Subtitle.Render("No row selected."))
	} else {
		for i, col := range m.preview.Columns {
			value := ""
			if i < len(row) {
				value = row[i]
			}
			lines = append(lines, renderField(col, value, width-12)...)
		}
	}

	lines = append(lines, "")
	lines = append(lines, styles.Help.Render("Esc or Enter: close"))

	return lipgloss.NewStyle().
		Width(maxInt(40, width-8)).
		Height(maxInt(10, height-4)).
		Border(lipgloss.DoubleBorder()).
		BorderForeground(lipgloss.Color("212")).
		Padding(1, 2).
		Background(lipgloss.Color("235")).
		Render(strings.Join(lines, "\n"))
}

func renderField(name, value string, width int) []string {
	label := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("110")).
		Render(name + ":")

	prefixWidth := runeLen(name) + 2
	if prefixWidth > width/2 {
		prefixWidth = maxInt(10, width/3)
	}

	valueWidth := maxInt(12, width-prefixWidth-1)
	wrapped := wrapText(strings.TrimSpace(value), valueWidth)
	if len(wrapped) == 0 {
		wrapped = []string{""}
	}

	lines := make([]string, 0, len(wrapped))
	for i, part := range wrapped {
		if i == 0 {
			lines = append(lines, label+" "+part)
			continue
		}
		lines = append(lines, strings.Repeat(" ", prefixWidth)+" "+part)
	}

	return lines
}
