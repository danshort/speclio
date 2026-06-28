package ui

import (
	"time"

	"charm.land/bubbles/v2/viewport"
	tea "charm.land/bubbletea/v2"
	"github.com/danshort/lectern/internal/openspec"
)

// Receiver convention
//
// The Update entry point and the dispatchKey / update* handlers are VALUE
// receivers (func (m Model) …): each mutates its own copy of m and MUST return
// it (return m, cmd). A handler that mutates m and forgets to return it silently
// drops the change — there is no compiler warning, so always thread m through.
//
// Helpers with a POINTER receiver (func (m *Model) …, e.g. setMode, enterIndex,
// mergeReloadedChange, loadTaskItems) mutate in place and are called for their
// effect. A value-receiver handler may call them freely because its receiver
// copy is addressable. Do not mix the two forms in one function: a method that
// both mutates via *Model and returns a Model is a trap.

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		contentH := m.contentHeight()
		if !m.vpReady {
			m.vp = viewport.New(viewport.WithWidth(m.innerWidth()), viewport.WithHeight(contentH))
			m.vpReady = true
		} else {
			m.vp.SetWidth(m.innerWidth())
			m.vp.SetHeight(contentH)
		}
		m.renderCache = make(map[Tab]string)
		return m, m.loadViewport()

	case renderedMsg:
		m.renderCache[msg.tab] = msg.content
		m.loading = false
		if m.viewer.tab == msg.tab {
			m.vp.SetContent(msg.content)
			m.vp.GotoTop()
		}
		return m, nil

	case specRenderedMsg:
		m.loading = false
		if m.mode == ModeViewingSpec {
			m.vp.SetContent(msg.content)
			if msg.jumpLine > 0 {
				m.vp.SetYOffset(msg.jumpLine)
			} else {
				m.vp.GotoTop()
			}
		}
		return m, nil

	case renderedConfigMsg:
		m.loading = false
		if m.mode == ModeViewingConfig {
			m.vp.SetContent(msg.content)
			m.vp.GotoTop()
		}
		return m, nil

	case tickMsg:
		cmd := m.handleTick()
		nextTick := tea.Tick(500*time.Millisecond, func(t time.Time) tea.Msg { return tickMsg(t) })
		return m, tea.Batch(nextTick, cmd)

	case editorReturnMsg:
		if msg.err != nil {
			// Terminal editor failed to launch/run — surface it; nothing to reload.
			m.errMsg = "editor: " + msg.err.Error()
			return m, clearErrAfter()
		}
		if m.mode == ModeViewingSpec {
			// A project spec was edited: reload specs from disk so the change
			// shows immediately, staying in ModeViewingSpec. Clamp the cursor
			// in case the spec list shrank.
			if specs, err := m.loader.LoadProjectSpecsFrom(m.root); err == nil {
				m.projectSpecs = specs
				if m.spec.Cursor >= len(m.projectSpecs) {
					m.spec.Cursor = len(m.projectSpecs) - 1
				}
				if m.spec.Cursor < 0 {
					m.spec.Cursor = 0
				}
			}
			return m, m.loadViewport()
		}
		if m.viewingWorktreeChange {
			// A foreign worktree change was edited. Reload it into
			// worktreeViewChange — NOT m.project.Changes[m.viewer.changeIdx], which
			// mergeReloadedChange would target — so the rooted project's state
			// is never overwritten with the sibling worktree's content.
			m.worktreeViewChange = m.loader.ReloadChange(m.worktreeViewChange)
			delete(m.renderCache, m.viewer.tab)
			return m, m.loadViewport()
		}
		ch := m.current()
		if ch != nil {
			var cursorText string
			if m.tasks.Cursor < len(m.tasks.Items) && m.tasks.Items[m.tasks.Cursor].Kind == openspec.KindTask {
				cursorText = m.tasks.Items[m.tasks.Cursor].Text
			}
			fresh := m.loader.ReloadChange(*ch)
			tasksChanged, _ := m.mergeReloadedChange(fresh)
			if tasksChanged {
				m.tasks.Cursor = openspec.FindCursorByText(m.tasks.Items, cursorText)
			}
			// Always drop the current tab's cached render: an unsaved exit
			// leaves content unchanged (so mergeReloadedChange keeps the
			// cache), but the viewport may have been resized during terminal
			// re-entry, leaving the cached ANSI wrapped at the old width.
			delete(m.renderCache, m.viewer.tab)
		}
		return m, m.loadViewport()

	case errClearMsg:
		m.errMsg = ""
		return m, nil

	case tea.MouseWheelMsg:
		return m.handleMouseWheel(msg)

	case tea.MouseClickMsg:
		return m.handleMouseClick(msg)

	case tea.KeyPressMsg:
		return m.dispatchKey(msg)
	}
	return m, nil
}

