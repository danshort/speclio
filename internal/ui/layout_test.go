package ui

import (
	"strings"
	"testing"

	"charm.land/bubbles/v2/viewport"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/danshort/lectern/internal/openspec"
)

// sized drives the WindowSizeMsg path so the model is laid out exactly as the
// production code would size it (vp height = contentHeight, vpReady set).
func sized(m Model, w, h int) Model {
	res, _ := m.Update(tea.WindowSizeMsg{Width: w, Height: h})
	return res.(Model)
}

// TestViewHeightInvariant is the seatbelt for coupling #1: the total rendered
// height must equal the terminal height in every viewport-backed mode, so a
// future change to the chrome rows or to contentHeight() can't silently clip a
// row or leave a blank gap.
func TestViewHeightInvariant(t *testing.T) {
	change := openspec.Change{
		Name:     "feat",
		Proposal: openspec.Artifact{Present: true, Content: "# p"},
		Design:   openspec.Artifact{Present: true, Content: "# d"},
		Specs:    openspec.Artifact{Present: true, Content: "# s"},
		Tasks:    openspec.Artifact{Present: true, Content: "- [ ] t"},
	}

	modes := map[string]func() Model{
		"normal": func() Model {
			m := Model{mode: ModeNormal, project: &openspec.Project{Changes: []openspec.Change{change}}, renderCache: map[Tab]string{}}
			m.tab = TabProposal
			return m
		},
		"archive": func() Model {
			m := Model{mode: ModeViewingArchive, project: &openspec.Project{}, renderCache: map[Tab]string{}}
			m.index.ArchiveChanges = []openspec.Change{change}
			m.tab = TabProposal
			return m
		},
		"index": func() Model {
			m := Model{mode: ModeIndex, project: &openspec.Project{Changes: []openspec.Change{change}}}
			m.index.ExpandedSpecs = map[int]bool{}
			m.index.ExpandedArchives = map[int]bool{}
			m.buildIndexItems()
			return m
		},
		"spec": func() Model {
			m := Model{mode: ModeViewingSpec, project: &openspec.Project{}}
			m.projectSpecs = []openspec.ProjectSpec{{Name: "s", Content: "## Purpose\n"}}
			return m
		},
		"config": func() Model {
			return Model{mode: ModeViewingConfig, project: &openspec.Project{}, projectConfig: openspec.ProjectConfig{Context: "ctx"}}
		},
		"worktrees": func() Model {
			return Model{mode: ModeWorktrees, project: &openspec.Project{}, renderCache: map[Tab]string{}}
		},
	}

	sizes := []struct{ w, h int }{{80, 24}, {120, 40}, {40, 20}}
	for name, mk := range modes {
		for _, s := range sizes {
			m := sized(mk(), s.w, s.h)
			if got := lipgloss.Height(m.View().Content); got != s.h {
				t.Errorf("%s @ %dx%d: rendered height %d, want %d", name, s.w, s.h, got, s.h)
			}
		}
	}

	t.Run("empty-project welcome view is exempt", func(t *testing.T) {
		m := sized(Model{mode: ModeNormal, project: &openspec.Project{}}, 80, 24)
		// The welcome view renders fixed content without a sized viewport, so
		// the invariant intentionally does NOT apply; just confirm it renders.
		if lipgloss.Height(m.View().Content) == 0 {
			t.Error("expected welcome view to render content")
		}
	})

	t.Run("below minimum height the viewport clamps", func(t *testing.T) {
		// At a height smaller than the chrome row count, contentHeight clamps to
		// 1 and the rendered content necessarily exceeds the terminal height.
		m := sized(modes["normal"](), 80, 3)
		if got := lipgloss.Height(m.View().Content); got <= 3 {
			t.Errorf("expected clamped content to exceed tiny terminal height, got %d", got)
		}
	})
}

// assertIndexRoundTrip drives the production render path, then verifies that
// indexItemAtContentLine resolves each visible item exactly once and no hidden
// item at all — render position and hit-test agree. Written against the public
// surface (renderIndexContent / indexItemAtContentLine) so it is unchanged by
// PR4, which only changes how indexItemAtContentLine is computed.
func assertIndexRoundTrip(t *testing.T, m *Model) {
	t.Helper()
	m.buildIndexItems()
	m.applyFilter()
	m.refreshIndexViewport()

	content, _ := m.renderIndexContent()
	n := lipgloss.Height(content)
	seen := map[int]int{}
	for line := 0; line < n; line++ {
		if idx, ok := m.indexItemAtContentLine(line); ok {
			if idx < 0 || idx >= len(m.index.Items) {
				t.Fatalf("resolved out-of-range item idx %d (have %d items)", idx, len(m.index.Items))
			}
			seen[idx]++
		}
	}
	for i := range m.index.Items {
		vis := m.isItemVisible(i)
		switch {
		case vis && seen[i] != 1:
			t.Errorf("visible item %d (kind %d) resolved %d times, want 1", i, m.index.Items[i].kind, seen[i])
		case !vis && seen[i] != 0:
			t.Errorf("hidden item %d resolved %d times, want 0", i, seen[i])
		}
	}
}

