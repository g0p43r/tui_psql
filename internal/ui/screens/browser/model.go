package browser

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/textarea"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/g0p43r/tui_psql/internal/domain"
	"github.com/g0p43r/tui_psql/internal/ui/styles"
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
	QueryWorkbench bool
	NewTableSchema string
	NewTableName   string
	TargetSchema   string
	TargetTable    string
}
type OpenHistoryMsg struct{}

type editorMode string

const (
	editorQuery  editorMode = "query"
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
	ModeBrowse  ViewMode = "browse"
	ModeRecord  ViewMode = "record"
	ModeEditor  ViewMode = "editor"
	ModeHistory ViewMode = "history"
	ModeQuery   ViewMode = "query"
)

type QueryHistoryEntry struct {
	SQL       string
	QueryType domain.SQLQueryType
	Success   bool
	Status    string
}

type Model struct {
	tables          []domain.DBObject
	selected        int
	tableOffset     int
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
	editorResult    domain.QueryResult
	editorErr       string
	resultRow       int
	resultRowOffset int
	resultColOffset int
	pendingSchema   string
	pendingTable    string
	layoutOffset    int
	history         []QueryHistoryEntry
	historySelected int
	historyOffset   int
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
	mode := m.mode
	m.tables = tables
	m.selected = 0
	m.tableOffset = 0
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
	if mode == ModeQuery || mode == ModeEditor {
		m.mode = mode
	}
	if len(tables) == 0 {
		m.status = "No tables found."
		return
	}

	m.status = fmt.Sprintf("Loaded %d objects.", len(tables))
	m.ensureTableVisible(m.tableViewportRows())
}

