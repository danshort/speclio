package ui

import (
	"errors"
	"strings"
	"testing"

	"charm.land/bubbles/v2/viewport"
	"github.com/danshort/lectern/internal/openspec"
)

// An unreadable spec opened in the viewport must not show the structural
// "Validation errors" banner — a read failure is not a structural one.
func TestUnreadableSpecViewportSkipsValidation(t *testing.T) {
	mk := func(content string, readErr error) Model {
		m := Model{mode: ModeViewingSpec, width: 80, height: 24, vpReady: true, project: &openspec.Project{}}
		m.projectSpecs = []openspec.ProjectSpec{{Name: "s", Content: content, ReadErr: readErr}}
		m.vp = viewport.New(viewport.WithWidth(78), viewport.WithHeight(18))
		return m
	}
	run := func(m Model) string {
		cmd := m.loadViewportForSpec()
		if cmd == nil {
			return ""
		}
		if sr, ok := cmd().(specRenderedMsg); ok {
			return sr.content
		}
		return ""
	}

	// Control: a readable but structurally-invalid spec shows the banner.
	if got := run(mk("not a real spec", nil)); !strings.Contains(got, "Validation errors") {
		t.Fatalf("control: expected validation banner for an invalid readable spec, got %q", got)
	}
	// Unreadable: ReadErr set ⇒ no banner.
	if got := run(mk("⚠ couldn't read .../spec.md: boom", errors.New("boom"))); strings.Contains(got, "Validation errors") {
		t.Errorf("unreadable spec should skip the validation banner, got %q", got)
	}
}

func newIndexModel() Model {
	m := Model{mode: ModeIndex, width: 80, project: &openspec.Project{}}
	m.index.ExpandedSpecs = map[int]bool{}
	m.index.ExpandedArchives = map[int]bool{}
	return m
}

func TestUnreadableSpecMarker(t *testing.T) {
	m := newIndexModel()
	m.projectSpecs = []openspec.ProjectSpec{
		{Name: "unreadable-spec", ReadErr: errors.New("permission denied"), Content: "⚠ couldn't read .../spec.md: permission denied"},
		{Name: "valid-spec", Content: "## Purpose\nP\n\n## Requirements\n\n### Requirement: R\n#### Scenario: S\n- **WHEN** a\n- **THEN** b\n"},
	}
	m.buildIndexItems()

	out, _ := m.renderIndexContent()
	if !strings.Contains(out, "⚠") {
		t.Error("expected ⚠ marker for the unreadable spec")
	}
	// The unreadable spec must NOT also be flagged ✗ (read failure ≠ invalid),
	// and the valid spec is fine — so no ✗ anywhere.
	if strings.Contains(out, "✗") {
		t.Error("unreadable spec should show ⚠ in place of ✗, and the valid spec none")
	}
}

func TestUnreadableChangeMarker(t *testing.T) {
	t.Run("unreadable artifact shows warn marker", func(t *testing.T) {
		m := newIndexModel()
		m.project.Changes = []openspec.Change{
			{Name: "feat", Proposal: openspec.Artifact{Present: true, ReadErr: errors.New("EIO")}},
		}
		m.buildIndexItems()
		out, _ := m.renderIndexContent()
		if !strings.Contains(out, "⚠") {
			t.Error("expected ⚠ for a change with an unreadable artifact")
		}
		if strings.Contains(out, "✗") {
			t.Error("unreadable artifact should not also produce a ✗")
		}
	})

	t.Run("genuinely missing proposal still shows validation cross", func(t *testing.T) {
		m := newIndexModel()
		m.project.Changes = []openspec.Change{{Name: "feat"}} // no proposal present
		m.buildIndexItems()
		out, _ := m.renderIndexContent()
		if !strings.Contains(out, "✗") {
			t.Error("a change missing its proposal should still show ✗")
		}
	})
}
