package ui

import (
	"fmt"
	"strings"

	"charm.land/lipgloss/v2"
	"github.com/danshort/lectern/internal/openspec"
)

// rowID identifies a chrome row so hit-testing can locate it without hardcoded
// screen offsets.
type rowID int

const (
	rowBoxTop rowID = iota
	rowHeader
	rowTabBar
	rowSep
	rowSubnav
	rowHelp
	rowBoxBottom
)

// layoutHasTabBar reports whether the current mode renders the artifact tab bar
// (and therefore can render the optional spec subnav).
func (m *Model) layoutHasTabBar() bool {
	return m.mode == ModeNormal || m.mode == ModeViewingArchive
}

// chromeRowsAbove returns the IDs of the chrome rows above the content viewport,
// in order. This pure structural list (no rendering) is the single source for
// viewport sizing (contentHeight), hit-testing (viewportTop / chromeRowIndex),
// and render order (viewportLayout).
func (m *Model) chromeRowsAbove() []rowID {
	ids := []rowID{rowBoxTop, rowHeader}
	if m.layoutHasTabBar() {
		ids = append(ids, rowTabBar)
	}
	ids = append(ids, rowSep)
	if m.layoutHasTabBar() && m.hasSpecSubnav() {
		ids = append(ids, rowSubnav)
	}
	return ids
}

// chromeRowsBelow returns the IDs of the chrome rows below the content viewport.
func (m *Model) chromeRowsBelow() []rowID {
	return []rowID{rowSep, rowHelp, rowBoxBottom}
}

// chromeRowCount is the number of non-viewport rows in the current layout.
func (m *Model) chromeRowCount() int { return len(m.chromeRowsAbove()) + len(m.chromeRowsBelow()) }

// viewportTop is the screen row (0-based) where the content viewport begins.
func (m *Model) viewportTop() int { return len(m.chromeRowsAbove()) }

// chromeRowIndex returns the screen row (0-based) of the given chrome row above
// the viewport, or -1 if that row is absent in the current layout.
func (m *Model) chromeRowIndex(id rowID) int {
	for i, r := range m.chromeRowsAbove() {
		if r == id {
			return i
		}
	}
	return -1
}

// renderChromeRow renders a single chrome row by ID. Only called from View()
// (where the project is populated); contentHeight/hit-testing use the ID lists
// above and never render.
func (m *Model) renderChromeRow(id rowID) string {
	switch id {
	case rowBoxTop:
		return m.boxTop()
	case rowHeader:
		return m.addBorderSides(m.renderHeader())
	case rowTabBar:
		return m.addBorderSides(m.renderTabBar())
	case rowSep:
		return m.boxInnerSep()
	case rowSubnav:
		return m.addBorderSides(m.renderSpecSubnav())
	case rowHelp:
		return m.addBorderSides(m.renderHelpBar())
	case rowBoxBottom:
		return m.boxBottom()
	}
	return ""
}

// viewportLayout assembles the full view — chrome above, the viewport, chrome
// below — from the same row-ID lists used to size it.
func (m *Model) viewportLayout() string {
	rows := make([]string, 0, m.chromeRowCount()+1)
	for _, id := range m.chromeRowsAbove() {
		rows = append(rows, m.renderChromeRow(id))
	}
	rows = append(rows, m.addBorderSides(m.vp.View()))
	for _, id := range m.chromeRowsBelow() {
		rows = append(rows, m.renderChromeRow(id))
	}
	return strings.Join(rows, "\n")
}

func (m *Model) emptyViewContent() string {
	return m.boxTop() + "\n" +
		m.addBorderSides(headerStyle.Render(m.project.Name)+
			"\n\n\n  No active changes. Create one with /opsx:propose\n"+
			helpStyle.Render("\n  a/Esc: index  q: quit")) + "\n" +
		m.boxInnerSep() + "\n" +
		m.addBorderSides(m.renderHelpBar()) + "\n" +
		m.boxBottom()
}

func (m *Model) renderHeader() string {
	if m.mode == ModeViewingConfig {
		return headerStyle.Width(m.innerWidth()).Render(m.project.Name + "  ·  project config")
	}
	if m.mode == ModeIndex {
		return headerStyle.Width(m.innerWidth()).Render(m.project.Name + "  ·  index")
	}
	if m.mode == ModeWorktrees {
		return headerStyle.Width(m.innerWidth()).Render(m.project.Name + "  ·  worktrees")
	}
	if m.mode == ModeViewingSpec {
		specName := ""
		if m.specViewer.Cursor < len(m.projectSpecs) {
			specName = m.projectSpecs[m.specViewer.Cursor].Name
		}
		if m.specViewer.FocusMode && m.specViewer.Cursor < len(m.projectSpecs) {
			ps := m.projectSpecs[m.specViewer.Cursor]
			return headerStyle.Width(m.innerWidth()).Render(
				fmt.Sprintf("%s  ·  %s  ·  Req %d/%d", m.project.Name, specName, m.specViewer.ReqCursor+1, len(ps.RequirementNames)),
			)
		}
		return headerStyle.Width(m.innerWidth()).Render(
			fmt.Sprintf("%s  ·  %s  [spec]", m.project.Name, specName),
		)
	}
	ch := m.current()
	if ch == nil {
		return headerStyle.Render(m.project.Name)
	}
	if m.mode == ModeViewingArchive {
		tag := "[archive]"
		if m.viewingWorktreeChange {
			tag = "[worktree]"
		}
		return headerStyle.Width(m.innerWidth()).Render(
			fmt.Sprintf("%s  ·  %s  %s", m.project.Name, ch.Name, tag),
		)
	}
	nav := fmt.Sprintf("[%d/%d]", m.changeIdx+1, len(m.project.Changes))
	return headerStyle.Width(m.innerWidth()).Render(
		fmt.Sprintf("%s  ·  %s  %s", m.project.Name, ch.Name, nav),
	)
}

