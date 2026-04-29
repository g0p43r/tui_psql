package app

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/g0p43r/tui_psql/internal/config"
	"github.com/g0p43r/tui_psql/internal/domain"
	"github.com/g0p43r/tui_psql/internal/errs"
	"github.com/g0p43r/tui_psql/internal/pg"
	"github.com/g0p43r/tui_psql/internal/ui/screens/browser"
	"github.com/g0p43r/tui_psql/internal/ui/screens/connection"
	"github.com/jackc/pgx/v5/pgxpool"
)

type screen string

const (
	screenConnection screen = "connection"
	screenBrowser    screen = "browser"
)

type Model struct {
	connection      connection.Model
	browser         browser.Model
	session         *dbSession
	current         *domain.ConnectionProfile
	screen          screen
	nextSessionID   int
	nextRequestID   int
	latestTablesID  int
	latestPreviewID int
}

type dbSession struct {
	id   int
	pool *pgxpool.Pool
}

func (s *dbSession) Close() {
	if s == nil || s.pool == nil {
		return
	}
	s.pool.Close()
}

func NewModel() Model {
	return Model{
		connection: connection.New(),
		browser:    browser.New(),
		screen:     screenConnection,
	}
}

func (m Model) Init() tea.Cmd {
	return tea.Batch(
		m.connection.Init(),
		loadProfilesCmd(),
	)
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if next, cmd, handled := m.handleGlobalMessage(msg); handled {
		return next, cmd
	}
	if next, cmd, handled := m.handleAppMessage(msg); handled {
		return next, cmd
	}

	switch m.screen {
	case screenBrowser:
		var cmd tea.Cmd
		m.browser, cmd = m.browser.Update(msg)
		return m, cmd
	default:
		var cmd tea.Cmd
		m.connection, cmd = m.connection.Update(msg)
		return m, cmd
	}
}

func (m Model) handleGlobalMessage(msg tea.Msg) (Model, tea.Cmd, bool) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.connection.SetSize(msg.Width, msg.Height)
		m.browser.SetSize(msg.Width, msg.Height)
		return m, nil, true
	case tea.KeyMsg:
		if msg.String() == "ctrl+q" {
			return m, tea.Quit, true
		}
	}
	return m, nil, false
}

func (m Model) handleAppMessage(msg tea.Msg) (Model, tea.Cmd, bool) {
	switch msg := msg.(type) {
	case connection.SubmitMsg:
		profile := m.connection.Profile()
		m.connection.SetStatus(connection.StatusConnecting, "Connecting to database...")
		return m, connectCmd(profile), true
	case connectSuccessMsg:
		return m.handleConnectSuccess(msg)
	case connectErrorMsg:
		m.connection.SetStatus(connection.StatusError, errs.Message(msg.err))
		return m, nil, true
	case profilesLoadedMsg:
		return m.handleProfilesLoaded(msg)
	case profileSavedMsg:
		if msg.err == nil {
			m.connection.SetProfiles(msg.profiles)
		}
		return m, nil, true
	case connection.DeleteProfileMsg:
		return m, deleteProfileCmd(msg.Name), true
	case profileDeletedMsg:
		return m.handleProfileDeleted(msg)
	case tablesLoadedMsg:
		if msg.requestID != m.latestTablesID {
			return m, nil, true
		}
		return m.handleTablesLoaded(msg)
	case tablesLoadErrorMsg:
		if msg.requestID != m.latestTablesID {
			return m, nil, true
		}
		m.browser.SetStatus(errs.Message(msg.err))
		return m, nil, true
	case browser.TableSelectedMsg:
		if m.session == nil {
			return m, nil, true
		}
		m.browser.SetPreviewStatus("Loading rows...")
		requestID := m.nextID()
		m.latestPreviewID = requestID
		return m, previewTableCmd(m.session, requestID, msg.Table), true
	case browser.OpenProfilesMsg:
		m.screen = screenConnection
		if !m.connection.FocusProfiles() {
			m.connection.SetStatus(connection.StatusError, "No saved profiles found.")
		}
		return m, nil, true
	case browser.DisconnectMsg:
		m.closeActiveConn()
		m.screen = screenConnection
		m.connection.SetStatus(connection.StatusIdle, "Disconnected. Choose a profile or edit the form to connect again.")
		return m, nil, true
	case browser.ReconnectMsg:
		m.closeActiveConn()
		profile := m.connection.Profile()
		if m.current != nil {
			profile = *m.current
		}
		m.screen = screenConnection
		m.connection.SetStatus(connection.StatusConnecting, "Reconnecting to database...")
		return m, connectCmd(profile), true
	case browser.ExecuteSQLMsg:
		if m.session == nil {
			m.browser.SetEditorStatus("No active connection.", true)
			return m, nil, true
		}
		if msg.EditorMode == "create" && msg.NewTableName != "" {
			m.browser.SetPendingSelection(msg.NewTableSchema, msg.NewTableName)
		}
		m.browser.SetEditorStatus("Executing query...", false)
		return m, executeSQLCmd(m.session, msg.SQL, msg.QueryType, string(msg.EditorMode), msg.QueryWorkbench, msg.NewTableSchema, msg.NewTableName, msg.TargetSchema, msg.TargetTable), true
	case previewLoadedMsg:
		if msg.requestID != m.latestPreviewID {
			return m, nil, true
		}
		m.browser.SetPreview(msg.table, msg.result)
		return m, nil, true
	case previewLoadErrorMsg:
		if msg.requestID != m.latestPreviewID {
			return m, nil, true
		}
		m.browser.SetPreviewError(msg.table, errs.Message(msg.err))
		return m, nil, true
	case sqlExecutedMsg:
		if m.session == nil || msg.sessionID != m.session.id {
			return m, nil, true
		}
		return m.handleSQLExecuted(msg)
	case sqlExecuteErrorMsg:
		if m.session == nil || msg.sessionID != m.session.id {
			return m, nil, true
		}
		if msg.editorMode == "create" {
			m.browser.ClearPendingSelection()
		}
		m.browser.AddHistory(msg.sql, msg.queryType, false, errs.Message(msg.err))
		m.browser.SetEditorError(errs.Message(msg.err))
		m.browser.SetEditorStatus(errs.Message(msg.err), true)
		return m, nil, true
	}
	return m, nil, false
}

