package ui

import (
	"sort"
	"strings"

	tea "charm.land/bubbletea/v2"
	"github.com/danshort/lectern/internal/openspec"
)

// enterWorktrees discovers the repository's git worktrees once, loads each
// worktree's active changes, builds the navigable item list, and switches to
// ModeWorktrees. When git is unavailable or the project is not inside a working
// tree, the view degrades to a single explanatory line.
func (m *Model) enterWorktrees() {
	m.worktrees = worktreesState{}
	m.viewingWorktreeChange = false

	wts, err := openspec.ListWorktrees(m.root)
	if err != nil {
		m.worktrees.Available = false
		m.worktrees.Message = "Worktrees unavailable: git is not on PATH or this is not a git working tree."
		m.mode = ModeWorktrees
		m.vp.SetHeight(m.contentHeight())
		m.refreshWorktreesViewport()
		return
	}

	m.worktrees.Available = true
	var entries []worktreeEntry
	for _, wt := range wts {
		if wt.Bare {
			continue // bare worktrees have no working tree to load changes from
		}
		var changes []openspec.Change
		if p, err := m.loader.LoadFrom(wt.Path); err == nil {
			changes = p.Changes
		}
		entries = append(entries, worktreeEntry{wt: wt, changes: changes})
	}
	sortWorktreeEntriesCurrentFirst(entries)
	m.worktrees.Entries = entries
	m.buildWorktreeItems()
	m.worktrees.Cursor = 0
	m.mode = ModeWorktrees
	m.vp.SetHeight(m.contentHeight())
	m.refreshWorktreesViewport()
}

// sortWorktreeEntriesCurrentFirst lists the current worktree first and
// preserves git's order for the rest.
func sortWorktreeEntriesCurrentFirst(entries []worktreeEntry) {
	sort.SliceStable(entries, func(i, j int) bool {
		return entries[i].wt.IsCurrent && !entries[j].wt.IsCurrent
	})
}

// buildWorktreeItems flattens foreign worktrees' changes into the navigable
// item list. The current worktree's changes are surfaced on the index already,
// so they are rendered for completeness but are not navigable/openable here.
func (m *Model) buildWorktreeItems() {
	m.worktrees.Items = nil
	for wi := range m.worktrees.Entries {
		e := m.worktrees.Entries[wi]
		if e.wt.IsCurrent {
			continue
		}
		for ci := range e.changes {
			m.worktrees.Items = append(m.worktrees.Items, worktreeItem{wtIdx: wi, changeIdx: ci})
		}
	}
}

func (m *Model) currentWorktreeItem() (worktreeItem, bool) {
	if m.worktrees.Cursor >= 0 && m.worktrees.Cursor < len(m.worktrees.Items) {
		return m.worktrees.Items[m.worktrees.Cursor], true
	}
	return worktreeItem{}, false
}

// worktreeHeader builds the per-worktree header label: the branch name, or a
// short HEAD SHA when detached, with a (current) badge and locked/prunable
// annotations where applicable.
func worktreeHeader(wt openspec.Worktree) string {
	label := wt.Branch
	if label == "" {
		if len(wt.Head) >= 7 {
			label = wt.Head[:7]
		} else {
			label = wt.Head
		}
	}
	if wt.IsCurrent {
		label += "  (current)"
	}
	var ann []string
	if wt.Locked {
		ann = append(ann, "locked")
	}
	if wt.Prunable {
		ann = append(ann, "prunable")
	}
	if len(ann) > 0 {
		label += "  [" + strings.Join(ann, ", ") + "]"
	}
	return label
}

func (m *Model) renderWorktreesContent() (string, int) {
	contentWidth := m.width - 2
	var sb strings.Builder
	line := 0
	cursorLine := 0

	sb.WriteString("\n")
	line++

	if !m.worktrees.Available {
		sb.WriteString("  " + helpStyle.Render(m.worktrees.Message) + "\n")
		return sb.String(), 0
	}

	if len(m.worktrees.Entries) == 0 {
		sb.WriteString(helpStyle.Render("  No worktrees found") + "\n")
		return sb.String(), 0
	}

	cur, hasCur := m.currentWorktreeItem()

	for wi := range m.worktrees.Entries {
		e := m.worktrees.Entries[wi]
		sb.WriteString("  " + sectionStyle.Render(worktreeHeader(e.wt)) + "\n")
		line++

		if len(e.changes) == 0 {
			sb.WriteString(helpStyle.Render("    (no active changes)") + "\n")
			line++
		} else {
			for ci := range e.changes {
				isCursor := hasCur && cur.wtIdx == wi && cur.changeIdx == ci
				if isCursor {
					cursorLine = line
				}
				sb.WriteString("  " + m.renderActiveItem(e.changes[ci], isCursor, contentWidth-2) + "\n")
				line++
			}
		}

		sb.WriteString("\n")
		line++
	}

	return sb.String(), cursorLine
}

