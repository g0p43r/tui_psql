package browser

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/g0p43r/tui_psql/internal/domain"
)

type TableSelectedMsg struct {
	Table domain.DBObject
}

type FocusArea string

const (
	FocusTables  FocusArea = "tables"
	FocusPreview FocusArea = "preview"
)

type Model struct {
	tables        []domain.DBObject
	selected      int
	width         int
	height        int
	status        string
	previewStatus string
	previewTable  *domain.DBObject
	preview       domain.QueryResult
	previewError  string
	focus         FocusArea
	selectedRow   int
	expanded      bool
}

func New() Model {
	return Model{
		status:        "No tables loaded.",
		previewStatus: "Select a table to inspect its rows.",
		focus:         FocusTables,
	}
}

func (m *Model) SetSize(width, height int) {
	m.width = width
	m.height = height
}

func (m *Model) SetStatus(status string) {
	m.status = status
}

func (m *Model) SetTables(tables []domain.DBObject) {
	m.tables = tables
	m.selected = 0
	m.preview = domain.QueryResult{}
	m.previewError = ""
	m.selectedRow = 0
	m.expanded = false
	if len(tables) == 0 {
		m.status = "No tables found."
		return
	}

	m.status = fmt.Sprintf("Loaded %d objects.", len(tables))
}

func (m *Model) SetPreviewStatus(status string) {
	m.previewStatus = status
	m.previewError = ""
	m.preview = domain.QueryResult{}
	m.selectedRow = 0
	m.expanded = false
}

func (m *Model) SetPreview(table domain.DBObject, result domain.QueryResult) {
	m.previewTable = &table
	m.preview = result
	m.previewError = ""
	m.selectedRow = 0
	m.expanded = false
	m.previewStatus = fmt.Sprintf("Showing %d rows from %s.%s", len(result.Rows), table.Schema, table.Name)
}

func (m *Model) SetPreviewError(table domain.DBObject, status string) {
	m.previewTable = &table
	m.preview = domain.QueryResult{}
	m.previewError = status
	m.selectedRow = 0
	m.expanded = false
	m.previewStatus = "Failed to load table preview."
}

func (m Model) SelectedTable() (domain.DBObject, bool) {
	if len(m.tables) == 0 || m.selected < 0 || m.selected >= len(m.tables) {
		return domain.DBObject{}, false
	}
	return m.tables[m.selected], true
}

func (m Model) selectCmd() tea.Cmd {
	table, ok := m.SelectedTable()
	if !ok {
		return nil
	}
	return func() tea.Msg {
		return TableSelectedMsg{Table: table}
	}
}

func (m Model) SelectedPreviewRow() ([]string, bool) {
	if len(m.preview.Rows) == 0 || m.selectedRow < 0 || m.selectedRow >= len(m.preview.Rows) {
		return nil, false
	}
	return m.preview.Rows[m.selectedRow], true
}
