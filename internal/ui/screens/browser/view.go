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
	frameWidth := styles.Panel.GetHorizontalFrameSize()
	frameHeight := styles.Panel.GetVerticalFrameSize()
	paneContentHeight := maxInt(1, usableHeight-frameHeight)

	leftWidth := maxInt(28, usableWidth/4)
	if leftWidth > 36 {
		leftWidth = 36
	}
	rightWidth := maxInt(48, usableWidth-leftWidth-2)

	leftContent := m.leftPane(maxInt(8, leftWidth-frameWidth), paneContentHeight)
	rightContent := m.rightPane(maxInt(20, rightWidth-frameWidth), paneContentHeight)

	leftStyle := styles.Panel.Width(leftWidth)
	rightStyle := styles.Panel.Width(rightWidth)
	if m.focus == FocusTables && m.IsBrowsing() {
		leftStyle = leftStyle.BorderForeground(lipgloss.Color("212"))
	}
	if m.focus == FocusPreview && m.IsBrowsing() {
		rightStyle = rightStyle.BorderForeground(lipgloss.Color("212"))
	}

	left := leftStyle.Render(leftContent)
	right := rightStyle.Render(rightContent)
	layout := lipgloss.JoinHorizontal(lipgloss.Top, left, right)
	content := styles.App.Render(layout)

	if m.IsEditorOpen() {
		content = styles.App.Render(m.sqlEditorView(usableWidth, usableHeight))
	}

	if m.IsRecordOpen() {
		content = styles.App.Render(m.expandedRecordView(usableWidth, usableHeight))
	}

	if m.width > 0 && m.height > 0 {
		return lipgloss.Place(m.width, m.height, lipgloss.Left, lipgloss.Top, content)
	}

	return content
}

func (m Model) leftPane(width, height int) string {
	var lines []string
	lines = append(lines, styles.Title.Render(fitCell("Tables", width)))
	lines = append(lines, styles.Subtitle.Render(fitCell(m.status, width)))
	lines = append(lines, "")

	if len(m.tables) == 0 {
		lines = append(lines, styles.Subtitle.Render(fitCell("No tables available.", width)))
	} else {
		for i, table := range m.tables {
			prefix := "  "
			if i == m.selected {
				prefix = "> "
			}

			line := fitCell(fmt.Sprintf("%s%s.%s", prefix, table.Schema, table.Name), width)
			if i == m.selected {
				line = lipgloss.NewStyle().Foreground(lipgloss.Color("212")).Bold(true).Render(line)
			}
			lines = append(lines, line)
		}
	}

	lines = append(lines, "")
	lines = append(lines, styles.Help.Render(fitCell("Tab: switch pane  Up/Down: move", width)))
	return fitPaneContent(lines, height)
}

func (m Model) rightPane(width, height int) string {
	var lines []string
	lines = append(lines, styles.Title.Render(fitCell("Table Preview", width)))
	lines = append(lines, styles.Subtitle.Render(fitCell(m.previewStatus, width)))
	lines = append(lines, "")

	if m.previewError != "" {
		lines = append(lines, lipgloss.NewStyle().Foreground(lipgloss.Color("203")).Render(m.previewError))
		lines = append(lines, "")
		lines = append(lines, styles.Help.Render(fitCell("Ctrl+C: quit", width)))
		return strings.Join(lines, "\n")
	}

	if m.previewTable != nil {
		lines = append(lines, styles.Subtitle.Render(fitCell(fmt.Sprintf("Selected: %s.%s", m.previewTable.Schema, m.previewTable.Name), width)))
		lines = append(lines, "")
	}

	if len(m.preview.Columns) == 0 {
		lines = append(lines, styles.Subtitle.Render(fitCell("No rows loaded yet.", width)))
		lines = append(lines, "")
		lines = append(lines, styles.Help.Render(fitCell("Ctrl+C: quit", width)))
		return strings.Join(lines, "\n")
	}

	tableHeight := maxInt(4, height-10)
	lines = append(lines, renderTable(m.preview, width, tableHeight, m.selectedRow, m.rowOffset, m.colOffset)...)
	lines = append(lines, "")
	lines = append(lines, styles.Help.Render(fitCell("Tab switch, Up/Down rows, Left/Right cols, Enter record", width)))
	lines = append(lines, styles.Help.Render(fitCell("F2/F3/F4 SQL, Ctrl+P profiles, Ctrl+X disconnect, Ctrl+R reconnect", width)))
	return fitPaneContent(lines, height)
}

func fitPaneContent(lines []string, height int) string {
	if height <= 0 {
		return ""
	}

	if len(lines) > height {
		lines = lines[:height]
	} else {
		for len(lines) < height {
			lines = append(lines, "")
		}
	}

	return strings.Join(lines, "\n")
}
