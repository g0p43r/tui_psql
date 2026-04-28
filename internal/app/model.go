package app

import (
	tea "github.com/charmbracelet/bubbletea"

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
	return m.connection.Init()
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.connection.SetSize(msg.Width, msg.Height)
		m.browser.SetSize(msg.Width, msg.Height)
	case tea.KeyMsg:
		if msg.String() == "ctrl+c" {
			return m, tea.Quit
		}
	case connection.SubmitMsg:
		profile := m.connection.Profile()
		m.connection.SetStatus(connection.StatusConnecting, "Connecting to database...")
		return m, connectCmd(profile)
	case connectSuccessMsg:
		if m.activeConn != nil {
			_ = m.activeConn.Close(m.connection.Context())
		}
		m.activeConn = msg.conn
		m.connection.SetStatus(connection.StatusSuccess, connection.SuccessMessage(msg.profile))
		m.browser.SetStatus("Loading tables...")
		m.screen = screenBrowser
		return m, listTablesCmd(msg.conn)
	case connectErrorMsg:
		m.connection.SetStatus(connection.StatusError, msg.err.Error())
		return m, nil
	case tablesLoadedMsg:
		m.browser.SetTables(msg.tables)
		if selected, ok := m.browser.SelectedTable(); ok && m.activeConn != nil {
			m.browser.SetPreviewStatus("Loading rows...")
			return m, previewTableCmd(m.activeConn, selected)
		}
		return m, nil
	case tablesLoadErrorMsg:
		m.browser.SetStatus(msg.err.Error())
		return m, nil
	case browser.TableSelectedMsg:
		if m.activeConn == nil {
			return m, nil
		}
		m.browser.SetPreviewStatus("Loading rows...")
		return m, previewTableCmd(m.activeConn, msg.Table)
	case previewLoadedMsg:
		m.browser.SetPreview(msg.table, msg.result)
		return m, nil
	case previewLoadErrorMsg:
		m.browser.SetPreviewError(msg.table, msg.err.Error())
		return m, nil
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