func (m *Model) renderTabBar() string {
	parts := make([]string, 0, tabCount)
	for t := Tab(0); t < tabCount; t++ {
		label := tabLabels[t]
		switch {
		case t == m.tab:
			parts = append(parts, tabActiveStyle.Render(label))
		case !m.tabAvailable(t):
			parts = append(parts, tabDisabledStyle.Render(label))
		default:
			parts = append(parts, tabInactiveStyle.Render(label))
		}
	}
	tabs := strings.Join(parts, " ")

	taskItems := m.tasks.Items
	if m.mode == ModeViewingArchive {
		if ch := m.currentArchive(); ch != nil && ch.Tasks.Present {
			taskItems = openspec.ParseTasks(ch.Tasks.Content)
		} else {
			taskItems = nil
		}
	}
	total, done := 0, 0
	for _, item := range taskItems {
		if item.Kind == openspec.KindTask {
			total++
			if item.Done {
				done++
			}
		}
	}
	if total > 0 {
		label := fmt.Sprintf(" %d/%d", done, total)
		barSpace := (m.innerWidth()) - lipgloss.Width(tabs) - 3 - len(label)
		if barSpace >= 3 {
			tabs = tabs + " [" + renderProgressBar(done, total, barSpace, "█", "░") + "]" + helpStyle.Render(label)
		}
	}
	return tabs
}

func (m *Model) renderSpecSubnav() string {
	ch := m.current()
	if ch == nil {
		return ""
	}
	var parts []string
	for i, s := range ch.SpecFiles {
		if i == m.specIdx {
			parts = append(parts, tabActiveStyle.Render(s.Name))
		} else {
			parts = append(parts, tabInactiveStyle.Render(s.Name))
		}
	}
	return strings.Join(parts, " ")
}

func (m *Model) hasSpecSubnav() bool {
	ch := m.current()
	return m.tab == TabSpecs && ch != nil && len(ch.SpecFiles) > 0
}

func (m *Model) boxTop() string {
	return separatorStyle.Render("┌" + strings.Repeat("─", m.innerWidth()) + "┐")
}

func (m *Model) boxBottom() string {
	return separatorStyle.Render("└" + strings.Repeat("─", m.innerWidth()) + "┘")
}

func (m *Model) boxInnerSep() string {
	return separatorStyle.Render("├" + strings.Repeat("─", m.innerWidth()) + "┤")
}

func (m *Model) addBorderSides(content string) string {
	lines := strings.Split(content, "\n")
	for len(lines) > 0 && lines[len(lines)-1] == "" {
		lines = lines[:len(lines)-1]
	}
	inner := m.innerWidth()
	result := make([]string, 0, len(lines))
	for _, line := range lines {
		pad := inner - lipgloss.Width(line)
		if pad < 0 {
			pad = 0
		}
		result = append(result, separatorStyle.Render("│")+line+strings.Repeat(" ", pad)+separatorStyle.Render("│"))
	}
	return strings.Join(result, "\n")
}

func (m *Model) renderHelpBar() string {
	if m.errMsg != "" {
		return errStyle.Render(m.errMsg)
	}
	if m.mode == ModeIndex {
		if m.index.FilterActive {
			return helpStyle.Render("/" + m.index.FilterText + "█")
		}
		sortHint := "s: sort by suffix"
		if m.index.SortBySuffix {
			sortHint = "s: sort by name"
		}
		text := "j/k: navigate  Enter: open  Space: expand  click: select  " + sortHint + "  w: worktrees  i: info  ?: help  Esc: quit"
		if m.index.FilterText != "" {
			text += "  [/" + m.index.FilterText + "]"
		}
		return helpStyle.Render(text)
	}
	if m.mode == ModeWorktrees {
		return helpStyle.Render("j/k: navigate  Enter: open (read-only)  ?: help  Esc: index  q: quit")
	}
	if m.mode == ModeViewingConfig {
		return helpStyle.Render("j/k: scroll  i/Esc: back  ?: help  q: quit")
	}
	if m.mode == ModeViewingSpec {
		if m.specViewer.FocusMode {
			return helpStyle.Render("h/l: prev/next req  j/k: scroll  e: edit  Esc: index  ?: help  q: quit")
		}
		return helpStyle.Render("j/k: scroll  e: edit  Esc: index  ?: help  q: quit")
	}
	if m.mode == ModeViewingArchive {
		if m.viewingWorktreeChange {
			return helpStyle.Render("1-4/Tab: artifact  j/k: scroll  e: edit  ?: help  Esc: worktrees  q: quit")
		}
		return helpStyle.Render("1-4/Tab: artifact  j/k: scroll  a/Esc: index  ?: help  q: quit")
	}
	if m.tab == TabTasks {
		return helpStyle.Render("h/l: change  1-4/Tab/←→: artifact  j/k: navigate  Space: toggle  e: edit  i: info  ?: help  Esc: index  q: quit")
	}
	return helpStyle.Render("h/l: change  1-4/Tab/←→: artifact  j/k: scroll  e: edit  i: info  ?: help  Esc: index  q: quit")
}
