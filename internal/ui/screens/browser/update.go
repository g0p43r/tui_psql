package browser

import tea "github.com/charmbracelet/bubbletea"

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if m.expanded {
			switch msg.String() {
			case "esc", "enter":
				m.expanded = false
			}
			return m, nil
		}

		switch msg.String() {
		case "tab":
			if m.focus == FocusTables {
				m.focus = FocusPreview
			} else {
				m.focus = FocusTables
			}
		case "enter":
			if m.focus == FocusPreview && len(m.preview.Rows) > 0 {
				m.expanded = true
			}
		case "up", "k":
			if m.focus == FocusTables {
				if m.selected > 0 {
					m.selected--
					return m, m.selectCmd()
				}
			} else if m.selectedRow > 0 {
				m.selectedRow--
			}
		case "down", "j":
			if m.focus == FocusTables {
				if m.selected < len(m.tables)-1 {
					m.selected++
					return m, m.selectCmd()
				}
			} else if m.selectedRow < len(m.preview.Rows)-1 {
				m.selectedRow++
			}
		}
	}

	return m, nil
}
