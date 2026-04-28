package connection

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
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
			m.zone = ZoneForm
			m.SetStatus(StatusIdle, "Connection form focused.")
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
			m.focus--
			if m.focus < 0 {
				m.focus = len(m.fields) - 1
			}
			for i := range m.fields {
				if i == m.focus {
					m.fields[i].input.Focus()
					continue
				}
				m.fields[i].input.Blur()
			}
			return m, nil
		case "down", "j":
			if m.zone == ZoneProfiles {
				if !m.SelectNextProfile() {
					m.SetStatus(StatusError, "No saved profiles found.")
				}
				return m, nil
			}
			m.focus++
			if m.focus >= len(m.fields) {
				m.focus = 0
			}
			for i := range m.fields {
				if i == m.focus {
					m.fields[i].input.Focus()
					continue
				}
				m.fields[i].input.Blur()
			}
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
					m.SetStatus(StatusError, err.Error())
					return m, nil
				}
				return m, func() tea.Msg { return SubmitMsg{} }
			}
			m.focus++
			if m.focus >= len(m.fields) {
				m.focus = 0
			}
			for i := range m.fields {
				if i == m.focus {
					m.fields[i].input.Focus()
					continue
				}
				m.fields[i].input.Blur()
			}
		case "tab", "shift+tab":
			if m.zone == ZoneProfiles {
				return m, nil
			}
			if msg.String() == "shift+tab" {
				m.focus--
			} else {
				m.focus++
			}

			if m.focus >= len(m.fields) {
				m.focus = 0
			}
			if m.focus < 0 {
				m.focus = len(m.fields) - 1
			}

			for i := range m.fields {
				if i == m.focus {
					m.fields[i].input.Focus()
					continue
				}
				m.fields[i].input.Blur()
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
