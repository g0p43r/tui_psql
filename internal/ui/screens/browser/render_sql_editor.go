package browser

import (
	"fmt"
	"regexp"
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
	case editorCreate:
		title = "SQL Editor: CREATE TABLE"
	case editorAlter:
		title = "SQL Editor: ALTER TABLE"
	case editorDrop:
		title = "SQL Editor: DROP TABLE"
	}

	lines := []string{
		styles.Title.Render(title),
		styles.Subtitle.Render(fmt.Sprintf("Type: %s", strings.ToUpper(string(m.EditorType())))),
		styles.Subtitle.Render("F5/Ctrl+Enter: execute  Ctrl+T: change type  Esc: close"),
		"",
		m.editor.View(),
	}

	if m.editorStatus != "" {
		status := styles.Subtitle.Render(m.editorStatus)
		if m.editorStatusErr {
			status = lipgloss.NewStyle().Foreground(lipgloss.Color("203")).Render(m.editorStatus)
		}
		lines = append(lines, "")
		lines = append(lines, status)
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

func buildSQLTemplate(mode editorMode, table domain.DBObject, columns []string, columnTypes []string, row []string) string {
	switch mode {
	case editorInsert:
		return buildInsertTemplate(table, columns, columnTypes)
	case editorUpdate:
		return buildUpdateTemplate(table, columns, columnTypes, row)
	case editorDelete:
		return buildDeleteTemplate(table, columns, columnTypes, row)
	default:
		return ""
	}
}

func buildCreateTableTemplate(schema string) string {
	return fmt.Sprintf(
		"CREATE TABLE %s.new_table (\n    id uuid PRIMARY KEY,\n    name text NOT NULL,\n    created_at timestamptz NOT NULL DEFAULT now()\n);\n\n-- Example column type notes:\n-- uuid, text, varchar, integer, bigint, numeric, boolean, jsonb, timestamptz",
		schema,
	)
}

func buildAlterTableTemplate(table domain.DBObject) string {
	return fmt.Sprintf(
		"ALTER TABLE %s.%s\n    ADD COLUMN new_column text;\n\n-- Examples:\n-- ALTER TABLE %s.%s RENAME COLUMN old_name TO new_name;\n-- ALTER TABLE %s.%s DROP COLUMN obsolete_column;",
		table.Schema,
		table.Name,
		table.Schema,
		table.Name,
		table.Schema,
		table.Name,
	)
}

func buildDropTableTemplate(table domain.DBObject) string {
	return fmt.Sprintf(
		"DROP TABLE IF EXISTS %s.%s;\n\n-- If the table has dependent objects, use:\n-- DROP TABLE IF EXISTS %s.%s CASCADE;",
		table.Schema,
		table.Name,
		table.Schema,
		table.Name,
	)
}

func buildInsertTemplate(table domain.DBObject, columns []string, columnTypes []string) string {
	if len(columns) == 0 {
		columns = []string{"column_name"}
	}

	columnLines := make([]string, 0, len(columns))
	valueLines := make([]string, 0, len(columns))
	for i, col := range columns {
		colType := columnTypeAt(columnTypes, i)
		columnLines = append(columnLines, "    "+col)
		valueLines = append(valueLines, fmt.Sprintf("    /* %s %s */ NULL", col, colType))
	}

	return fmt.Sprintf(
		"INSERT INTO %s.%s (\n%s\n)\nVALUES (\n%s\n);",
		table.Schema,
		table.Name,
		strings.Join(columnLines, ",\n"),
		strings.Join(valueLines, ",\n"),
	)
}

func buildUpdateTemplate(table domain.DBObject, columns []string, columnTypes []string, row []string) string {
	if len(columns) == 0 {
		columns = []string{"column_name"}
	}

	setLines := make([]string, 0, len(columns))
	for i, col := range columns {
		colType := columnTypeAt(columnTypes, i)
		current := currentValuePreview(row, i)
		setLines = append(setLines, fmt.Sprintf("    %s = /* %s current: %s */ %s", col, colType, current, sqlValue(row, i)))
	}

	whereColumn := columns[0]
	whereType := columnTypeAt(columnTypes, 0)
	whereValue := currentValuePreview(row, 0)

	return fmt.Sprintf(
		"UPDATE %s.%s\nSET\n%s\nWHERE\n    %s = /* %s current: %s */ %s;",
		table.Schema,
		table.Name,
		strings.Join(setLines, ",\n"),
		whereColumn,
		whereType,
		whereValue,
		sqlValue(row, 0),
	)
}

func buildDeleteTemplate(table domain.DBObject, columns []string, columnTypes []string, row []string) string {
	whereColumn := "id"
	if len(columns) > 0 {
		whereColumn = columns[0]
	}
	whereType := columnTypeAt(columnTypes, 0)
	whereValue := currentValuePreview(row, 0)

	return fmt.Sprintf(
		"DELETE FROM %s.%s\nWHERE\n    %s = /* %s current: %s */ %s;",
		table.Schema,
		table.Name,
		whereColumn,
		whereType,
		whereValue,
		sqlValue(row, 0),
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

func columnTypeAt(columnTypes []string, index int) string {
	if index >= 0 && index < len(columnTypes) && columnTypes[index] != "" {
		return columnTypes[index]
	}
	return "unknown"
}

func sqlValue(row []string, index int) string {
	if index < 0 || index >= len(row) {
		return "NULL"
	}

	value := strings.TrimSpace(row[index])
	if value == "" || strings.EqualFold(value, "NULL") {
		return "NULL"
	}
	if strings.EqualFold(value, "true") || strings.EqualFold(value, "false") {
		return strings.ToUpper(value)
	}
	if isNumeric(value) {
		return value
	}

	escaped := strings.ReplaceAll(value, "'", "''")
	return "'" + escaped + "'"
}

var numericPattern = regexp.MustCompile(`^[+-]?(\d+(\.\d+)?|\.\d+)$`)

func isNumeric(value string) bool {
	return numericPattern.MatchString(value)
}
