package ui

import (
	"os/exec"
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/danshort/lectern/internal/config"
)

func (m Model) updateViewer(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {

	case "q", "ctrl+c":
		return m, tea.Quit

	case "i":
		m.prevMode = m.mode
		m.setMode(ModeViewingConfig)
		return m.commitStateChange()

	case "a", "esc":
		// A foreign worktree change returns to the worktrees view it was
		// opened from, not the index.
		if m.viewingWorktreeChange {
			m.viewingWorktreeChange = false
			m.renderCache = make(map[Tab]string)
			m.setMode(ModeWorktrees)
			m.vp.SetHeight(m.contentHeight())
			m.refreshWorktreesViewport()
			return m, nil
		}
		m.enterIndex()
		return m, nil

	case "h":
		if len(m.project.Changes) > 0 {
			m.viewer.changeIdx = (m.viewer.changeIdx - 1 + len(m.project.Changes)) % len(m.project.Changes)
			m.renderCache = make(map[Tab]string)
			m.loadTaskItems()
			m.viewer.tab = m.defaultTab()
			m.viewer.specIdx = 0
			return m.commitStateChange()
		}

	case "l":
		if len(m.project.Changes) > 0 {
			m.viewer.changeIdx = (m.viewer.changeIdx + 1) % len(m.project.Changes)
			m.renderCache = make(map[Tab]string)
			m.loadTaskItems()
			m.viewer.tab = m.defaultTab()
			m.viewer.specIdx = 0
			return m.commitStateChange()
		}

	case "1":
		if m.tabAvailable(TabProposal) {
			m.viewer.tab = TabProposal
			return m.commitStateChange()
		}
	case "2":
		if m.tabAvailable(TabDesign) {
			m.viewer.tab = TabDesign
			return m.commitStateChange()
		}
	case "3":
		// Plain primary-tab selector like 1/2/4. specIdx is preserved so
		// returning to the specs tab keeps the last-viewed spec; it only resets
		// on a change switch (h/l). Spec switching is the job of [ / ] below.
		if m.tabAvailable(TabSpecs) {
			m.viewer.tab = TabSpecs
			return m.commitStateChange()
		}
	case "4":
		if m.tabAvailable(TabTasks) {
			m.viewer.tab = TabTasks
			return m.commitStateChange()
		}

	case "[":
		// Secondary sub-navigation: previous spec on the specs chip row.
		return m.moveSpec(-1)
	case "]":
		// Secondary sub-navigation: next spec on the specs chip row.
		return m.moveSpec(1)

	case "tab", "right":
		// Right arrow mirrors Tab as secondary tab navigation.
		nxt := m.nextAvailableTab(m.viewer.tab, 1)
		if nxt != m.viewer.tab {
			m.viewer.tab = nxt
			return m.commitStateChange()
		}
	case "shift+tab", "left":
		// Left arrow mirrors Shift+Tab as secondary tab navigation.
		prv := m.nextAvailableTab(m.viewer.tab, -1)
		if prv != m.viewer.tab {
			m.viewer.tab = prv
			return m.commitStateChange()
		}

	case "j", "down":
		// The interactive task cursor only exists in normal mode; archived
		// changes render tasks as read-only markdown, so scroll instead.
		if m.viewer.tab == TabTasks && m.mode == ModeNormal {
			m.moveCursorDown()
			m.refreshTasksViewport()
		} else {
			m.vp.ScrollDown(1)
		}

	case "k", "up":
		if m.viewer.tab == TabTasks && m.mode == ModeNormal {
			m.moveCursorUp()
			m.refreshTasksViewport()
		} else {
			m.vp.ScrollUp(1)
		}

	case "space":
		if m.viewer.tab == TabTasks && m.mode == ModeNormal {
			return m, m.doToggle()
		}

	case "e":
		if m.tabAvailable(m.viewer.tab) {
			if path := m.artifactPath(); path != "" {
				return m, m.openInEditor(path)
			}
		}
	}
	return m, nil
}

// moveSpec advances the visible spec on the specs tab by delta (with
// wrap-around), when the current change has more than one spec. It is the
// secondary sub-navigation for the specs chip row (`[` / `]`) and is a no-op on
// any other tab or when the change has a single spec. Works identically for
// active and archived changes via current().
func (m Model) moveSpec(delta int) (tea.Model, tea.Cmd) {
	if m.viewer.tab != TabSpecs {
		return m, nil
	}
	ch := m.current()
	if ch == nil || len(ch.SpecFiles) < 2 {
		return m, nil
	}
	n := len(ch.SpecFiles)
	m.viewer.specIdx = (m.viewer.specIdx + delta + n) % n
	delete(m.renderCache, TabSpecs)
	return m.commitStateChange()
}

// openInEditor opens path using the resolved editor.open_with preference
// (internal/config.ResolveOpener). A terminal opener ($EDITOR/vi or a custom
// command) yields the terminal via tea.ExecProcess and posts an editorReturnMsg
// on exit; the "system" handler is launched detached (the TUI keeps running and
// the save is picked up by the normal reload). A missing or unlaunchable opener
// sets m.errMsg instead of failing silently. Shared by the change and spec
// viewers.
func (m *Model) openInEditor(path string) tea.Cmd {
	op := config.ResolveOpener(m.editorOpenWith)

	// Verify the opener exists before launching, so a bad config/$EDITOR can't
	// blank the screen (terminal mode) or silently no-op (detached mode).
	if _, err := exec.LookPath(op.Name); err != nil {
		m.errMsg = "editor not found: " + op.Name
		return clearErrAfter()
	}

	args := append(append([]string{}, op.Args...), path)
	// #nosec G204 -- opens the user's own editor/handler on a file in their own
	// project. No shell is invoked (exec.Command does not interpret shell
	// metacharacters), so this is not command injection.
	cmd := exec.Command(op.Name, args...)

	if op.Mode == config.OpenDetached {
		// Fire-and-forget: launch the GUI handler, don't yield the terminal.
		if err := cmd.Start(); err != nil {
			m.errMsg = "could not open: " + err.Error()
			return clearErrAfter()
		}
		return nil
	}

	// Terminal editor: yield the terminal and resume on exit.
	return tea.ExecProcess(cmd, func(err error) tea.Msg {
		return editorReturnMsg{err: err}
	})
}

// clearErrAfter clears the transient error message after a short delay,
// matching the tasks-toggle error pattern.
func clearErrAfter() tea.Cmd {
	return tea.Tick(3*time.Second, func(time.Time) tea.Msg { return errClearMsg{} })
}
