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
	profiles   []domain.ConnectionProfile
	profileIdx int
	zone       FocusZone
}

type FocusZone string

const (
	ZoneForm     FocusZone = "form"
	ZoneProfiles FocusZone = "profiles"
)

type StatusKind string

const (
	StatusIdle       StatusKind = "idle"
	StatusConnecting StatusKind = "connecting"
	StatusSuccess    StatusKind = "success"
	StatusError      StatusKind = "error"
)

type SubmitMsg struct{}
type DeleteProfileMsg struct {
	Name string
}

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
		zone:       ZoneForm,
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
		Name:     profileName(values["user"], values["host"], values["port"], values["database"]),
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

func (m *Model) SetProfiles(profiles []domain.ConnectionProfile) {
	m.profiles = profiles
	if len(profiles) == 0 {
		m.profileIdx = 0
		m.zone = ZoneForm
		return
	}
	if m.profileIdx >= len(profiles) {
		m.profileIdx = len(profiles) - 1
	}
	m.ApplyProfile(profiles[m.profileIdx])
	if m.statusKind != StatusSuccess {
		m.status = fmt.Sprintf("Loaded %d saved profile(s).", len(profiles))
		m.statusKind = StatusIdle
	}
}

func (m *Model) ApplyProfile(profile domain.ConnectionProfile) {
	values := map[string]string{
		"host":     profile.Host,
		"port":     profile.Port,
		"database": profile.Database,
		"user":     profile.User,
		"password": profile.Password,
	}

	for i := range m.fields {
		key := strings.ToLower(m.fields[i].label)
		m.fields[i].input.SetValue(values[key])
	}
}

func (m *Model) SelectNextProfile() bool {
	return m.cycleProfile(1)
}

func (m *Model) SelectPrevProfile() bool {
	return m.cycleProfile(-1)
}

func (m *Model) cycleProfile(delta int) bool {
	if len(m.profiles) == 0 {
		return false
	}

	m.profileIdx += delta
	if m.profileIdx >= len(m.profiles) {
		m.profileIdx = 0
	}
	if m.profileIdx < 0 {
		m.profileIdx = len(m.profiles) - 1
	}

	m.ApplyProfile(m.profiles[m.profileIdx])
	m.statusKind = StatusIdle
	m.status = fmt.Sprintf("Profile: %s", m.profiles[m.profileIdx].Name)
	return true
}

func (m Model) SelectedProfileName() string {
	if len(m.profiles) == 0 || m.profileIdx < 0 || m.profileIdx >= len(m.profiles) {
		return ""
	}
	return m.profiles[m.profileIdx].Name
}

func (m Model) CanFocusProfiles() bool {
	return len(m.profiles) > 0
}

func (m *Model) FocusProfiles() bool {
	if !m.CanFocusProfiles() {
		return false
	}
	m.zone = ZoneProfiles
	m.SetStatus(StatusIdle, fmt.Sprintf("Profiles list focused. Current: %s", m.profiles[m.profileIdx].Name))
	return true
}

func (m *Model) FocusForm() {
	m.zone = ZoneForm
	m.SetStatus(StatusIdle, "Connection form focused.")
}

func (m *Model) moveFieldFocus(delta int) {
	m.focus += delta
	if m.focus >= len(m.fields) {
		m.focus = 0
	}
	if m.focus < 0 {
		m.focus = len(m.fields) - 1
	}
	m.syncFieldFocus()
}

func (m *Model) syncFieldFocus() {
	for i := range m.fields {
		if i == m.focus {
			m.fields[i].input.Focus()
			continue
		}
		m.fields[i].input.Blur()
	}
}

func profileName(user, host, port, database string) string {
	return fmt.Sprintf("%s@%s:%s/%s", user, host, port, database)
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