func (m *Model) refreshWorktreesViewport() {
	content, cursorLine := m.renderWorktreesContent()
	m.vp.SetContent(content)
	if cursorLine < m.vp.YOffset() {
		m.vp.SetYOffset(cursorLine)
	} else if cursorLine >= m.vp.YOffset()+m.vp.Height() {
		m.vp.SetYOffset(cursorLine - m.vp.Height() + 1)
	}
}

func (m Model) updateWorktrees(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {

	case "q", "ctrl+c":
		return m, tea.Quit

	case "a", "esc":
		m.enterIndex()
		return m, nil

	case "j", "down":
		if m.worktrees.Cursor < len(m.worktrees.Items)-1 {
			m.worktrees.Cursor++
			m.refreshWorktreesViewport()
		}

	case "k", "up":
		if m.worktrees.Cursor > 0 {
			m.worktrees.Cursor--
			m.refreshWorktreesViewport()
		}

	case "enter":
		return m.openWorktreeChange()
	}
	return m, nil
}

// openWorktreeChange opens the change under the cursor read-only, reusing the
// ModeViewingArchive path. The change is stored in its own field so no
// assumptions about project.Changes indices are violated, and all mutation
// remains gated on ModeNormal.
func (m Model) openWorktreeChange() (tea.Model, tea.Cmd) {
	item, ok := m.currentWorktreeItem()
	if !ok {
		return m, nil
	}
	e := m.worktrees.Entries[item.wtIdx]
	if item.changeIdx >= len(e.changes) {
		return m, nil
	}
	ch := m.loader.ReloadChange(e.changes[item.changeIdx])
	m.worktreeViewChange = ch
	m.viewingWorktreeChange = true
	m.renderCache = make(map[Tab]string)
	m.tab = firstAvailableTab(ch)
	m.specIdx = 0
	m.prevMode = ModeWorktrees
	m.mode = ModeViewingArchive
	return m.commitStateChange()
}

// pollWorktrees refreshes the loaded changes' task content on the tick so the
// progress bars track agents live. The worktree set itself is captured on entry
// and not re-enumerated here.
func (m *Model) pollWorktrees() tea.Cmd {
	if !m.worktrees.Available {
		return nil
	}
	changed := false
	for wi := range m.worktrees.Entries {
		e := &m.worktrees.Entries[wi]
		for ci := range e.changes {
			fresh := m.loader.ReloadChange(e.changes[ci])
			if fresh.Tasks.Present != e.changes[ci].Tasks.Present || fresh.Tasks.Content != e.changes[ci].Tasks.Content {
				e.changes[ci].Tasks = fresh.Tasks
				changed = true
			}
		}
	}
	if changed {
		m.refreshWorktreesViewport()
	}
	return nil
}

// pollWorktreeChange reloads the open foreign change on the tick. When the
// visible artifact's content changes, its render cache is invalidated and the
// viewport is refreshed.
func (m *Model) pollWorktreeChange() tea.Cmd {
	fresh := m.loader.ReloadChange(m.worktreeViewChange)
	dirty := false
	if fresh.Proposal.Content != m.worktreeViewChange.Proposal.Content {
		delete(m.renderCache, TabProposal)
		dirty = dirty || m.tab == TabProposal
	}
	if fresh.Design.Content != m.worktreeViewChange.Design.Content {
		delete(m.renderCache, TabDesign)
		dirty = dirty || m.tab == TabDesign
	}
	if fresh.Specs.Content != m.worktreeViewChange.Specs.Content {
		delete(m.renderCache, TabSpecs)
		dirty = dirty || m.tab == TabSpecs
	}
	if fresh.Tasks.Content != m.worktreeViewChange.Tasks.Content {
		delete(m.renderCache, TabTasks)
		dirty = dirty || m.tab == TabTasks
	}
	m.worktreeViewChange = fresh
	if dirty {
		return m.loadViewport()
	}
	return nil
}
