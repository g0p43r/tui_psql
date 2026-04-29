package browser

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"

	"github.com/g0p43r/tui_psql/internal/domain"
	"github.com/g0p43r/tui_psql/internal/ui/styles"
)

func renderTable(result domain.QueryResult, width, height, selectedRow, rowOffset, colOffset int) []string {
	if len(result.Columns) == 0 {
		return []string{"No data."}
	}

	if len(result.Rows) == 0 {
		return []string{"No rows returned."}
	}

	if width < 24 {
		return []string{"Not enough space to render table preview."}
	}

	layout := buildColumnLayout(result, width, colOffset)
	headerCells := make([]string, 0, len(layout.widths))
	for i, colIndex := range layout.indexes {
		headerCells = append(
			headerCells,
			lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("230")).Render(padCell(result.Columns[colIndex], layout.widths[i])),
		)
	}

	headerLine := strings.Join(headerCells, " │ ")
	lines := []string{headerLine}
	lines = append(lines, strings.Repeat("─", minInt(width, visibleLen(headerLine))))

	visibleRows := maxInt(1, height-3)
	endRow := minInt(len(result.Rows), rowOffset+visibleRows)
	for rowIndex := rowOffset; rowIndex < endRow; rowIndex++ {
		row := result.Rows[rowIndex]
		rowLine := renderRow(row, layout)
		if rowIndex == selectedRow {
			rowLine = lipgloss.NewStyle().Foreground(lipgloss.Color("212")).Bold(true).Render(rowLine)
		}
		lines = append(lines, rowLine)
	}

	statusParts := []string{
		fmt.Sprintf("rows %d-%d/%d", rowOffset+1, endRow, len(result.Rows)),
		fmt.Sprintf("cols %d-%d/%d", layout.firstColumn+1, layout.lastColumn+1, len(result.Columns)),
		fmt.Sprintf("offset %d/%d %s", layout.scrollOffset, layout.maxOffset, horizontalBar(layout.scrollOffset, layout.maxOffset, 10)),
	}
	if result.Truncated && result.Limit > 0 {
		statusParts = append(statusParts, fmt.Sprintf("limited to %d rows", result.Limit))
	}
	lines = append(lines, strings.Repeat("─", width))
	lines = append(lines, styles.Subtitle.Render(strings.Join(statusParts, "  ")))

	if layout.hiddenColumns > 0 || layout.firstColumn > 0 {
		lines = append(lines, "")
		leftHidden := layout.firstColumn
		rightHidden := layout.hiddenColumns
		lines = append(lines, styles.Subtitle.Render(fmt.Sprintf("hidden columns: left %d, right %d", leftHidden, rightHidden)))
	}

	return lines
}

func buildColumnLayout(result domain.QueryResult, totalWidth int, offset int) columnLayout {
	const minColumnWidth = 8
	const maxColumnWidth = 28
	const separatorWidth = 3

	var layout columnLayout
	if len(result.Columns) == 0 {
		return layout
	}

	if offset < 0 {
		offset = 0
	}
	maxOffset := maxInt(0, len(result.Columns)-2)
	if offset > maxOffset {
		offset = maxOffset
	}

	layout.scrollOffset = offset
	layout.maxOffset = maxOffset
	layout.firstColumn = 0
	usedWidth := 0

	// Sticky first column is always rendered when present.
	firstWidth := measureColumnWidth(result, 0, minColumnWidth, maxColumnWidth)
	firstWidth = minInt(firstWidth, maxInt(minColumnWidth, totalWidth/2))
	layout.indexes = append(layout.indexes, 0)
	layout.widths = append(layout.widths, firstWidth)
	usedWidth += firstWidth

	startCol := offset + 1
	for colIndex := startCol; colIndex < len(result.Columns); colIndex++ {
		columnWidth := measureColumnWidth(result, colIndex, minColumnWidth, maxColumnWidth)

		extraWidth := columnWidth
		if len(layout.widths) > 0 {
			extraWidth += separatorWidth
		}

		if usedWidth+extraWidth > totalWidth {
			layout.hiddenColumns = len(result.Columns) - colIndex
			break
		}

		layout.indexes = append(layout.indexes, colIndex)
		layout.widths = append(layout.widths, columnWidth)
		usedWidth += extraWidth
	}

	if len(layout.indexes) > 0 {
		layout.lastColumn = layout.indexes[len(layout.indexes)-1]
	}
	if layout.lastColumn == 0 && len(result.Columns) > 1 {
		layout.hiddenColumns = len(result.Columns) - 1
	}

	return layout
}

func renderRow(row []string, layout columnLayout) string {
	cells := make([]string, 0, len(layout.widths))
	for i, colIndex := range layout.indexes {
		value := ""
		if colIndex < len(row) {
			value = row[colIndex]
		}
		cells = append(cells, fitCell(value, layout.widths[i]))
	}
	return strings.Join(cells, " │ ")
}

func measureColumnWidth(result domain.QueryResult, colIndex, minColumnWidth, maxColumnWidth int) int {
	columnWidth := runeLen(result.Columns[colIndex])
	for _, row := range result.Rows {
		if colIndex >= len(row) {
			continue
		}
		cellWidth := runeLen(row[colIndex])
		if cellWidth > columnWidth {
			columnWidth = cellWidth
		}
	}

	if columnWidth < minColumnWidth {
		columnWidth = minColumnWidth
	}
	if columnWidth > maxColumnWidth {
		columnWidth = maxColumnWidth
	}
	return columnWidth
}

func horizontalBar(offset, maxOffset, width int) string {
	if width <= 0 {
		return ""
	}
	if maxOffset <= 0 {
		return "[" + strings.Repeat("=", width) + "]"
	}

	pos := int(float64(offset) / float64(maxOffset) * float64(width-1))
	if pos < 0 {
		pos = 0
	}
	if pos >= width {
		pos = width - 1
	}

	cells := make([]rune, width)
	for i := range cells {
		cells[i] = '·'
	}
	cells[pos] = '■'
	return "[" + string(cells) + "]"
}