func (m Model) dispatchKey(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	// The keyboard-help overlay is a transient layer over the current screen.
	// While it is open, only the dismiss keys act; every other key is swallowed
	// so the underlying screen stays inert.
	if m.helpOpen {
		switch msg.String() {
		case "?", "esc", "q":
			m.helpOpen = false
		case "ctrl+c":
			// Honor the quit shortcut the overlay itself advertises.
			return m, tea.Quit
		}
		return m, nil
	}
	// `?` opens the overlay from any mode. The index filter input takes
	// precedence so a `?` can still be typed into a filter query.
	if msg.String() == "?" && !(m.mode == ModeIndex && m.index.FilterActive) {
		m.helpOpen = true
		return m, nil
	}

	switch m.mode {
	case ModeNormal, ModeViewingArchive:
		return m.updateViewer(msg)
	case ModeIndex:
		return m.updateIndex(msg)
	case ModeViewingSpec:
		return m.updateSpec(msg)
	case ModeViewingConfig:
		return m.updateConfig(msg)
	case ModeWorktrees:
		return m.updateWorktrees(msg)
	}
	return m, nil
}

func (m Model) commitStateChange() (tea.Model, tea.Cmd) {
	m.vp.SetHeight(m.contentHeight())
	return m, m.loadViewport()
}

// setMode transitions the model to next and is the ONLY place m.mode is
// assigned. It enforces the state invariants each mode owns so a caller cannot
// leave stale state from the outgoing mode:
//   - leaving ModeViewingSpec clears the requirement-focus state (FocusMode,
//     JumpTarget, ReqCursor) so it cannot leak into a later spec view;
//   - entering a tabbed mode (ModeNormal / ModeViewingArchive) clamps the
//     selected tab to an available one and the spec index into range;
//   - entering ModeViewingSpec clamps the spec cursor into the project specs.
//
// It deliberately does NOT touch renderCache: cache invalidation policy stays
// with callers, which vary it on purpose (tab switches keep the cache for an
// instant return; moveSpec drops only TabSpecs; activateIndexItem drops all).
// The mode is assigned before the clamps run so current()/tabAvailable() resolve
// against the destination mode. Cursor clamps are safety nets — no-ops when the
// caller has already set a valid cursor.
func (m *Model) setMode(next Mode) {
	prev := m.mode
	m.mode = next

	if prev == ModeViewingSpec && next != ModeViewingSpec {
		m.spec.FocusMode = false
		m.spec.JumpTarget = ""
		m.spec.ReqCursor = 0
	}

	switch next {
	case ModeNormal, ModeViewingArchive:
		if ch := m.current(); ch != nil {
			if !m.tabAvailable(m.viewer.tab) {
				m.viewer.tab = m.defaultTab()
			}
			if m.viewer.specIdx < 0 || m.viewer.specIdx >= len(ch.SpecFiles) {
				m.viewer.specIdx = 0
			}
		}
	case ModeViewingSpec:
		switch n := len(m.projectSpecs); {
		case n == 0, m.spec.Cursor < 0:
			m.spec.Cursor = 0
		case m.spec.Cursor >= n:
			m.spec.Cursor = n - 1
		}
	}
}
