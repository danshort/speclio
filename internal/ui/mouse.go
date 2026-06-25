package ui

import tea "charm.land/bubbletea/v2"

func (m Model) handleMouseWheel(msg tea.MouseWheelMsg) (tea.Model, tea.Cmd) {
	switch msg.Button {
	case tea.MouseWheelUp:
		if m.mode == ModeIndex {
			if m.index.Cursor > 0 {
				m.index.Cursor--
			}
			m.refreshIndexViewport()
			return m, nil
		}
		if m.tab == TabTasks && m.mode == ModeNormal {
			m.moveCursorUp()
			m.refreshTasksViewport()
			return m, nil
		}
		m.vp.ScrollUp(3)
		return m, nil
	case tea.MouseWheelDown:
		if m.mode == ModeIndex {
			if m.index.Cursor < m.visibleItemCount()-1 {
				m.index.Cursor++
			}
			m.refreshIndexViewport()
			return m, nil
		}
		if m.tab == TabTasks && m.mode == ModeNormal {
			m.moveCursorDown()
			m.refreshTasksViewport()
			return m, nil
		}
		m.vp.ScrollDown(3)
		return m, nil
	}
	return m, nil
}

func (m Model) handleMouseClick(msg tea.MouseClickMsg) (tea.Model, tea.Cmd) {
	if msg.Button != tea.MouseLeft {
		return m, nil
	}

	if m.mode == ModeIndex {
		top := m.viewportTop()
		if msg.Y < top || msg.Y >= top+m.vp.Height() {
			return m, nil
		}
		contentLine := msg.Y - top + m.vp.YOffset()
		idx, found := m.indexItemAtContentLine(contentLine)
		if !found {
			return m, nil
		}
		if m.index.FilterIndices != nil {
			cursorOOB := m.index.Cursor < 0 || m.index.Cursor >= len(m.index.FilterIndices)
			if cursorOOB || m.index.FilterIndices[m.index.Cursor] != idx {
				for ci, ri := range m.index.FilterIndices {
					if ri == idx {
						m.index.Cursor = ci
						break
					}
				}
				m.refreshIndexViewport()
				return m, nil
			}
		} else {
			if m.index.Cursor != idx {
				m.index.Cursor = idx
				m.refreshIndexViewport()
				return m, nil
			}
		}
		return m.clickIndexItem(idx)
	}

	if m.mode != ModeNormal && m.mode != ModeViewingArchive {
		return m, nil
	}

	if msg.Y == m.chromeRowIndex(rowHeader) {
		m.enterIndex()
		return m, nil
	}

	tabRow := m.chromeRowIndex(rowTabBar)
	if tabRow < 0 || msg.Y != tabRow {
		return m, nil
	}

	x := 1
	for t := Tab(0); t < tabCount; t++ {
		w := len(tabLabels[t]) + 2
		if msg.X >= x && msg.X <= x+w-1 {
			if !m.tabAvailable(t) {
				return m, nil
			}
			if t == TabSpecs && m.tab == TabSpecs {
				ch := m.current()
				if ch != nil && len(ch.SpecFiles) > 1 {
					m.specIdx = (m.specIdx + 1) % len(ch.SpecFiles)
					delete(m.renderCache, TabSpecs)
				}
			} else {
				m.tab = t
				if t == TabSpecs {
					m.specIdx = 0
				}
			}
			m.vp.SetHeight(m.contentHeight())
			return m, m.loadViewport()
		}
		x += w + 1
	}

	return m, nil
}

func (m Model) clickIndexItem(idx int) (tea.Model, tea.Cmd) {
	item := m.index.Items[idx]
	// Clicking a spec toggles its expansion — the mouse analogue of Space.
	// Every other kind opens/navigates, identical to the keyboard Enter path.
	if item.kind == indexKindSpec {
		m.renderCache = make(map[Tab]string)
		m.index.ExpandedSpecs[item.idx] = !m.index.ExpandedSpecs[item.idx]
		m.toggleExpansion(indexKindSpec, item.idx)
		return m, nil
	}
	return m.activateIndexItem(item)
}
