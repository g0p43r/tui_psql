package connection

import tea "github.com/charmbracelet/bubbletea"

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "tab", "shift+tab", "enter", "up", "down":
			if msg.String() == "enter" && m.focus == len(m.fields)-1 {
				if err := m.Validate(); err != nil {
					m.SetStatus(StatusError, err.Error())
					return m, nil
				}
				return m, func() tea.Msg { return SubmitMsg{} }
			}

			if msg.String() == "up" || msg.String() == "shift+tab" {
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
