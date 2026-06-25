package ui

import (
	"strings"
	"testing"

	"charm.land/bubbles/v2/viewport"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/danshort/lectern/internal/openspec"
)

// specNavModel builds a viewer model parked on the specs tab with the given
// spec names, in the given mode (ModeNormal or ModeViewingArchive). It mirrors
// the production sizing (vp height = contentHeight) so chrome-row hit-testing
// resolves the same screen rows the renderer lays out.
func specNavModel(mode Mode, names []string) Model {
	specs := make([]openspec.NamedSpec, len(names))
	for i, n := range names {
		specs[i] = openspec.NamedSpec{Name: n, Content: "## " + n + "\n"}
	}
	ch := openspec.Change{
		Name:      "feat",
		Proposal:  openspec.Artifact{Present: true},
		Specs:     openspec.Artifact{Present: true},
		SpecFiles: specs,
	}
	m := Model{mode: mode, width: 80, height: 24, vpReady: true, renderCache: map[Tab]string{}}
	if mode == ModeViewingArchive {
		m.project = &openspec.Project{}
		m.index.ArchiveChanges = []openspec.Change{ch}
	} else {
		m.project = &openspec.Project{Changes: []openspec.Change{ch}}
	}
	m.tab = TabSpecs
	m.vp = viewport.New(viewport.WithWidth(78), viewport.WithHeight(18))
	m.vp.SetHeight(m.contentHeight())
	return m
}

// TestSpecRangesMatchRenderedWidths pins the spec-chip range arithmetic: spans
// start at column 1, match each chip's rendered (lipgloss.Width) span, and are
// separated by the single-space join — the spec analogue of
// TestTabRangesMatchRenderedWidths.
func TestSpecRangesMatchRenderedWidths(t *testing.T) {
	m := specNavModel(ModeNormal, []string{"alpha", "beta", "gamma"})
	m.specIdx = 1 // active chip must not change widths (styles share Padding)

	ranges := m.specRanges()
	if len(ranges) != 3 {
		t.Fatalf("expected 3 spec ranges, got %d", len(ranges))
	}
	ch := m.current()
	wantStart := 1
	for i := range ch.SpecFiles {
		w := lipgloss.Width(m.styledSpec(ch, i))
		if ranges[i].start != wantStart || ranges[i].end != wantStart+w-1 {
			t.Errorf("spec %d range = {%d,%d}, want {%d,%d}", i, ranges[i].start, ranges[i].end, wantStart, wantStart+w-1)
		}
		wantStart = ranges[i].end + 2 // end + 1 (last col) + 1 (single-space join)
	}
}

// TestSpecChipHitTestRoundTrip is the seatbelt coupling the rendered chip row to
// the click handler: a click on a chip's rendered position selects that spec.
// The expected X is derived from the rendered (ANSI-stripped) sub-nav, not the
// handler's own geometry, so it is an independent oracle.
func TestSpecChipHitTestRoundTrip(t *testing.T) {
	for _, mode := range []struct {
		name string
		mode Mode
	}{{"normal", ModeNormal}, {"archive", ModeViewingArchive}} {
		t.Run(mode.name, func(t *testing.T) {
			names := []string{"alpha", "beta", "gamma"}
			m := specNavModel(mode.mode, names)
			subRow := m.chromeRowIndex(rowSubnav)
			if subRow < 0 {
				t.Fatal("expected the spec sub-nav row to be present")
			}
			plain := ansiRe.ReplaceAllString(m.renderSpecSubnav(), "")
			for i, name := range names {
				col := strings.Index(plain, name)
				if col < 0 {
					t.Fatalf("chip %q not found in rendered sub-nav %q", name, plain)
				}
				screenX := 1 + col // +1 for the left │ border column
				res, _ := m.handleMouseClick(tea.MouseClickMsg{Button: tea.MouseLeft, X: screenX, Y: subRow})
				if got := res.(Model).specIdx; got != i {
					t.Errorf("click on %q (x=%d) selected spec %d, want %d", name, screenX, got, i)
				}
			}
		})
	}
}

// TestSpecChipClickIgnoredOffChip verifies clicks that miss every chip (between
// chips or past the row) leave the selection unchanged.
func TestSpecChipClickIgnoredOffChip(t *testing.T) {
	m := specNavModel(ModeNormal, []string{"alpha", "beta"})
	m.specIdx = 0
	subRow := m.chromeRowIndex(rowSubnav)
	ranges := m.specRanges()
	gap := ranges[0].end + 1 // the single-space join between chips

	res, _ := m.handleMouseClick(tea.MouseClickMsg{Button: tea.MouseLeft, X: gap, Y: subRow})
	if got := res.(Model).specIdx; got != 0 {
		t.Errorf("click in the gap changed specIdx to %d, want 0", got)
	}
	res, _ = m.handleMouseClick(tea.MouseClickMsg{Button: tea.MouseLeft, X: ranges[len(ranges)-1].end + 50, Y: subRow})
	if got := res.(Model).specIdx; got != 0 {
		t.Errorf("click past the chips changed specIdx to %d, want 0", got)
	}
}