func (m *Model) SetPreviewStatus(status string) {
	m.previewStatus = status
	m.previewError = ""
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

func (m *Model) ensureTableVisible(viewportRows int) {
	if viewportRows <= 0 {
		return
	}
	if m.selected < m.tableOffset {
		m.tableOffset = m.selected
	}
	maxVisible := m.tableOffset + viewportRows - 1
	if m.selected > maxVisible {
		m.tableOffset = m.selected - viewportRows + 1
	}
	if m.tableOffset < 0 {
		m.tableOffset = 0
	}
}

func (m Model) previewViewportRows() int {
	paneContentHeight := maxInt(1, m.mainAreaHeight()-styles.Panel.GetVerticalFrameSize())
	tableHeight := maxInt(4, paneContentHeight-8)
	return maxInt(1, tableHeight-3)
}

func (m Model) tableViewportRows() int {
	paneContentHeight := maxInt(1, m.mainAreaHeight()-styles.Panel.GetVerticalFrameSize())
	// title + status + spacer = 3 lines
	return maxInt(1, paneContentHeight-3)
}

func (m Model) mainAreaHeight() int {
	usableHeight := maxInt(12, m.height-2)
	hotkeysHeight := m.hotkeysAreaHeight()
	return maxInt(6, usableHeight-hotkeysHeight)
}

func (m Model) hotkeysAreaHeight() int {
	usableHeight := maxInt(12, m.height-2)
	hotkeysHeight := int(float64(usableHeight) * 0.18)
	// Need at least 3 content rows + panel frame (border+padding).
	hotkeysHeight = maxInt(7, hotkeysHeight)
	hotkeysHeight = minInt(hotkeysHeight, maxInt(7, usableHeight-6))
	return hotkeysHeight
}

func newEditor() textarea.Model {
	editor := textarea.New()
	editor.Prompt = ""
	editor.CharLimit = 0
	editor.ShowLineNumbers = true
	editor.FocusedStyle.CursorLine = lipgloss.NewStyle()
	editor.BlurredStyle.CursorLine = lipgloss.NewStyle()
	editor.SetWidth(80)
	editor.SetHeight(20)
	editor.Focus()
	return editor
}

func (m *Model) OpenEditor(mode editorMode) {
	table, ok := m.SelectedTable()
	if !ok && mode != editorCreate && mode != editorQuery {
		return
	}

	m.editor = newEditor()
	m.editorMode = mode
	m.mode = ModeEditor
	m.editorType = queryTypeForEditorMode(mode)
	m.editorStatus = ""
	m.editorStatusErr = false
	m.editorResult = domain.QueryResult{}
	m.editorErr = ""
	if mode == editorQuery {
		m.editorType = domain.QueryTypeAuto
		m.editor.SetValue("SELECT 1;")
		return
	}
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
	layout := newSQLEditorLayout(maxInt(80, width-4), maxInt(12, height-2), false)
	m.editor.SetWidth(layout.paneContentWidth)
	m.editor.SetHeight(layout.editorHeight)
}

func (m *Model) SetQueryWorkbenchEditorSize(width, height int) {
	layout := newQueryWorkbenchLayout(maxInt(80, width-4), maxInt(12, height))
	editorWidth := maxInt(1, layout.contentWidth)
	editorHeight := maxInt(1, layout.editorHeight)
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

func (m *Model) OpenHistory() {
	if len(m.history) == 0 {
		return
	}
	m.mode = ModeHistory
	if m.historySelected < 0 || m.historySelected >= len(m.history) {
		m.historySelected = 0
	}
	m.ensureHistoryVisible(m.historyViewportRows())
}

func (m *Model) CloseHistory() {
	m.mode = ModeBrowse
}

func (m *Model) OpenQueryWorkbench() {
	m.mode = ModeQuery
	m.layoutOffset = 0
	if strings.TrimSpace(m.editor.Value()) == "" {
		m.editor = newEditor()
		m.editor.SetValue("SELECT 1;")
	}
	m.editorType = domain.QueryTypeAuto
	m.editorStatus = ""
	m.editorStatusErr = false
}

func (m *Model) CloseQueryWorkbench() {
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

func (m Model) IsHistoryOpen() bool {
	return m.mode == ModeHistory
}

func (m Model) IsQueryWorkbenchOpen() bool {
	return m.mode == ModeQuery
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

func (m *Model) SetEditorResult(result domain.QueryResult) {
	m.editorResult = result
	m.editorErr = ""
	m.resultRow = 0
	m.resultRowOffset = 0
	m.resultColOffset = 0
}

func (m *Model) SetEditorError(err string) {
	m.editorErr = strings.TrimSpace(err)
}

func (m *Model) MoveResultUp() {
	if m.resultRow > 0 {
		m.resultRow--
		m.ensureResultRowVisible(m.resultViewportRows())
	}
}

func (m *Model) MoveResultDown() {
	if m.resultRow < len(m.editorResult.Rows)-1 {
		m.resultRow++
		m.ensureResultRowVisible(m.resultViewportRows())
	}
}

func (m *Model) MoveResultPageUp() {
	if len(m.editorResult.Rows) == 0 {
		return
	}
	step := maxInt(1, m.resultViewportRows()-1)
	m.resultRow -= step
	if m.resultRow < 0 {
		m.resultRow = 0
	}
	m.ensureResultRowVisible(m.resultViewportRows())
}

func (m *Model) MoveResultPageDown() {
	if len(m.editorResult.Rows) == 0 {
		return
	}
	step := maxInt(1, m.resultViewportRows()-1)
	m.resultRow += step
	if m.resultRow >= len(m.editorResult.Rows) {
		m.resultRow = len(m.editorResult.Rows) - 1
	}
	m.ensureResultRowVisible(m.resultViewportRows())
}

func (m *Model) MoveResultLeft() {
	if m.resultColOffset > 0 {
		m.resultColOffset--
	}
}

func (m *Model) MoveResultRight() {
	maxOffset := maxInt(0, len(m.editorResult.Columns)-2)
	if m.resultColOffset < maxOffset {
		m.resultColOffset++
	}
}

func (m *Model) ensureResultRowVisible(viewportRows int) {
	if viewportRows <= 0 {
		return
	}
	if m.resultRow < m.resultRowOffset {
		m.resultRowOffset = m.resultRow
	}
	maxVisible := m.resultRowOffset + viewportRows - 1
	if m.resultRow > maxVisible {
		m.resultRowOffset = m.resultRow - viewportRows + 1
	}
	if m.resultRowOffset < 0 {
		m.resultRowOffset = 0
	}
}

func (m Model) resultViewportRows() int {
	layout := newQueryWorkbenchLayout(maxInt(80, m.width-4), maxInt(12, m.height))
	return maxInt(1, layout.resultHeight-3)
}

func (m *Model) SetPendingSelection(schema, table string) {
	m.pendingSchema = strings.TrimSpace(schema)
	m.pendingTable = strings.TrimSpace(table)
}

func (m *Model) ClearPendingSelection() {
	m.pendingSchema = ""
	m.pendingTable = ""
}

func (m *Model) AddHistory(sql string, queryType domain.SQLQueryType, success bool, status string) {
	sql = strings.TrimSpace(sql)
	if sql == "" {
		return
	}
	entry := QueryHistoryEntry{
		SQL:       sql,
		QueryType: queryType,
		Success:   success,
		Status:    status,
	}
	m.history = append([]QueryHistoryEntry{entry}, m.history...)
	if len(m.history) > 200 {
		m.history = m.history[:200]
	}
	m.historySelected = 0
	m.historyOffset = 0
}

func (m Model) historyViewportRows() int {
	usableHeight := maxInt(12, m.height-2)
	return maxInt(4, usableHeight-10)
}

func (m *Model) ensureHistoryVisible(viewportRows int) {
	if viewportRows <= 0 {
		return
	}
	if m.historySelected < m.historyOffset {
		m.historyOffset = m.historySelected
	}
	maxVisible := m.historyOffset + viewportRows - 1
	if m.historySelected > maxVisible {
		m.historyOffset = m.historySelected - viewportRows + 1
	}
	if m.historyOffset < 0 {
		m.historyOffset = 0
	}
}

func (m *Model) MoveHistoryUp() {
	if m.historySelected > 0 {
		m.historySelected--
		m.ensureHistoryVisible(m.historyViewportRows())
	}
}

func (m *Model) MoveHistoryDown() {
	if m.historySelected < len(m.history)-1 {
		m.historySelected++
		m.ensureHistoryVisible(m.historyViewportRows())
	}
}

func (m Model) SelectedHistory() (QueryHistoryEntry, bool) {
	if len(m.history) == 0 || m.historySelected < 0 || m.historySelected >= len(m.history) {
		return QueryHistoryEntry{}, false
	}
	return m.history[m.historySelected], true
}

func (m *Model) OpenEditorWithHistory(entry QueryHistoryEntry) {
	m.editor = newEditor()
	m.editorMode = editorUpdate
	m.mode = ModeEditor
	m.editorType = entry.QueryType
	m.editorStatus = ""
	m.editorStatusErr = false
	m.editor.SetValue(entry.SQL)
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
	case editorQuery:
		return domain.QueryTypeAuto
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
