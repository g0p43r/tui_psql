package browser

import tea "github.com/charmbracelet/bubbletea"

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	keyMsg, ok := msg.(tea.KeyMsg)
	if !ok {
		return m, nil
	}

	if m.IsEditorOpen() {
		return m.handleEditorModeKey(keyMsg)
	}

	if m.IsRecordOpen() {
		return m.handleRecordModeKey(keyMsg)
	}

	return m.handleBrowseModeKey(keyMsg)
}

func (m Model) handleEditorModeKey(msg tea.KeyMsg) (Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		m.CloseEditor()
		return m, nil
	case "ctrl+t":
		m.CycleEditorType()
		return m, nil
	case "f5", "ctrl+enter", "ctrl+j":
		sql := m.editor.Value()
		m.SetEditorStatus("Executing query...", false)
		return m, func() tea.Msg {
			return ExecuteSQLMsg{
				SQL:       sql,
				QueryType: m.EditorType(),
			}
		}
	}

	var cmd tea.Cmd
	m.editor, cmd = m.editor.Update(msg)
	return m, cmd
}

func (m Model) handleRecordModeKey(msg tea.KeyMsg) (Model, tea.Cmd) {
	switch msg.String() {
	case "esc", "enter":
		m.CloseRecordView()
	}
	return m, nil
}

func (m Model) handleBrowseModeKey(msg tea.KeyMsg) (Model, tea.Cmd) {
	switch msg.String() {
	case "ctrl+p":
		return m, func() tea.Msg { return OpenProfilesMsg{} }
	case "ctrl+x":
		return m, func() tea.Msg { return DisconnectMsg{} }
	case "ctrl+r":
		return m, func() tea.Msg { return ReconnectMsg{} }
	case "f2":
		return m.openEditor(editorInsert)
	case "f3":
		return m.openEditor(editorUpdate)
	case "f4":
		return m.openEditor(editorDelete)
	case "tab":
		m.toggleFocus()
	case "enter":
		if m.focus == FocusPreview && len(m.preview.Rows) > 0 {
			m.OpenRecordView()
		}
	case "up", "k":
		return m.moveUp()
	case "down", "j":
		return m.moveDown()
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

	return m, nil
}

func (m Model) openEditor(mode editorMode) (Model, tea.Cmd) {
	m.OpenEditor(mode)
	m.SetEditorSize(m.width, m.height)
	return m, nil
}

func (m *Model) toggleFocus() {
	if m.focus == FocusTables {
		m.focus = FocusPreview
		return
	}
	m.focus = FocusTables
}

func (m Model) moveUp() (Model, tea.Cmd) {
	if m.focus == FocusTables {
		if m.selected > 0 {
			m.selected--
			return m, m.selectCmd()
		}
		return m, nil
	}

	if m.selectedRow > 0 {
		m.selectedRow--
		m.ensureRowVisible(m.previewViewportRows())
	}
	return m, nil
}

func (m Model) moveDown() (Model, tea.Cmd) {
	if m.focus == FocusTables {
		if m.selected < len(m.tables)-1 {
			m.selected++
			return m, m.selectCmd()
		}
		return m, nil
	}

	if m.selectedRow < len(m.preview.Rows)-1 {
		m.selectedRow++
		m.ensureRowVisible(m.previewViewportRows())
	}
	return m, nil
}