func TestSpecKeyNavigation(t *testing.T) {
	press := func(m Model, s string) Model {
		res, _ := m.dispatchKey(tea.KeyPressMsg{Text: s})
		return res.(Model)
	}

	t.Run("] advances to the next spec", func(t *testing.T) {
		m := specNavModel(ModeNormal, []string{"a", "b", "c"})
		m = press(m, "]")
		if m.specIdx != 1 {
			t.Errorf("after ] specIdx = %d, want 1", m.specIdx)
		}
	})

	t.Run("[ goes to the previous spec", func(t *testing.T) {
		m := specNavModel(ModeNormal, []string{"a", "b", "c"})
		m.specIdx = 2
		m = press(m, "[")
		if m.specIdx != 1 {
			t.Errorf("after [ specIdx = %d, want 1", m.specIdx)
		}
	})

	t.Run("] wraps from last to first", func(t *testing.T) {
		m := specNavModel(ModeNormal, []string{"a", "b", "c"})
		m.specIdx = 2
		m = press(m, "]")
		if m.specIdx != 0 {
			t.Errorf("after ] from last specIdx = %d, want 0", m.specIdx)
		}
	})

	t.Run("[ wraps from first to last", func(t *testing.T) {
		m := specNavModel(ModeNormal, []string{"a", "b", "c"})
		m = press(m, "[")
		if m.specIdx != 2 {
			t.Errorf("after [ from first specIdx = %d, want 2", m.specIdx)
		}
	})

	t.Run("single spec is a no-op", func(t *testing.T) {
		m := specNavModel(ModeNormal, []string{"only"})
		if m = press(m, "]"); m.specIdx != 0 {
			t.Errorf("] with one spec changed specIdx to %d", m.specIdx)
		}
		if m = press(m, "["); m.specIdx != 0 {
			t.Errorf("[ with one spec changed specIdx to %d", m.specIdx)
		}
	})

	t.Run("arrows do not change the spec", func(t *testing.T) {
		m := specNavModel(ModeNormal, []string{"a", "b", "c"})
		m.specIdx = 1
		res, _ := m.dispatchKey(tea.KeyPressMsg{Code: tea.KeyRight})
		if got := res.(Model).specIdx; got != 1 {
			t.Errorf("→ changed specIdx to %d, want 1", got)
		}
		res, _ = m.dispatchKey(tea.KeyPressMsg{Code: tea.KeyLeft})
		if got := res.(Model).specIdx; got != 1 {
			t.Errorf("← changed specIdx to %d, want 1", got)
		}
	})

	t.Run("3 selects the specs tab without cycling", func(t *testing.T) {
		m := specNavModel(ModeNormal, []string{"a", "b", "c"})
		m.specIdx = 1
		m = press(m, "3")
		if m.tab != TabSpecs {
			t.Errorf("3 left tab = %d, want TabSpecs", m.tab)
		}
		if m.specIdx != 1 {
			t.Errorf("3 cycled specIdx to %d, want 1 (no cycle)", m.specIdx)
		}
	})

	t.Run("selected spec is preserved across a tab round-trip", func(t *testing.T) {
		m := specNavModel(ModeNormal, []string{"a", "b", "c"})
		m.specIdx = 2
		m = press(m, "1") // proposal
		if m.tab != TabProposal {
			t.Fatalf("1 left tab = %d, want TabProposal", m.tab)
		}
		m = press(m, "3") // back to specs
		if m.specIdx != 2 {
			t.Errorf("returning to specs reset specIdx to %d, want 2 (preserved)", m.specIdx)
		}
	})

	t.Run("changing change resets the selected spec", func(t *testing.T) {
		mk := func(name string) openspec.Change {
			return openspec.Change{
				Name: name, Proposal: openspec.Artifact{Present: true}, Specs: openspec.Artifact{Present: true},
				SpecFiles: []openspec.NamedSpec{{Name: "x"}, {Name: "y"}, {Name: "z"}},
			}
		}
		m := Model{mode: ModeNormal, width: 80, height: 24, vpReady: true, renderCache: map[Tab]string{},
			project: &openspec.Project{Changes: []openspec.Change{mk("one"), mk("two")}}}
		m.tab = TabSpecs
		m.specIdx = 2
		m.vp = viewport.New(viewport.WithWidth(78), viewport.WithHeight(18))
		m = press(m, "l") // next change
		if m.changeIdx != 1 {
			t.Fatalf("l left changeIdx = %d, want 1", m.changeIdx)
		}
		if m.specIdx != 0 {
			t.Errorf("switching change left specIdx = %d, want 0", m.specIdx)
		}
	})
}

