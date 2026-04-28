package connection

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/g0p43r/tui_psql/internal/domain"
	"github.com/g0p43r/tui_psql/internal/ui/styles"
)

type field struct {
	label string
	input textinput.Model
}

type Model struct {
	fields     []field
	focus      int
	width      int
	height     int
	status     string
	statusKind StatusKind
}

type StatusKind string

const (
	StatusIdle       StatusKind = "idle"
	StatusConnecting StatusKind = "connecting"
	StatusSuccess    StatusKind = "success"
	StatusError      StatusKind = "error"
)

type SubmitMsg struct{}

func New() Model {
	fields := []field{
		newField("Host", "localhost"),
		newField("Port", "5432"),
		newField("Database", "postgres"),
		newField("User", "postgres"),
		newField("Password", ""),
	}

	fields[0].input.Focus()
	fields[4].input.EchoMode = textinput.EchoPassword
	fields[4].input.EchoCharacter = '•'

	return Model{
		fields:     fields,
		status:     "Fill in the connection details and press Enter on Password.",
		statusKind: StatusIdle,
	}
}

func (m Model) Init() tea.Cmd {
	return textinput.Blink
}

func (m *Model) SetSize(width, height int) {
	m.width = width
	m.height = height
}

func newField(label, value string) field {
	input := textinput.New()
	input.SetValue(value)
	input.Prompt = ""
	input.CharLimit = 256
	input.Width = 32
	input.Cursor.Style = input.Cursor.Style.Foreground(lipgloss.Color("212"))

	return field{
		label: label,
		input: input,
	}
}

func (m Model) Profile() domain.ConnectionProfile {
	values := m.Values()
	return domain.ConnectionProfile{
		Host:     values["host"],
		Port:     values["port"],
		Database: values["database"],
		User:     values["user"],
		Password: values["password"],
		SSLMode:  "disable",
	}
}

func (m Model) Values() map[string]string {
	values := make(map[string]string, len(m.fields))
	for _, f := range m.fields {
		values[strings.ToLower(f.label)] = strings.TrimSpace(f.input.Value())
	}
	return values
}

func (m Model) Validate() error {
	profile := m.Profile()

	if profile.Host == "" {
		return fmt.Errorf("host is required")
	}
	if profile.Port == "" {
		return fmt.Errorf("port is required")
	}
	port, err := strconv.Atoi(profile.Port)
	if err != nil || port < 1 || port > 65535 {
		return fmt.Errorf("port must be a number between 1 and 65535")
	}
	if profile.Database == "" {
		return fmt.Errorf("database is required")
	}
	if profile.User == "" {
		return fmt.Errorf("user is required")
	}

	return nil
}

func (m *Model) SetStatus(kind StatusKind, status string) {
	m.statusKind = kind
	m.status = status
}

func (m Model) Context() context.Context {
	return context.Background()
}

func SuccessMessage(profile domain.ConnectionProfile) string {
	return fmt.Sprintf(
		"Connected to %s@%s:%s/%s",
		profile.User,
		profile.Host,
		profile.Port,
		profile.Database,
	)
}

func (m Model) renderStatus() string {
	switch m.statusKind {
	case StatusSuccess:
		return lipgloss.NewStyle().Foreground(lipgloss.Color("42")).Render(m.status)
	case StatusError:
		return lipgloss.NewStyle().Foreground(lipgloss.Color("203")).Render(m.status)
	case StatusConnecting:
		return lipgloss.NewStyle().Foreground(lipgloss.Color("220")).Render(m.status)
	default:
		return styles.Subtitle.Render(m.status)
	}
}

func (m Model) String() string {
	return fmt.Sprintf("%v", m.Values())
}
