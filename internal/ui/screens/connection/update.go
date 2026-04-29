package connection

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/g0p43r/tui_psql/internal/errs"
)

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+p":
			if m.CanFocusProfiles() {
				m.zone = ZoneProfiles
				m.SetStatus(StatusIdle, "Profiles list focused.")
			} else {
				m.SetStatus(StatusError, "No saved profiles found.")
			}
			return m, nil
		case "ctrl+f":
			m.FocusForm()
			return m, nil
		case "ctrl+r":
			if !m.SelectNextProfile() {
				m.SetStatus(StatusError, "No saved profiles found.")
			}
			return m, nil
		case "ctrl+d":
			if m.zone == ZoneProfiles && m.SelectedProfileName() != "" {
				return m, func() tea.Msg { return DeleteProfileMsg{Name: m.SelectedProfileName()} }
			}
			return m, nil
		case "up", "k":
			if m.zone == ZoneProfiles {
				if !m.SelectPrevProfile() {
					m.SetStatus(StatusError, "No saved profiles found.")
				}
				return m, nil
			}
			m.moveFieldFocus(-1)
			return m, nil
		case "down", "j":
			if m.zone == ZoneProfiles {
				if !m.SelectNextProfile() {
					m.SetStatus(StatusError, "No saved profiles found.")
				}
				return m, nil
			}
			m.moveFieldFocus(1)
			return m, nil
		case "enter":
			if m.zone == ZoneProfiles {
				if m.SelectedProfileName() == "" {
					m.SetStatus(StatusError, "No saved profiles found.")
					return m, nil
				}
				m.ApplyProfile(m.profiles[m.profileIdx])
				m.zone = ZoneForm
				m.SetStatus(StatusIdle, fmt.Sprintf("Applied profile: %s", m.profiles[m.profileIdx].Name))
				return m, nil
			}
			if m.focus == len(m.fields)-1 {
				if err := m.Validate(); err != nil {
					m.SetStatus(StatusError, errs.Message(err))
					return m, nil
				}
				return m, func() tea.Msg { return SubmitMsg{} }
			}
			m.moveFieldFocus(1)
		case "tab", "shift+tab":
			if m.zone == ZoneProfiles {
				return m, nil
			}
			if msg.String() == "shift+tab" {
				m.moveFieldFocus(-1)
			} else {
				m.moveFieldFocus(1)
			}
		}
	}

	cmds := make([]tea.Cmd, 0, len(m.fields))
	for i := range m.fields {
		var cmd tea.Cmd
		m.fields[i].input, cmd = m.fields[i].input.Update(msg)
		cmds = append(cmds, cmd)
	}

	return m, tea.Batch(cmds...)
}
