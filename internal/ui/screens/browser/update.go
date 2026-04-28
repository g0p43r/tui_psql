package browser

import tea "github.com/charmbracelet/bubbletea"

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if m.editorActive {
			switch msg.String() {
			case "esc":
				m.CloseEditor()
				return m, nil
			}

			var cmd tea.Cmd
			m.editor, cmd = m.editor.Update(msg)
			return m, cmd
		}

		if m.expanded {
			switch msg.String() {
			case "esc", "enter":
				m.expanded = false
			}
			return m, nil
		}

		switch msg.String() {
		case "ctrl+p":
			return m, func() tea.Msg { return OpenProfilesMsg{} }
		case "ctrl+x":
			return m, func() tea.Msg { return DisconnectMsg{} }
		case "ctrl+r":
			return m, func() tea.Msg { return ReconnectMsg{} }
		case "f2":
			m.OpenEditor(editorInsert)
			m.SetEditorSize(m.width, m.height)
			return m, nil
		case "f3":
			m.OpenEditor(editorUpdate)
			m.SetEditorSize(m.width, m.height)
			return m, nil
		case "f4":
			m.OpenEditor(editorDelete)
			m.SetEditorSize(m.width, m.height)
			return m, nil
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
				m.ensureRowVisible(m.previewViewportRows())
			}
		case "down", "j":
			if m.focus == FocusTables {
				if m.selected < len(m.tables)-1 {
					m.selected++
					return m, m.selectCmd()
				}
			} else if m.selectedRow < len(m.preview.Rows)-1 {
				m.selectedRow++
				m.ensureRowVisible(m.previewViewportRows())
			}
		case "left", "h":
			if m.focus == FocusPreview && m.colOffset > 0 {
				m.colOffset--
			}
		case "right", "l":
			maxOffset := maxInt(0, len(m.preview.Columns)-2)
			if m.focus == FocusPreview && m.colOffset < maxOffset {
				m.colOffset++
			}
		case "pgdown":
			if m.focus == FocusPreview && len(m.preview.Rows) > 0 {
				step := maxInt(1, m.previewViewportRows()-1)
				m.selectedRow += step
				if m.selectedRow >= len(m.preview.Rows) {
					m.selectedRow = len(m.preview.Rows) - 1
				}
				m.ensureRowVisible(m.previewViewportRows())
			}
		case "pgup":
			if m.focus == FocusPreview && len(m.preview.Rows) > 0 {
				step := maxInt(1, m.previewViewportRows()-1)
				m.selectedRow -= step
				if m.selectedRow < 0 {
					m.selectedRow = 0
				}
				m.ensureRowVisible(m.previewViewportRows())
			}
		case "home":
			if m.focus == FocusPreview && len(m.preview.Rows) > 0 {
				m.selectedRow = 0
				m.ensureRowVisible(m.previewViewportRows())
			}
		case "end":
			if m.focus == FocusPreview && len(m.preview.Rows) > 0 {
				m.selectedRow = len(m.preview.Rows) - 1
				m.ensureRowVisible(m.previewViewportRows())
			}
		}
	}

	return m, nil
}
