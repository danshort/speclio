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

	// Spec chip row: present (chromeRowIndex >= 0) only on the specs tab when
	// the change has specs. Clicking a chip selects that spec; clicks between or
	// outside chips are ignored.
	if subRow := m.chromeRowIndex(rowSubnav); subRow >= 0 && msg.Y == subRow {
		ranges := m.specRanges()
		for i := range ranges {
			if msg.X < ranges[i].start || msg.X > ranges[i].end {
				continue
			}
			if i != m.specIdx {
				m.specIdx = i
				delete(m.renderCache, TabSpecs)
				m.vp.SetHeight(m.contentHeight())
				return m, m.loadViewport()
			}
			return m, nil
		}
		return m, nil
	}

	tabRow := m.chromeRowIndex(rowTabBar)
	if tabRow < 0 || msg.Y != tabRow {
		return m, nil
	}

	ranges := m.tabRanges()
	for t := Tab(0); t < tabCount; t++ {
		if msg.X < ranges[t].start || msg.X > ranges[t].end {
			continue
		}
		if !m.tabAvailable(t) {
			return m, nil
		}
		// Selecting the specs tab preserves specIdx (the last-viewed spec);
		// switching specs is done via the chip row or [ / ].
		m.tab = t
		m.vp.SetHeight(m.contentHeight())
		return m, m.loadViewport()
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