// TestIndexHitTestRoundTrip is the seatbelt for coupling #3.
func TestIndexHitTestRoundTrip(t *testing.T) {
	base := func() Model {
		m := Model{
			mode: ModeIndex, width: 80, height: 24, vpReady: true,
			project: &openspec.Project{Changes: []openspec.Change{{Name: "data-export"}, {Name: "user-auth"}}},
			projectSpecs: []openspec.ProjectSpec{
				{Name: "mouse-nav", RequirementCount: 2, RequirementNames: []string{"Wheel", "Click"}, Content: "## Purpose\n"},
				{Name: "tui-viewer", RequirementCount: 1, RequirementNames: []string{"Layout"}, Content: "## Purpose\n"},
			},
		}
		m.index.ExpandedSpecs = map[int]bool{}
		m.index.ExpandedArchives = map[int]bool{}
		m.index.ArchiveChanges = []openspec.Change{{Name: "old-feat", Proposal: openspec.Artifact{Present: true}, Tasks: openspec.Artifact{Present: true}}}
		m.vp = viewport.New(viewport.WithWidth(80), viewport.WithHeight(20))
		return m
	}

	t.Run("empty", func(t *testing.T) {
		m := Model{mode: ModeIndex, width: 80, height: 24, vpReady: true, project: &openspec.Project{}}
		m.index.ExpandedSpecs = map[int]bool{}
		m.index.ExpandedArchives = map[int]bool{}
		m.vp = viewport.New(viewport.WithWidth(80), viewport.WithHeight(20))
		assertIndexRoundTrip(t, &m)
	})
	t.Run("active+specs+archives", func(t *testing.T) { m := base(); assertIndexRoundTrip(t, &m) })
	t.Run("expanded-specs", func(t *testing.T) {
		m := base()
		m.index.ExpandedSpecs[0] = true
		assertIndexRoundTrip(t, &m)
	})
	t.Run("expanded-archives", func(t *testing.T) {
		m := base()
		m.index.ExpandedArchives[0] = true
		assertIndexRoundTrip(t, &m)
	})
	t.Run("active-filter", func(t *testing.T) {
		m := base()
		m.index.FilterText = "data"
		assertIndexRoundTrip(t, &m)
	})
	t.Run("no-match", func(t *testing.T) {
		m := base()
		m.index.FilterText = "zzzznomatch"
		assertIndexRoundTrip(t, &m)
	})
	t.Run("sorted-by-suffix", func(t *testing.T) {
		m := base()
		m.index.SortBySuffix = true
		assertIndexRoundTrip(t, &m)
	})
}

// TestTabHitTestRoundTrip is the seatbelt for coupling #2: a click on a tab's
// rendered position selects that tab. The expected x is derived from the
// rendered tab bar (ANSI-stripped), not from the handler's own geometry, so the
// test is an independent oracle that PR3 must keep satisfying.
func TestTabHitTestRoundTrip(t *testing.T) {
	m := Model{
		mode: ModeNormal, width: 80, height: 24, vpReady: true,
		project: &openspec.Project{Changes: []openspec.Change{{
			Name:     "feat",
			Proposal: openspec.Artifact{Present: true},
			Design:   openspec.Artifact{Present: true},
			Specs:    openspec.Artifact{Present: true},
			Tasks:    openspec.Artifact{Present: true},
		}}},
		renderCache: map[Tab]string{},
	}
	m.tab = TabProposal
	m.vp = viewport.New(viewport.WithWidth(78), viewport.WithHeight(18))

	plain := ansiRe.ReplaceAllString(m.renderTabBar(), "")
	const tabBarRow = 2 // boxTop=0, header=1, tabBar=2
	for tab := Tab(0); tab < tabCount; tab++ {
		col := strings.Index(plain, tabLabels[tab])
		if col < 0 {
			t.Fatalf("label %q not found in rendered tab bar %q", tabLabels[tab], plain)
		}
		screenX := 1 + col // +1 for the left │ border column
		res, _ := m.handleMouseClick(tea.MouseClickMsg{Button: tea.MouseLeft, X: screenX, Y: tabBarRow})
		if got := res.(Model).tab; got != tab {
			t.Errorf("click on %q (x=%d) selected tab %d, want %d", tabLabels[tab], screenX, got, tab)
		}
	}
}
