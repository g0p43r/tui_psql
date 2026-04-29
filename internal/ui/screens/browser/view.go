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
	hotkeysHeight := m.hotkeysAreaHeight()
	mainHeight := m.mainAreaHeight()
	frameWidth := styles.Panel.GetHorizontalFrameSize()
	frameHeight := styles.Panel.GetVerticalFrameSize()
	paneContentHeight := maxInt(1, mainHeight-frameHeight)

	leftWidth := int(float64(usableWidth) * 0.24)
	leftWidth = maxInt(22, leftWidth)
	if leftWidth > 34 {
		leftWidth = 34
	}
	rightWidth := maxInt(40, usableWidth-leftWidth-2)

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
	panes := lipgloss.JoinHorizontal(lipgloss.Top, left, right)
	hotkeys := m.hotkeysBlock(usableWidth, hotkeysHeight)
	layout := lipgloss.JoinVertical(lipgloss.Left, panes, hotkeys)
	content := styles.App.Render(layout)

	if m.IsEditorOpen() {
		content = styles.App.Render(m.sqlEditorView(usableWidth, usableHeight))
	}

	if m.IsQueryWorkbenchOpen() {
		content = m.queryWorkbenchView(usableWidth, m.height)
	}

	if m.IsHistoryOpen() {
		content = styles.App.Render(m.historyView(usableWidth, usableHeight))
	}

	if m.IsRecordOpen() {
		content = styles.App.Render(m.expandedRecordView(usableWidth, usableHeight))
	}

	if m.width > 0 && m.height > 0 {
		return renderScrollable(content, m.width, m.height, m.layoutOffset)
	}

	return content
}

func (m Model) leftPane(width, height int) string {
	var lines []string
	lines = append(lines, styles.Title.Render("Tables"))
	lines = append(lines, styles.Subtitle.Render(strings.Join(wrapText(m.status, width), "\n")))
	lines = append(lines, "")

	if len(m.tables) == 0 {
		lines = append(lines, styles.Subtitle.Render(fitCell("No tables available.", width)))
	} else {
		maxRows := maxInt(1, height-3)
		end := minInt(len(m.tables), m.tableOffset+maxRows)
		for i := m.tableOffset; i < end; i++ {
			table := m.tables[i]
			prefix := "  "
			if i == m.selected {
				prefix = "> "
			}

			full := fmt.Sprintf("%s%s.%s", prefix, table.Schema, table.Name)
			wrapped := wrapText(full, width)
			for j, part := range wrapped {
				line := part
				if j > 0 {
					line = "  " + part
				}
				if i == m.selected {
					line = lipgloss.NewStyle().Foreground(lipgloss.Color("212")).Bold(true).Render(line)
				}
				lines = append(lines, line)
			}
		}
	}

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
	}

	if len(m.preview.Columns) == 0 {
		lines = append(lines, styles.Subtitle.Render(fitCell("No rows loaded yet.", width)))
		lines = append(lines, "")
		lines = append(lines, styles.Help.Render(fitCell("Ctrl+Q: quit", width)))
		return strings.Join(lines, "\n")
	}

	tableHeight := maxInt(4, height-10)
	lines = append(lines, renderTable(m.preview, width, tableHeight, m.selectedRow, m.rowOffset, m.colOffset)...)
	return fitPaneContent(lines, height)
}

func (m Model) hotkeysBlock(width, height int) string {
	contentWidth := maxInt(20, width-styles.Panel.GetHorizontalFrameSize())
	var lines []string
	lines = append(lines, styles.Help.Render(fitCell("Table: Ctrl+C create  Ctrl+U alter  Ctrl+D drop    Row: Ctrl+C create  Ctrl+U update  Ctrl+D delete", contentWidth)))
	lines = append(lines, styles.Help.Render(fitCell("SQL: Alt+Enter execute  Ctrl+T type    Nav: Tab pane  Arrows move  PgUp/PgDn page  Home/End jump  Enter/Esc", contentWidth)))
	lines = append(lines, styles.Help.Render(fitCell("System: Ctrl+P profiles  Ctrl+H history  Ctrl+X disconnect  Ctrl+R reconnect  Alt+Up/Alt+Down scroll  Ctrl+Q quit", contentWidth)))
	contentHeight := maxInt(1, height-styles.Panel.GetVerticalFrameSize())
	return styles.Panel.Width(width).Render(fitPaneContent(lines, contentHeight))
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

func renderScrollable(content string, width, height, offset int) string {
	lines := strings.Split(content, "\n")
	if height <= 0 {
		return ""
	}
	if len(lines) <= height {
		return lipgloss.Place(width, height, lipgloss.Left, lipgloss.Top, content)
	}

	maxOffset := len(lines) - height
	if offset < 0 {
		offset = 0
	}
	if offset > maxOffset {
		offset = maxOffset
	}

	visible := strings.Join(lines[offset:offset+height], "\n")
	return lipgloss.Place(width, height, lipgloss.Left, lipgloss.Top, visible)
}
