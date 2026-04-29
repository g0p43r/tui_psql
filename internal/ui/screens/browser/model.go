package browser

import (
	"fmt"
	"regexp"
	"strings"

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
type ExecuteSQLMsg struct {
	SQL            string
	QueryType      domain.SQLQueryType
	EditorMode     editorMode
	NewTableSchema string
	NewTableName   string
}

type editorMode string

const (
	editorInsert editorMode = "insert"
	editorUpdate editorMode = "update"
	editorDelete editorMode = "delete"
	editorCreate editorMode = "create"
	editorAlter  editorMode = "alter"
	editorDrop   editorMode = "drop"
)

type FocusArea string

const (
	FocusTables  FocusArea = "tables"
	FocusPreview FocusArea = "preview"
)

type ViewMode string

const (
	ModeBrowse ViewMode = "browse"
	ModeRecord ViewMode = "record"
	ModeEditor ViewMode = "editor"
)

type Model struct {
	tables          []domain.DBObject
	selected        int
	width           int
	height          int
	status          string
	previewStatus   string
	previewTable    *domain.DBObject
	preview         domain.QueryResult
	previewError    string
	focus           FocusArea
	selectedRow     int
	rowOffset       int
	colOffset       int
	mode            ViewMode
	editorMode      editorMode
	editor          textarea.Model
	editorType      domain.SQLQueryType
	editorStatus    string
	editorStatusErr bool
	pendingSchema   string
	pendingTable    string
}

func New() Model {
	return Model{
		status:        "No tables loaded.",
		previewStatus: "Select a table to inspect its rows.",
		focus:         FocusTables,
		mode:          ModeBrowse,
		editor:        newEditor(),
		editorType:    domain.QueryTypeAuto,
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
	if m.pendingTable != "" {
		for i, t := range tables {
			if t.Name == m.pendingTable && (m.pendingSchema == "" || t.Schema == m.pendingSchema) {
				m.selected = i
				break
			}
		}
		m.pendingSchema = ""
		m.pendingTable = ""
	}
	m.resetPreviewState()
	if len(tables) == 0 {
		m.status = "No tables found."
		return
	}

	m.status = fmt.Sprintf("Loaded %d objects.", len(tables))
}

func (m *Model) SetPreviewStatus(status string) {
	m.previewStatus = status
	m.resetPreviewState()
}

func (m *Model) SetPreview(table domain.DBObject, result domain.QueryResult) {
	m.previewTable = &table
	m.preview = result
	m.resetPreviewNavigation()
	m.mode = ModeBrowse
	m.previewError = ""
	m.previewStatus = fmt.Sprintf("Showing %d rows from %s.%s", len(result.Rows), table.Schema, table.Name)
}

func (m *Model) SetPreviewError(table domain.DBObject, status string) {
	m.previewTable = &table
	m.resetPreviewState()
	m.previewError = status
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
	if !ok && mode != editorCreate {
		return
	}

	m.editor = newEditor()
	m.editorMode = mode
	m.mode = ModeEditor
	m.editorType = queryTypeForEditorMode(mode)
	m.editorStatus = ""
	m.editorStatusErr = false
	if mode == editorCreate {
		schema := "public"
		if ok && table.Schema != "" {
			schema = table.Schema
		}
		m.editor.SetValue(buildCreateTableTemplate(schema))
		return
	}
	if mode == editorAlter {
		if !ok {
			return
		}
		m.editor.SetValue(buildAlterTableTemplate(table))
		return
	}
	if mode == editorDrop {
		if !ok {
			return
		}
		m.editor.SetValue(buildDropTableTemplate(table))
		return
	}
	m.editor.SetValue(buildSQLTemplate(mode, table, m.preview.Columns, m.preview.ColumnTypes, m.currentRow()))
}

func (m *Model) CloseEditor() {
	m.mode = ModeBrowse
	m.editorStatus = ""
	m.editorStatusErr = false
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

func (m *Model) OpenRecordView() {
	m.mode = ModeRecord
}

func (m *Model) CloseRecordView() {
	m.mode = ModeBrowse
}

func (m Model) IsBrowsing() bool {
	return m.mode == ModeBrowse
}

func (m Model) IsEditorOpen() bool {
	return m.mode == ModeEditor
}

func (m Model) IsRecordOpen() bool {
	return m.mode == ModeRecord
}

func (m Model) EditorType() domain.SQLQueryType {
	return m.editorType
}

func (m *Model) CycleEditorType() {
	switch m.editorType {
	case domain.QueryTypeAuto:
		m.editorType = domain.QueryTypeSelect
	case domain.QueryTypeSelect:
		m.editorType = domain.QueryTypeInsert
	case domain.QueryTypeInsert:
		m.editorType = domain.QueryTypeUpdate
	case domain.QueryTypeUpdate:
		m.editorType = domain.QueryTypeDelete
	case domain.QueryTypeDelete:
		m.editorType = domain.QueryTypeExec
	default:
		m.editorType = domain.QueryTypeAuto
	}
}

func (m *Model) SetEditorStatus(status string, isError bool) {
	m.editorStatus = status
	m.editorStatusErr = isError
}

func (m *Model) SetPendingSelection(schema, table string) {
	m.pendingSchema = strings.TrimSpace(schema)
	m.pendingTable = strings.TrimSpace(table)
}

func parseCreateTableTarget(sql string) (schema, table string, ok bool) {
	trimmed := strings.TrimSpace(sql)
	re := regexp.MustCompile(`(?is)^create\s+table\s+(if\s+not\s+exists\s+)?([a-zA-Z_][\w]*)(?:\.([a-zA-Z_][\w]*))?`)
	match := re.FindStringSubmatch(trimmed)
	if len(match) == 0 {
		return "", "", false
	}

	if match[3] != "" {
		return match[2], match[3], true
	}
	return "", match[2], true
}

func (m *Model) resetPreviewState() {
	m.preview = domain.QueryResult{}
	m.previewError = ""
	m.resetPreviewNavigation()
	m.mode = ModeBrowse
}

func (m *Model) resetPreviewNavigation() {
	m.selectedRow = 0
	m.rowOffset = 0
	m.colOffset = 0
}

func queryTypeForEditorMode(mode editorMode) domain.SQLQueryType {
	switch mode {
	case editorInsert:
		return domain.QueryTypeInsert
	case editorUpdate:
		return domain.QueryTypeUpdate
	case editorDelete:
		return domain.QueryTypeDelete
	case editorCreate:
		return domain.QueryTypeExec
	case editorAlter:
		return domain.QueryTypeExec
	case editorDrop:
		return domain.QueryTypeExec
	default:
		return domain.QueryTypeAuto
	}
}
