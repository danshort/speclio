package ui

import (
	"strings"
	"testing"

	"charm.land/bubbles/v2/viewport"
	tea "charm.land/bubbletea/v2"
	"github.com/danshort/lectern/internal/openspec"
)

// TestHelpOverlayOpensFromEachMode verifies `?` opens the overlay from every
// mode without changing the underlying mode (task 4.1).
func TestHelpOverlayOpensFromEachMode(t *testing.T) {
	modes := []Mode{ModeNormal, ModeIndex, ModeViewingArchive, ModeViewingSpec, ModeViewingConfig}
	for _, mode := range modes {
		m := Model{mode: mode}
		result, cmd := m.dispatchKey(tea.KeyPressMsg{Text: "?"})
		updated := result.(Model)
		if !updated.helpOpen {
			t.Errorf("mode %d: expected helpOpen true after ?", mode)
		}
		if updated.mode != mode {
			t.Errorf("mode %d: expected mode unchanged, got %d", mode, updated.mode)
		}
		if cmd != nil {
			t.Errorf("mode %d: expected nil cmd when opening overlay", mode)
		}
	}
}

// TestHelpOverlayDismissKeys verifies `?`, `Esc`, and `q` each close the
// overlay, restore the originating mode, and do not quit (task 4.2).
func TestHelpOverlayDismissKeys(t *testing.T) {
	cases := []struct {
		name string
		msg  tea.KeyPressMsg
	}{
		{"question mark", tea.KeyPressMsg{Text: "?"}},
		{"esc", tea.KeyPressMsg{Code: tea.KeyEsc}},
		{"q", tea.KeyPressMsg{Text: "q"}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			m := Model{mode: ModeIndex, helpOpen: true}
			result, cmd := m.dispatchKey(tc.msg)
			updated := result.(Model)
			if updated.helpOpen {
				t.Error("expected helpOpen false after dismiss")
			}
			if updated.mode != ModeIndex {
				t.Errorf("expected mode restored to ModeIndex, got %d", updated.mode)
			}
			if cmd != nil {
				t.Error("expected nil cmd (dismiss must not quit)")
			}
		})
	}
}

// TestHelpOverlaySwallowsOtherKeys verifies a non-dismiss key is inert while
// the overlay is open (task 4.3).
func TestHelpOverlaySwallowsOtherKeys(t *testing.T) {
	m := Model{mode: ModeIndex, helpOpen: true, index: indexState{Cursor: 2}}
	result, cmd := m.dispatchKey(tea.KeyPressMsg{Text: "j"})
	updated := result.(Model)
	if !updated.helpOpen {
		t.Error("expected overlay to stay open after a non-dismiss key")
	}
	if updated.index.Cursor != 2 {
		t.Errorf("expected cursor untouched, got %d", updated.index.Cursor)
	}
	if cmd != nil {
		t.Error("expected nil cmd while overlay swallows keys")
	}
}

// TestHelpOverlayFilterPrecedence verifies `?` types into an active filter
// instead of opening the overlay (task 4.4).
func TestHelpOverlayFilterPrecedence(t *testing.T) {
	m := &Model{
		width:   80,
		project: &openspec.Project{},
		index:   indexState{FilterActive: true, FilterText: ""},
	}
	m.vp = viewport.New(viewport.WithWidth(80), viewport.WithHeight(20))
	m.vpReady = true
	m.mode = ModeIndex

	result, _ := m.dispatchKey(tea.KeyPressMsg{Text: "?"})
	updated := result.(Model)
	if updated.helpOpen {
		t.Error("expected overlay NOT to open while filter input is active")
	}
	if updated.index.FilterText != "?" {
		t.Errorf("expected '?' appended to filter text, got %q", updated.index.FilterText)
	}
}

// TestHelpOverlayContent verifies the rendered overlay lists each screen group
// and representative shortcuts (task 4.5).
func TestHelpOverlayContent(t *testing.T) {
	m := &Model{width: 100, height: 40}
	out := m.renderHelpOverlay()

	wants := []string{
		// group headings
		"Global", "Index", "Change viewer", "Archive viewer", "Spec viewer", "Config viewer",
		// representative shortcuts
		"j/k", "Enter", "Space", "1-4", "h / l", "q / Ctrl+C",
		// spec sub-navigation in the change/archive viewers
		"[ / ]",
		// arrow aliases must be documented (j/k↔↑↓ and Tab↔←→)
		"↑↓", "Tab/Shift+Tab/←→",
	}
	for _, w := range wants {
		if !strings.Contains(out, w) {
			t.Errorf("expected overlay to contain %q", w)
		}
	}
}

// TestHelpBarAffordance verifies the help bar advertises `?: help` in a regular
// state and omits it while the index filter input is active (task 4.6).
func TestHelpBarAffordance(t *testing.T) {
	t.Run("regular index state shows ?: help", func(t *testing.T) {
		m := &Model{mode: ModeIndex}
		if out := m.renderHelpBar(); !strings.Contains(out, "?: help") {
			t.Errorf("expected help bar to contain '?: help', got %q", out)
		}
	})

	t.Run("active filter hides ?: help", func(t *testing.T) {
		m := &Model{mode: ModeIndex, index: indexState{FilterActive: true, FilterText: "x"}}
		if out := m.renderHelpBar(); strings.Contains(out, "?: help") {
			t.Errorf("expected help bar to omit '?: help' during filter input, got %q", out)
		}
	})
}
