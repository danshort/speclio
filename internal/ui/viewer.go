package ui

import (
	"os"
	"os/exec"
	"strings"

	tea "charm.land/bubbletea/v2"
)

func (m Model) updateViewer(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {

	case "q", "ctrl+c":
		return m, tea.Quit

	case "i":
		m.prevMode = m.mode
		m.mode = ModeViewingConfig
		return m.commitStateChange()

	case "a", "esc":
		// A foreign worktree change returns to the worktrees view it was
		// opened from, not the index.
		if m.viewingWorktreeChange {
			m.viewingWorktreeChange = false
			m.renderCache = make(map[Tab]string)
			m.mode = ModeWorktrees
			m.vp.SetHeight(m.contentHeight())
			m.refreshWorktreesViewport()
			return m, nil
		}
		m.enterIndex()
		return m, nil

	case "h":
		if len(m.project.Changes) > 0 {
			m.changeIdx = (m.changeIdx - 1 + len(m.project.Changes)) % len(m.project.Changes)
			m.renderCache = make(map[Tab]string)
			m.loadTaskItems()
			m.tab = m.defaultTab()
			m.specIdx = 0
			return m.commitStateChange()
		}

	case "l":
		if len(m.project.Changes) > 0 {
			m.changeIdx = (m.changeIdx + 1) % len(m.project.Changes)
			m.renderCache = make(map[Tab]string)
			m.loadTaskItems()
			m.tab = m.defaultTab()
			m.specIdx = 0
			return m.commitStateChange()
		}

	case "1":
		if m.tabAvailable(TabProposal) {
			m.tab = TabProposal
			return m.commitStateChange()
		}
	case "2":
		if m.tabAvailable(TabDesign) {
			m.tab = TabDesign
			return m.commitStateChange()
		}
	case "3":
		// Plain primary-tab selector like 1/2/4. specIdx is preserved so
		// returning to the specs tab keeps the last-viewed spec; it only resets
		// on a change switch (h/l). Spec switching is the job of [ / ] below.
		if m.tabAvailable(TabSpecs) {
			m.tab = TabSpecs
			return m.commitStateChange()
		}
	case "4":
		if m.tabAvailable(TabTasks) {
			m.tab = TabTasks
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
		nxt := m.nextAvailableTab(m.tab, 1)
		if nxt != m.tab {
			m.tab = nxt
			return m.commitStateChange()
		}
	case "shift+tab", "left":
		// Left arrow mirrors Shift+Tab as secondary tab navigation.
		prv := m.nextAvailableTab(m.tab, -1)
		if prv != m.tab {
			m.tab = prv
			return m.commitStateChange()
		}

	case "j", "down":
		// The interactive task cursor only exists in normal mode; archived
		// changes render tasks as read-only markdown, so scroll instead.
		if m.tab == TabTasks && m.mode == ModeNormal {
			m.moveCursorDown()
			m.refreshTasksViewport()
		} else {
			m.vp.ScrollDown(1)
		}

	case "k", "up":
		if m.tab == TabTasks && m.mode == ModeNormal {
			m.moveCursorUp()
			m.refreshTasksViewport()
		} else {
			m.vp.ScrollUp(1)
		}

	case "space":
		if m.tab == TabTasks && m.mode == ModeNormal {
			return m, m.doToggle()
		}

	case "e":
		if m.tabAvailable(m.tab) {
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
	if m.tab != TabSpecs {
		return m, nil
	}
	ch := m.current()
	if ch == nil || len(ch.SpecFiles) < 2 {
		return m, nil
	}
	n := len(ch.SpecFiles)
	m.specIdx = (m.specIdx + delta + n) % n
	delete(m.renderCache, TabSpecs)
	return m.commitStateChange()
}

// openInEditor launches the user's $EDITOR (falling back to vi) on path and
// posts an editorReturnMsg when the editor exits. The TUI yields the terminal
// via tea.ExecProcess and resumes on return. Shared by the change viewer and
// the spec viewer.
func (m *Model) openInEditor(path string) tea.Cmd {
	// Split EDITOR so values like "code --wait" or "emacs -nw" work.
	fields := strings.Fields(os.Getenv("EDITOR"))
	if len(fields) == 0 {
		fields = []string{"vi"}
	}
	args := append(fields[1:], path)
	// #nosec G204 G702 -- by design: opens the user's own $EDITOR on a file
	// in their own project. No shell is invoked (exec.Command does not
	// interpret shell metacharacters), so this is not command injection.
	cmd := exec.Command(fields[0], args...)
	return tea.ExecProcess(cmd, func(err error) tea.Msg {
		return editorReturnMsg{}
	})
}
