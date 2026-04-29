package browser

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"

	"github.com/g0p43r/tui_psql/internal/ui/styles"
)

func (m Model) queryWorkbenchView(width, height int) string {
	mainHeight := maxInt(12, height)
	topHeight := queryWorkbenchTopHeight(mainHeight)
	bottomHeight := maxInt(6, mainHeight-topHeight)

	topContent := []string{
		styles.Title.Render("SQL Query"),
		styles.Subtitle.Render(fmt.Sprintf("Type: %s  Alt+Enter execute  Ctrl+T type  Alt+Arrows result  Esc back", strings.ToUpper(string(m.EditorType())))),
		"",
		m.editor.View(),
	}
	top := styles.Panel.
		Width(width).
		Height(topHeight).
		Render(fitPaneContent(topContent, maxInt(1, topHeight-styles.Panel.GetVerticalFrameSize())))

	output := m.editorOutputView(width-8, bottomHeight-styles.Panel.GetVerticalFrameSize())
	bottomContent := []string{
		styles.Title.Render("Result"),
		"",
		output,
	}
	bottom := styles.Panel.
		Width(width).
		Height(bottomHeight).
		Render(fitPaneContent(bottomContent, maxInt(1, bottomHeight-styles.Panel.GetVerticalFrameSize())))

	return lipgloss.JoinVertical(lipgloss.Left, top, bottom)
}

func queryWorkbenchTopHeight(height int) int {
	topHeight := int(float64(height) * 0.20)
	topHeight = maxInt(7, topHeight)
	return minInt(topHeight, maxInt(7, height-6))
}
