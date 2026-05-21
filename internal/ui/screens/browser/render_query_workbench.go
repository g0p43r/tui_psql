package browser

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"

	"github.com/g0p43r/tui_psql/internal/ui/styles"
)

func (m Model) queryWorkbenchView(width, height int) string {
	layout := newQueryWorkbenchLayout(width, height)

	topContent := []string{
		styles.Title.Render(fitCell("SQL Query", layout.contentWidth)),
		styles.Subtitle.Render(fitCell(fmt.Sprintf("Type: %s  Alt+Enter execute  Ctrl+T type  Alt+Arrows result  Esc back", strings.ToUpper(string(m.EditorType()))), layout.contentWidth)),
		"",
	}
	topContent = append(topContent, strings.Split(fitSQLBlock(m.editor.View(), layout.contentWidth, layout.editorHeight), "\n")...)
	top := styles.Panel.
		Width(layout.panelWidth).
		Height(layout.topStyleHeight).
		Render(fitPaneContent(topContent, layout.topContentHeight))

	output := fitSQLBlock(m.editorOutputView(layout.contentWidth, layout.resultHeight), layout.contentWidth, layout.resultHeight)
	bottomContent := []string{
		styles.Title.Render("Result"),
		"",
	}
	bottomContent = append(bottomContent, strings.Split(output, "\n")...)
	bottom := styles.Panel.
		Width(layout.panelWidth).
		Height(layout.bottomStyleHeight).
		Render(fitPaneContent(bottomContent, layout.bottomContentHeight))

	return lipgloss.JoinVertical(lipgloss.Left, top, bottom)
}

type queryWorkbenchLayout struct {
	panelWidth          int
	contentWidth        int
	topStyleHeight      int
	bottomStyleHeight   int
	topContentHeight    int
	bottomContentHeight int
	editorHeight        int
	resultHeight        int
}

func newQueryWorkbenchLayout(width, height int) queryWorkbenchLayout {
	mainHeight := maxInt(12, height)
	panelWidth := maxInt(1, width)
	contentWidth := maxInt(1, width-styles.Panel.GetHorizontalFrameSize())
	topOuterHeight := queryWorkbenchTopHeight(mainHeight)
	bottomOuterHeight := maxInt(6, mainHeight-topOuterHeight)

	borderHeight := styles.Panel.GetVerticalBorderSize()
	paddingHeight := styles.Panel.GetVerticalPadding()
	topStyleHeight := maxInt(1, topOuterHeight-borderHeight)
	bottomStyleHeight := maxInt(1, bottomOuterHeight-borderHeight)
	topContentHeight := maxInt(1, topStyleHeight-paddingHeight)
	bottomContentHeight := maxInt(1, bottomStyleHeight-paddingHeight)

	return queryWorkbenchLayout{
		panelWidth:          panelWidth,
		contentWidth:        contentWidth,
		topStyleHeight:      topStyleHeight,
		bottomStyleHeight:   bottomStyleHeight,
		topContentHeight:    topContentHeight,
		bottomContentHeight: bottomContentHeight,
		editorHeight:        maxInt(1, topContentHeight-3),
		resultHeight:        maxInt(1, bottomContentHeight-2),
	}
}

func queryWorkbenchTopHeight(height int) int {
	maxTopHeight := maxInt(1, height-6)
	topHeight := int(float64(height) * 0.20)
	topHeight = maxInt(7, topHeight)
	return minInt(topHeight, maxTopHeight)
}