func (m Model) handleConnectSuccess(msg connectSuccessMsg) (Model, tea.Cmd, bool) {
	m.closeActiveConn()
	m.nextSessionID++
	m.session = &dbSession{id: m.nextSessionID, pool: msg.pool}
	profile := msg.profile
	m.current = &profile
	m.connection.SetStatus(connection.StatusSuccess, connection.SuccessMessage(msg.profile))
	m.browser.SetStatus("Loading tables...")
	m.screen = screenBrowser
	requestID := m.nextID()
	m.latestTablesID = requestID
	return m, tea.Batch(
		listTablesCmd(m.session, requestID),
		saveProfileCmd(msg.profile),
	), true
}

func (m Model) handleProfilesLoaded(msg profilesLoadedMsg) (Model, tea.Cmd, bool) {
	if msg.err != nil {
		m.connection.SetStatus(connection.StatusError, errs.Message(msg.err))
		return m, nil, true
	}
	m.connection.SetProfiles(msg.profiles)
	return m, nil, true
}

func (m Model) handleProfileDeleted(msg profileDeletedMsg) (Model, tea.Cmd, bool) {
	if msg.err != nil {
		m.connection.SetStatus(connection.StatusError, errs.Message(msg.err))
		return m, nil, true
	}
	m.connection.SetProfiles(msg.profiles)
	if len(msg.profiles) == 0 {
		m.connection.SetStatus(connection.StatusIdle, "All saved profiles removed.")
	} else {
		m.connection.SetStatus(connection.StatusIdle, "Profile removed.")
	}
	return m, nil, true
}

func (m Model) handleTablesLoaded(msg tablesLoadedMsg) (Model, tea.Cmd, bool) {
	m.browser.SetTables(msg.tables)
	if selected, ok := m.browser.SelectedTable(); ok && m.session != nil {
		m.browser.SetPreviewStatus("Loading rows...")
		requestID := m.nextID()
		m.latestPreviewID = requestID
		return m, previewTableCmd(m.session, requestID, selected), true
	}
	return m, nil, true
}

func (m Model) handleSQLExecuted(msg sqlExecutedMsg) (Model, tea.Cmd, bool) {
	m.browser.AddHistory(msg.sql, msg.queryType, true, msg.result.CommandTag)
	if msg.queryWorkbench {
		return m.handleWorkbenchSQLExecuted(msg)
	}
	return m.handleEditorSQLExecuted(msg)
}

func (m Model) handleWorkbenchSQLExecuted(msg sqlExecutedMsg) (Model, tea.Cmd, bool) {
	if len(msg.result.Columns) > 0 {
		m.browser.SetEditorResult(msg.result)
		m.browser.SetEditorStatus("Query OK: "+msg.result.CommandTag, false)
		return m, nil, true
	}

	m.browser.SetEditorStatus("Statement OK: "+msg.result.CommandTag, false)
	m.browser.SetEditorResult(msg.result)
	if msg.editorMode == "create" || msg.editorMode == "alter" || msg.editorMode == "drop" {
		return m.refreshSchemaObjects()
	}
	return m, nil, true
}

func (m Model) handleEditorSQLExecuted(msg sqlExecutedMsg) (Model, tea.Cmd, bool) {
	if len(msg.result.Columns) > 0 {
		m.browser.SetEditorResult(msg.result)
		table := domain.DBObject{Schema: "query", Name: "result", Type: domain.ObjectView}
		m.browser.SetPreview(table, msg.result)
		m.browser.SetEditorStatus("Query OK: "+msg.result.CommandTag, false)
		return m, nil, true
	}

	m.browser.SetEditorStatus("Statement OK: "+msg.result.CommandTag, false)
	m.browser.SetEditorResult(msg.result)
	if msg.editorMode == "create" || msg.editorMode == "alter" || msg.editorMode == "drop" {
		return m.refreshSchemaObjects()
	}
	return m.reloadPreviewAfterMutation(msg)
}

