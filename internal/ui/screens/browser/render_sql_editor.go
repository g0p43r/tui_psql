package browser

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"

	"github.com/g0p43r/tui_psql/internal/domain"
	"github.com/g0p43r/tui_psql/internal/ui/styles"
)

func (m Model) sqlEditorView(width, height int) string {
	title := "SQL Editor"
	switch m.editorMode {
	case editorInsert:
		title = "SQL Editor: INSERT"
	case editorUpdate:
		title = "SQL Editor: UPDATE"
	case editorDelete:
		title = "SQL Editor: DELETE"
	}

	lines := []string{
		styles.Title.Render(title),
		styles.Subtitle.Render("Esc: close  Ctrl+I/Ctrl+U/Ctrl+D: regenerate template"),
		"",
		m.editor.View(),
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

func buildSQLTemplate(mode editorMode, table domain.DBObject, columns []string, row []string) string {
	switch mode {
	case editorInsert:
		return buildInsertTemplate(table, columns)
	case editorUpdate:
		return buildUpdateTemplate(table, columns, row)
	case editorDelete:
		return buildDeleteTemplate(table, columns, row)
	default:
		return ""
	}
}

func buildInsertTemplate(table domain.DBObject, columns []string) string {
	if len(columns) == 0 {
		columns = []string{"column_name"}
	}

	columnLines := make([]string, 0, len(columns))
	valueLines := make([]string, 0, len(columns))
	for _, col := range columns {
		columnLines = append(columnLines, "    "+col)
		valueLines = append(valueLines, "    /* "+col+" */ NULL")
	}

	return fmt.Sprintf(
		"INSERT INTO %s.%s (\n%s\n)\nVALUES (\n%s\n);",
		table.Schema,
		table.Name,
		strings.Join(columnLines, ",\n"),
		strings.Join(valueLines, ",\n"),
	)
}

func buildUpdateTemplate(table domain.DBObject, columns []string, row []string) string {
	if len(columns) == 0 {
		columns = []string{"column_name"}
	}

	setLines := make([]string, 0, len(columns))
	for i, col := range columns {
		current := currentValuePreview(row, i)
		setLines = append(setLines, fmt.Sprintf("    %s = /* current: %s */ NULL", col, current))
	}

	whereColumn := columns[0]
	whereValue := currentValuePreview(row, 0)

	return fmt.Sprintf(
		"UPDATE %s.%s\nSET\n%s\nWHERE\n    %s = /* current: %s */ NULL;",
		table.Schema,
		table.Name,
		strings.Join(setLines, ",\n"),
		whereColumn,
		whereValue,
	)
}

func buildDeleteTemplate(table domain.DBObject, columns []string, row []string) string {
	whereColumn := "id"
	if len(columns) > 0 {
		whereColumn = columns[0]
	}
	whereValue := currentValuePreview(row, 0)

	return fmt.Sprintf(
		"DELETE FROM %s.%s\nWHERE\n    %s = /* current: %s */ NULL;",
		table.Schema,
		table.Name,
		whereColumn,
		whereValue,
	)
}

func currentValuePreview(row []string, index int) string {
	if index >= 0 && index < len(row) {
		value := row[index]
		if len(value) > 40 {
			return value[:40] + "..."
		}
		return value
	}
	return "unknown"
}
