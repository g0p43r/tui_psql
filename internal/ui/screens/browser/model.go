package browser

import (
	"fmt"

	"github.com/charmbracelet/bubbles/textarea"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/g0p43r/tui_psql/internal/domain"
)

type TableSelectedMsg struct {
	Table domain.DBObject
}

type OpenProfilesMsg struct{}
type DisconnectMsg struct{}
type ReconnectMsg struct{}

type editorMode string

const (
	editorInsert editorMode = "insert"
	editorUpdate editorMode = "update"
	editorDelete editorMode = "delete"
)

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
	rowOffset     int
	colOffset     int
	expanded      bool
	editorActive  bool
	editorMode    editorMode
	editor        textarea.Model
}

func New() Model {
	return Model{
		status:        "No tables loaded.",
		previewStatus: "Select a table to inspect its rows.",
		focus:         FocusTables,
		editor:        newEditor(),
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
	m.rowOffset = 0
	m.colOffset = 0
	m.expanded = false
	m.editorActive = false
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
	m.rowOffset = 0
	m.colOffset = 0
	m.expanded = false
	m.editorActive = false
}

func (m *Model) SetPreview(table domain.DBObject, result domain.QueryResult) {
	m.previewTable = &table
	m.preview = result
	m.previewError = ""
	m.selectedRow = 0
	m.rowOffset = 0
	m.colOffset = 0
	m.expanded = false
	m.editorActive = false
	m.previewStatus = fmt.Sprintf("Showing %d rows from %s.%s", len(result.Rows), table.Schema, table.Name)
}

func (m *Model) SetPreviewError(table domain.DBObject, status string) {
	m.previewTable = &table
	m.preview = domain.QueryResult{}
	m.previewError = status
	m.selectedRow = 0
	m.rowOffset = 0
	m.colOffset = 0
	m.expanded = false
	m.editorActive = false
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

func (m *Model) ensureRowVisible(viewportRows int) {
	if viewportRows <= 0 {
		return
	}
	if m.selectedRow < m.rowOffset {
		m.rowOffset = m.selectedRow
	}
	maxVisible := m.rowOffset + viewportRows - 1
	if m.selectedRow > maxVisible {
		m.rowOffset = m.selectedRow - viewportRows + 1
	}
	if m.rowOffset < 0 {
		m.rowOffset = 0
	}
}

func (m Model) previewViewportRows() int {
	usableHeight := maxInt(12, m.height-2)
	paneContentHeight := maxInt(1, usableHeight-4)
	tableHeight := maxInt(4, paneContentHeight-8)
	return maxInt(1, tableHeight-3)
}

func newEditor() textarea.Model {
	editor := textarea.New()
	editor.Prompt = ""
	editor.CharLimit = 0
	editor.ShowLineNumbers = true
	editor.SetWidth(80)
	editor.SetHeight(20)
	editor.Focus()
	return editor
}

func (m *Model) OpenEditor(mode editorMode) {
	table, ok := m.SelectedTable()
	if !ok {
		return
	}

	m.editor = newEditor()
	m.editorMode = mode
	m.editorActive = true
	m.editor.SetValue(buildSQLTemplate(mode, table, m.preview.Columns, m.currentRow()))
}

func (m *Model) CloseEditor() {
	m.editorActive = false
}

func (m *Model) SetEditorSize(width, height int) {
	editorWidth := maxInt(40, width-16)
	editorHeight := maxInt(10, height-10)
	m.editor.SetWidth(editorWidth)
	m.editor.SetHeight(editorHeight)
}

func (m Model) currentRow() []string {
	row, ok := m.SelectedPreviewRow()
	if !ok {
		return nil
	}
	return row
}