// TestActivateIndexResetsSpecIdx pins that opening a change/archive from the
// index resets the selected spec — a flow that does not pass through the h/l
// reset, so a stale specIdx from a previously-viewed change must not leak in.
func TestActivateIndexResetsSpecIdx(t *testing.T) {
	mk := func(name string) openspec.Change {
		return openspec.Change{
			Name: name, Proposal: openspec.Artifact{Present: true}, Specs: openspec.Artifact{Present: true},
			SpecFiles: []openspec.NamedSpec{{Name: "x"}, {Name: "y"}, {Name: "z"}},
		}
	}
	cases := []struct {
		name string
		item indexItem
		mk   func(*Model)
	}{
		{"active change", indexItem{kind: indexKindActive, idx: 1}, func(m *Model) {
			m.project = &openspec.Project{Changes: []openspec.Change{mk("one"), mk("two")}}
		}},
		{"archived change", indexItem{kind: indexKindArchived, idx: 0}, func(m *Model) {
			m.project = &openspec.Project{}
			m.index.ArchiveChanges = []openspec.Change{mk("old")}
		}},
		{"archived artifact (specs tab)", indexItem{kind: indexKindArchivedArtifact, idx: 0, reqIdx: int(TabSpecs)}, func(m *Model) {
			m.project = &openspec.Project{}
			m.index.ArchiveChanges = []openspec.Change{mk("old")}
		}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			m := Model{mode: ModeIndex, width: 80, height: 24, vpReady: true, renderCache: map[Tab]string{}}
			tc.mk(&m)
			m.vp = viewport.New(viewport.WithWidth(78), viewport.WithHeight(18))
			m.specIdx = 2 // stale selection from a previously-viewed change
			res, _ := m.activateIndexItem(tc.item)
			if got := res.(Model).specIdx; got != 0 {
				t.Errorf("opening from index left specIdx = %d, want 0 (reset)", got)
			}
		})
	}
}

// TestSpecHelpBarHint covers the tui-viewer "Barra de ayuda de teclado"
// scenarios: the specs tab advertises [ / ] when the change has multiple specs
// (in ModeNormal and ModeViewingArchive) and omits it for a single spec, and the
// removed 3-cycle is never advertised.
func TestSpecHelpBarHint(t *testing.T) {
	t.Run("multiple specs advertises [/] in normal mode", func(t *testing.T) {
		m := specNavModel(ModeNormal, []string{"a", "b"})
		if out := m.renderHelpBar(); !strings.Contains(out, "[/]: spec") {
			t.Errorf("expected specs help bar to advertise '[/]: spec', got %q", out)
		}
	})
	t.Run("multiple specs advertises [/] in archive mode", func(t *testing.T) {
		m := specNavModel(ModeViewingArchive, []string{"a", "b"})
		if out := m.renderHelpBar(); !strings.Contains(out, "[/]: spec") {
			t.Errorf("expected archive specs help bar to advertise '[/]: spec', got %q", out)
		}
	})
	t.Run("single spec omits the [/] hint", func(t *testing.T) {
		m := specNavModel(ModeNormal, []string{"only"})
		if out := m.renderHelpBar(); strings.Contains(out, "[/]") {
			t.Errorf("expected single-spec help bar to omit '[/]', got %q", out)
		}
	})
	t.Run("help bar never advertises the removed 3-cycle", func(t *testing.T) {
		m := specNavModel(ModeNormal, []string{"a", "b"})
		if out := m.renderHelpBar(); strings.Contains(strings.ToLower(out), "cycle") {
			t.Errorf("help bar should not mention cycling, got %q", out)
		}
	})
}

// TestSpecNavWorktreeChange covers the third entry into ModeViewingArchive — a
// foreign-worktree change opened read-only — to ensure [ / ] works there too.
func TestSpecNavWorktreeChange(t *testing.T) {
	m := Model{mode: ModeViewingArchive, width: 80, height: 24, vpReady: true, renderCache: map[Tab]string{},
		project: &openspec.Project{}, viewingWorktreeChange: true,
		worktreeViewChange: openspec.Change{
			Name: "wt", Specs: openspec.Artifact{Present: true},
			SpecFiles: []openspec.NamedSpec{{Name: "a"}, {Name: "b"}},
		}}
	m.tab = TabSpecs
	m.vp = viewport.New(viewport.WithWidth(78), viewport.WithHeight(18))
	res, _ := m.dispatchKey(tea.KeyPressMsg{Text: "]"})
	if got := res.(Model).specIdx; got != 1 {
		t.Errorf("] in worktree-change view specIdx = %d, want 1", got)
	}
}
