package browser

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"

	"github.com/g0p43r/tui_psql/internal/ui/styles"
)

func (m Model) View() string {
	usableWidth := maxInt(80, m.width-4)
	usableHeight := maxInt(12, m.height-2)
	paneContentHeight := maxInt(1, usableHeight-4)

	leftWidth := maxInt(28, usableWidth/4)
	if leftWidth > 36 {
		leftWidth = 36
	}
	rightWidth := maxInt(48, usableWidth-leftWidth-2)

	leftContent := lipgloss.NewStyle().Height(paneContentHeight).Render(m.leftPane())
	rightContent := lipgloss.NewStyle().Height(paneContentHeight).Render(m.rightPane(rightWidth))

	leftStyle := styles.Panel.Width(leftWidth)
	rightStyle := styles.Panel.Width(rightWidth)
	if m.focus == FocusTables && !m.expanded {
		leftStyle = leftStyle.BorderForeground(lipgloss.Color("212"))
	}
	if m.focus == FocusPreview && !m.expanded {
		rightStyle = rightStyle.BorderForeground(lipgloss.Color("212"))
	}

	left := leftStyle.Render(leftContent)
	right := rightStyle.Render(rightContent)
	layout := lipgloss.JoinHorizontal(lipgloss.Top, left, right)
	content := styles.App.Render(layout)

	if m.expanded {
		content = styles.App.Render(m.expandedRecordView(usableWidth, usableHeight))
	}

	if m.width > 0 && m.height > 0 {
		return lipgloss.Place(m.width, m.height, lipgloss.Left, lipgloss.Top, content)
	}

	return content
}

func (m Model) leftPane() string {
	var lines []string
	lines = append(lines, styles.Title.Render("Tables"))
	lines = append(lines, styles.Subtitle.Render(m.status))
	lines = append(lines, "")

	if len(m.tables) == 0 {
		lines = append(lines, styles.Subtitle.Render("No tables available."))
	} else {
		for i, table := range m.tables {
			prefix := "  "
			if i == m.selected {
				prefix = "> "
			}

			line := fmt.Sprintf("%s%s.%s", prefix, table.Schema, table.Name)
			if i == m.selected {
				line = lipgloss.NewStyle().Foreground(lipgloss.Color("212")).Bold(true).Render(line)
			}
			lines = append(lines, line)
		}
	}

	lines = append(lines, "")
	lines = append(lines, styles.Help.Render("Tab: switch pane  Up/Down: move"))
	return strings.Join(lines, "\n")
}

func (m Model) rightPane(width int) string {
	var lines []string
	lines = append(lines, styles.Title.Render("Table Preview"))
	lines = append(lines, styles.Subtitle.Render(m.previewStatus))
	lines = append(lines, "")

	if m.previewError != "" {
		lines = append(lines, lipgloss.NewStyle().Foreground(lipgloss.Color("203")).Render(m.previewError))
		lines = append(lines, "")
		lines = append(lines, styles.Help.Render("Ctrl+C: quit"))
		return strings.Join(lines, "\n")
	}

	if m.previewTable != nil {
		lines = append(lines, styles.Subtitle.Render(fmt.Sprintf("Selected: %s.%s", m.previewTable.Schema, m.previewTable.Name)))
		lines = append(lines, "")
	}

	if len(m.preview.Columns) == 0 {
		lines = append(lines, styles.Subtitle.Render("No rows loaded yet."))
		lines = append(lines, "")
		lines = append(lines, styles.Help.Render("Ctrl+C: quit"))
		return strings.Join(lines, "\n")
	}

	lines = append(lines, renderTable(m.preview, width-6, m.selectedRow)...)
	lines = append(lines, "")
	lines = append(lines, styles.Help.Render("Tab: switch pane  Up/Down: row  Enter: open record"))
	return strings.Join(lines, "\n")
}
