package browser

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"

	"github.com/g0p43r/tui_psql/internal/domain"
	"github.com/g0p43r/tui_psql/internal/ui/styles"
)

func renderTable(result domain.QueryResult, width int, selectedRow int) []string {
	if len(result.Columns) == 0 {
		return []string{"No data."}
	}

	if len(result.Rows) == 0 {
		return []string{"No rows returned."}
	}

	if width < 24 {
		return []string{"Not enough space to render table preview."}
	}

	layout := buildColumnLayout(result, width)
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

	for rowIndex, row := range result.Rows {
		rowLine := renderRow(row, layout)
		if rowIndex == selectedRow {
			rowLine = lipgloss.NewStyle().Foreground(lipgloss.Color("212")).Bold(true).Render(rowLine)
		}
		lines = append(lines, rowLine)
	}

	if layout.hiddenColumns > 0 {
		lines = append(lines, "")
		lines = append(lines, styles.Subtitle.Render(fmt.Sprintf("+ %d more columns hidden", layout.hiddenColumns)))
	}

	return lines
}

func buildColumnLayout(result domain.QueryResult, totalWidth int) columnLayout {
	const minColumnWidth = 8
	const maxColumnWidth = 28
	const separatorWidth = 3

	var layout columnLayout
	usedWidth := 0

	for colIndex, colName := range result.Columns {
		columnWidth := runeLen(colName)
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

		extraWidth := columnWidth
		if len(layout.widths) > 0 {
			extraWidth += separatorWidth
		}

		if usedWidth+extraWidth > totalWidth {
			layout.hiddenColumns = len(result.Columns) - len(layout.widths)
			break
		}

		layout.indexes = append(layout.indexes, colIndex)
		layout.widths = append(layout.widths, columnWidth)
		usedWidth += extraWidth
	}

	if len(layout.widths) == 0 && len(result.Columns) > 0 {
		layout.indexes = []int{0}
		layout.widths = []int{maxInt(minColumnWidth, minInt(maxColumnWidth, totalWidth))}
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
