package app

import (
	tea "github.com/charmbracelet/bubbletea"

	"github.com/g0p43r/tui_psql/internal/config"
	"github.com/g0p43r/tui_psql/internal/domain"
	"github.com/g0p43r/tui_psql/internal/pg"
	"github.com/g0p43r/tui_psql/internal/ui/screens/browser"
	"github.com/g0p43r/tui_psql/internal/ui/screens/connection"
	"github.com/jackc/pgx/v5"
)

type screen string

const (
	screenConnection screen = "connection"
	screenBrowser    screen = "browser"
)

type Model struct {
	connection connection.Model
	browser    browser.Model
	activeConn *pgx.Conn
	current    *domain.ConnectionProfile
	screen     screen
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
		if msg.String() == "ctrl+c" {
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
		m.connection.SetStatus(connection.StatusError, msg.err.Error())
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
		return m.handleTablesLoaded(msg)
	case tablesLoadErrorMsg:
		m.browser.SetStatus(msg.err.Error())
		return m, nil, true
	case browser.TableSelectedMsg:
		if m.activeConn == nil {
			return m, nil, true
		}
		m.browser.SetPreviewStatus("Loading rows...")
		return m, previewTableCmd(m.activeConn, msg.Table), true
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
	case previewLoadedMsg:
		m.browser.SetPreview(msg.table, msg.result)
		return m, nil, true
	case previewLoadErrorMsg:
		m.browser.SetPreviewError(msg.table, msg.err.Error())
		return m, nil, true
	}
	return m, nil, false
}

func (m Model) handleConnectSuccess(msg connectSuccessMsg) (Model, tea.Cmd, bool) {
	m.closeActiveConn()
	m.activeConn = msg.conn
	profile := msg.profile
	m.current = &profile
	m.connection.SetStatus(connection.StatusSuccess, connection.SuccessMessage(msg.profile))
	m.browser.SetStatus("Loading tables...")
	m.screen = screenBrowser
	return m, tea.Batch(
		listTablesCmd(msg.conn),
		saveProfileCmd(msg.profile),
	), true
}

func (m Model) handleProfilesLoaded(msg profilesLoadedMsg) (Model, tea.Cmd, bool) {
	if msg.err != nil {
		m.connection.SetStatus(connection.StatusError, msg.err.Error())
		return m, nil, true
	}
	m.connection.SetProfiles(msg.profiles)
	return m, nil, true
}

func (m Model) handleProfileDeleted(msg profileDeletedMsg) (Model, tea.Cmd, bool) {
	if msg.err != nil {
		m.connection.SetStatus(connection.StatusError, msg.err.Error())
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
	if selected, ok := m.browser.SelectedTable(); ok && m.activeConn != nil {
		m.browser.SetPreviewStatus("Loading rows...")
		return m, previewTableCmd(m.activeConn, selected), true
	}
	return m, nil, true
}

func (m *Model) closeActiveConn() {
	if m.activeConn != nil {
		_ = m.activeConn.Close(m.connection.Context())
		m.activeConn = nil
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

func connectCmd(profile domain.ConnectionProfile) tea.Cmd {
	return func() tea.Msg {
		conn, err := pg.Connect(profile)
		if err != nil {
			return connectErrorMsg{err: err}
		}

		return connectSuccessMsg{
			profile: profile,
			conn:    conn,
		}
	}
}

func listTablesCmd(conn *pgx.Conn) tea.Cmd {
	return func() tea.Msg {
		tables, err := pg.ListTables(conn)
		if err != nil {
			return tablesLoadErrorMsg{err: err}
		}

		return tablesLoadedMsg{tables: tables}
	}
}

func previewTableCmd(conn *pgx.Conn, table domain.DBObject) tea.Cmd {
	return func() tea.Msg {
		result, err := pg.PreviewTable(conn, table, 50)
		if err != nil {
			return previewLoadErrorMsg{
				table: table,
				err:   err,
			}
		}

		return previewLoadedMsg{
			table:  table,
			result: result,
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