func (m Model) refreshSchemaObjects() (Model, tea.Cmd, bool) {
	m.browser.SetStatus("Refreshing schema objects...")
	if m.session == nil {
		return m, nil, true
	}
	requestID := m.nextID()
	m.latestTablesID = requestID
	return m, listTablesCmd(m.session, requestID), true
}

func (m Model) reloadPreviewAfterMutation(msg sqlExecutedMsg) (Model, tea.Cmd, bool) {
	if m.session != nil {
		if msg.targetTable != "" {
			target := domain.DBObject{
				Schema: msg.targetSchema,
				Name:   msg.targetTable,
				Type:   domain.ObjectTable,
			}
			if target.Schema == "" {
				if selected, ok := m.browser.SelectedTable(); ok {
					target.Schema = selected.Schema
				}
			}
			m.browser.SetPreviewStatus("Reloading table preview...")
			requestID := m.nextID()
			m.latestPreviewID = requestID
			return m, previewTableCmd(m.session, requestID, target), true
		}
		if selected, ok := m.browser.SelectedTable(); ok {
			m.browser.SetPreviewStatus("Reloading table preview...")
			requestID := m.nextID()
			m.latestPreviewID = requestID
			return m, previewTableCmd(m.session, requestID, selected), true
		}
	}
	return m, nil, true
}

func (m *Model) closeActiveConn() {
	if m.session != nil {
		m.session.Close()
		m.session = nil
	}
}

func (m Model) View() string {
	switch m.screen {
	case screenBrowser:
		return m.browser.View()
	default:
		return m.connection.View()
	}
}

func (m *Model) nextID() int {
	m.nextRequestID++
	return m.nextRequestID
}

func connectCmd(profile domain.ConnectionProfile) tea.Cmd {
	return func() tea.Msg {
		conn, err := pg.Connect(profile)
		if err != nil {
			return connectErrorMsg{err: err}
		}

		return connectSuccessMsg{
			profile: profile,
			pool:    conn,
		}
	}
}

func listTablesCmd(session *dbSession, requestID int) tea.Cmd {
	return func() tea.Msg {
		tables, err := pg.ListTables(session.pool)
		if err != nil {
			return tablesLoadErrorMsg{requestID: requestID, err: err}
		}

		return tablesLoadedMsg{requestID: requestID, tables: tables}
	}
}

func previewTableCmd(session *dbSession, requestID int, table domain.DBObject) tea.Cmd {
	return func() tea.Msg {
		result, err := pg.PreviewTable(session.pool, table, 50)
		if err != nil {
			return previewLoadErrorMsg{
				requestID: requestID,
				table:     table,
				err:       err,
			}
		}

		return previewLoadedMsg{
			requestID: requestID,
			table:     table,
			result:    result,
		}
	}
}

func executeSQLCmd(session *dbSession, sql string, queryType domain.SQLQueryType, editorMode string, queryWorkbench bool, newTableSchema, newTableName, targetSchema, targetTable string) tea.Cmd {
	return func() tea.Msg {
		resolvedSchema := newTableSchema
		if editorMode == "create" && strings.TrimSpace(newTableName) != "" && strings.TrimSpace(resolvedSchema) == "" {
			schema, err := pg.CurrentSchema(session.pool)
			if err == nil && schema != "" {
				resolvedSchema = schema
			}
		}
		resolvedTargetSchema := targetSchema
		if strings.TrimSpace(targetTable) != "" && strings.TrimSpace(resolvedTargetSchema) == "" {
			schema, err := pg.CurrentSchema(session.pool)
			if err == nil && schema != "" {
				resolvedTargetSchema = schema
			}
		}

		result, err := pg.ExecuteSQL(session.pool, sql, queryType)
		if err != nil {
			return sqlExecuteErrorMsg{
				sessionID:      session.id,
				queryType:      queryType,
				editorMode:     editorMode,
				queryWorkbench: queryWorkbench,
				targetSchema:   resolvedTargetSchema,
				targetTable:    targetTable,
				sql:            sql,
				err:            err,
			}
		}

		return sqlExecutedMsg{
			sessionID:      session.id,
			queryType:      queryType,
			editorMode:     editorMode,
			queryWorkbench: queryWorkbench,
			newTableSchema: resolvedSchema,
			newTableName:   newTableName,
			targetSchema:   resolvedTargetSchema,
			targetTable:    targetTable,
			sql:            sql,
			result:         result,
		}
	}
}

func loadProfilesCmd() tea.Cmd {
	return func() tea.Msg {
		profiles, err := config.LoadProfiles()
		return profilesLoadedMsg{
			profiles: profiles,
			err:      err,
		}
	}
}

func saveProfileCmd(profile domain.ConnectionProfile) tea.Cmd {
	return func() tea.Msg {
		profiles, err := config.SaveProfile(profile)
		return profileSavedMsg{
			profiles: profiles,
			err:      err,
		}
	}
}

func deleteProfileCmd(name string) tea.Cmd {
	return func() tea.Msg {
		profiles, err := config.DeleteProfile(name)
		return profileDeletedMsg{
			profiles: profiles,
			err:      err,
		}
	}
}
