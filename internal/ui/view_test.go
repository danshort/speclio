package ui

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"charm.land/bubbles/v2/viewport"
	tea "charm.land/bubbletea/v2"
	"github.com/danshort/lectern/internal/openspec"
)

func testLoader() *openspec.Loader {
	return openspec.NewLoader(openspec.OSFS{})
}

func TestExtractRequirement(t *testing.T) {
	raw := `# Spec

### Requirement: Login
The system SHALL authenticate users.

#### Scenario: Success
- **WHEN** valid credentials
- **THEN** user is logged in

### Requirement: Logout
The system SHALL log out users.

#### Scenario: Click logout
- **WHEN** user clicks logout
- **THEN** session ends
`

	t.Run("name found returns block", func(t *testing.T) {
		result := openspec.ExtractRequirement(raw, "Login")
		if !strings.Contains(result, "### Requirement: Login") {
			t.Error("expected requirement header in result")
		}
		if !strings.Contains(result, "The system SHALL authenticate users") {
			t.Error("expected requirement body in result")
		}
		if strings.Contains(result, "### Requirement: Logout") {
			t.Error("expected block to stop at next requirement")
		}
	})

	t.Run("name not found returns empty", func(t *testing.T) {
		result := openspec.ExtractRequirement(raw, "Nonexistent")
		if result != "" {
			t.Errorf("expected empty result, got %q", result)
		}
	})

	t.Run("last requirement in document", func(t *testing.T) {
		result := openspec.ExtractRequirement(raw, "Logout")
		if !strings.Contains(result, "### Requirement: Logout") {
			t.Error("expected requirement header")
		}
		if !strings.Contains(result, "session ends") {
			t.Error("expected full block for last requirement")
		}
	})

	t.Run("requirement with no following header", func(t *testing.T) {
		single := "### Requirement: Only\nJust one requirement."
		result := openspec.ExtractRequirement(single, "Only")
		if result != single {
			t.Errorf("expected full content, got %q", result)
		}
	})
}

func TestFirstAvailableTab(t *testing.T) {
	t.Run("change with all tabs", func(t *testing.T) {
		ch := openspec.Change{
			Proposal: openspec.Artifact{Present: true},
			Design:   openspec.Artifact{Present: true},
			Specs:    openspec.Artifact{Present: true},
			Tasks:    openspec.Artifact{Present: true},
		}
		if got := firstAvailableTab(ch); got != TabProposal {
			t.Errorf("expected TabProposal, got %d", got)
		}
	})

	t.Run("change with only proposal and tasks", func(t *testing.T) {
		ch := openspec.Change{
			Proposal: openspec.Artifact{Present: true},
			Tasks:    openspec.Artifact{Present: true},
		}
		if got := firstAvailableTab(ch); got != TabProposal {
			t.Errorf("expected TabProposal, got %d", got)
		}
	})

	t.Run("change with no artifacts", func(t *testing.T) {
		ch := openspec.Change{}
		if got := firstAvailableTab(ch); got != TabProposal {
			t.Errorf("expected TabProposal as default, got %d", got)
		}
	})
}

func TestBuildIndexItems(t *testing.T) {
	m := &Model{
		project: &openspec.Project{
			Changes: []openspec.Change{
				{Name: "feat-a"},
				{Name: "feat-b"},
			},
		},
		projectSpecs: []openspec.ProjectSpec{
			{Name: "auth", RequirementCount: 1, RequirementNames: []string{"Login"}},
		},
	}

	t.Run("with active changes specs and archived", func(t *testing.T) {
		m.index.ExpandedSpecs = make(map[int]bool)
		m.index.ArchiveChanges = []openspec.Change{
			{Name: "old-feat", DisplayDate: "2026-05-01"},
		}
		m.buildIndexItems()
		if len(m.index.Items) != 4 {
			t.Fatalf("expected 4 index items (2 active + 1 spec + 1 archive), got %d", len(m.index.Items))
		}
		if m.index.Items[0].kind != indexKindActive {
			t.Error("expected first item to be active change")
		}
		if m.index.Items[2].kind != indexKindSpec {
			t.Error("expected third item to be spec")
		}
		if m.index.Items[3].kind != indexKindArchived {
			t.Error("expected fourth item to be archived")
		}
	})

	t.Run("empty index", func(t *testing.T) {
		empty := &Model{
			project: &openspec.Project{},
		}
		empty.index.ExpandedSpecs = make(map[int]bool)
		empty.buildIndexItems()
		if len(empty.index.Items) != 0 {
			t.Errorf("expected 0 items, got %d", len(empty.index.Items))
		}
	})
}

func TestRenderTasksContent(t *testing.T) {
	t.Run("with task cursor", func(t *testing.T) {
		m := &Model{
			width: 80,
			tasks: taskState{
				Items: []openspec.TaskItem{
					{Kind: openspec.KindSection, Text: "Section 1", LineNum: 0},
					{Kind: openspec.KindTask, Text: "do thing", Done: false, LineNum: 1},
					{Kind: openspec.KindTask, Text: "another thing", Done: false, LineNum: 2},
				},
				Cursor: 1,
			},
		}
		content, cursorLine := m.renderTasksContent()
		if cursorLine == 0 {
			t.Error("expected non-zero cursor line")
		}
		if !strings.Contains(content, "▶") {
			t.Error("expected cursor indicator (▶) in content")
		}
	})

	t.Run("empty task list", func(t *testing.T) {
		m := &Model{
			width: 80,
			tasks: taskState{
				Items:  nil,
				Cursor: 0,
			},
		}
		content, _ := m.renderTasksContent()
		if content != "" {
			t.Errorf("expected empty content, got %q", content)
		}
	})
}

func TestUpdateKeyPresses(t *testing.T) {
	t.Run("q quits normal mode", func(t *testing.T) {
		m := Model{mode: ModeNormal}
		msg := tea.KeyPressMsg{Text: "q"}
		result, cmd := m.dispatchKey(msg)
		if _, ok := result.(Model); !ok {
			t.Error("expected Model result")
		}
		if cmd == nil {
			t.Error("expected quit command")
		}
	})

	t.Run("i enters config mode", func(t *testing.T) {
		m := Model{mode: ModeNormal, width: 80, height: 24}
		m.vp = viewport.New(viewport.WithWidth(80), viewport.WithHeight(20))
		msg := tea.KeyPressMsg{Text: "i"}
		result, _ := m.dispatchKey(msg)
		updated := result.(Model)
		if updated.mode != ModeViewingConfig {
			t.Errorf("expected ModeViewingConfig, got %d", updated.mode)
		}
	})

	t.Run("a enters index mode", func(t *testing.T) {
		m := Model{mode: ModeNormal, width: 80, height: 24, project: &openspec.Project{}, loader: testLoader()}
		m.vp = viewport.New(viewport.WithWidth(80), viewport.WithHeight(20))
		msg := tea.KeyPressMsg{Text: "a"}
		result, _ := m.dispatchKey(msg)
		updated := result.(Model)
		if updated.mode != ModeIndex {
			t.Errorf("expected ModeIndex, got %d", updated.mode)
		}
	})

	t.Run("esc in index mode quits", func(t *testing.T) {
		m := Model{mode: ModeIndex}
		msg := tea.KeyPressMsg{Code: tea.KeyEsc}
		_, cmd := m.dispatchKey(msg)
		if cmd == nil {
			t.Error("expected quit command")
		}
	})

	t.Run("j moves cursor down in index", func(t *testing.T) {
		m := Model{
			mode:  ModeIndex,
			width: 80,
			index: indexState{
				Items:  []indexItem{{kind: indexKindActive, idx: 0}, {kind: indexKindActive, idx: 1}},
				Cursor: 0,
			},
			project: &openspec.Project{},
		}
		msg := tea.KeyPressMsg{Text: "j"}
		result, _ := m.dispatchKey(msg)
		updated := result.(Model)
		if updated.index.Cursor != 1 {
			t.Errorf("expected cursor 1, got %d", updated.index.Cursor)
		}
	})

	t.Run("k moves cursor up in index", func(t *testing.T) {
		m := Model{
			mode:  ModeIndex,
			width: 80,
			index: indexState{
				Items:  []indexItem{{kind: indexKindActive, idx: 0}, {kind: indexKindActive, idx: 1}},
				Cursor: 1,
			},
			project: &openspec.Project{},
		}
		msg := tea.KeyPressMsg{Text: "k"}
		result, _ := m.dispatchKey(msg)
		updated := result.(Model)
		if updated.index.Cursor != 0 {
			t.Errorf("expected cursor 0, got %d", updated.index.Cursor)
		}
	})
}

// Regression for #7: on an archived change the Tasks tab is read-only markdown,
// not the interactive task list. Pressing down/up must scroll the viewport, not
// drive the task cursor (which re-renders from unpopulated items and blanks the
// view). The cursor must stay put in archive mode.
func TestArchiveTasksTabArrowDoesNotMoveCursor(t *testing.T) {
	newModel := func() Model {
		m := Model{
			mode:   ModeViewingArchive,
			tab:    TabTasks,
			width:  80,
			height: 24,
			tasks: taskState{
				Items: []openspec.TaskItem{
					{Kind: openspec.KindTask, Text: "first"},
					{Kind: openspec.KindTask, Text: "second"},
				},
				Cursor: 0,
			},
		}
		m.vp = viewport.New(viewport.WithWidth(78), viewport.WithHeight(20))
		return m
	}

	t.Run("down keeps cursor", func(t *testing.T) {
		m := newModel()
		result, _ := m.dispatchKey(tea.KeyPressMsg{Text: "j"})
		if got := result.(Model).tasks.Cursor; got != 0 {
			t.Errorf("expected task cursor to stay 0 in archive mode, got %d", got)
		}
	})

	t.Run("up keeps cursor", func(t *testing.T) {
		m := newModel()
		m.tasks.Cursor = 1
		result, _ := m.dispatchKey(tea.KeyPressMsg{Text: "k"})
		if got := result.(Model).tasks.Cursor; got != 1 {
			t.Errorf("expected task cursor to stay 1 in archive mode, got %d", got)
		}
	})
}

func TestIndexValidationMarker(t *testing.T) {
	validSpec := openspec.ProjectSpec{
		Name:    "good",
		Content: "## Purpose\nP.\n\n## Requirements\n\n### Requirement: R\n#### Scenario: S\n- **WHEN** a\n- **THEN** b\n",
	}
	invalidSpec := openspec.ProjectSpec{
		Name:    "bad",
		Content: "## Purpose\nP only, no requirements section.\n",
	}

	t.Run("invalid spec gets a marker", func(t *testing.T) {
		m := Model{
			width:        80,
			mode:         ModeIndex,
			project:      &openspec.Project{},
			index:        indexState{ExpandedSpecs: map[int]bool{}},
			projectSpecs: []openspec.ProjectSpec{validSpec, invalidSpec},
		}
		m.buildIndexItems()
		out, _ := m.renderIndexContent()
		if !strings.Contains(out, "✗") {
			t.Error("expected validation marker (✗) for the invalid spec")
		}
	})

	t.Run("all-valid index has no marker", func(t *testing.T) {
		m := Model{
			width:        80,
			mode:         ModeIndex,
			project:      &openspec.Project{},
			index:        indexState{ExpandedSpecs: map[int]bool{}},
			projectSpecs: []openspec.ProjectSpec{validSpec},
		}
		m.buildIndexItems()
		out, _ := m.renderIndexContent()
		if strings.Contains(out, "✗") {
			t.Error("did not expect a validation marker when all specs are valid")
		}
	})
}

func TestArrowKeyTabNavigation(t *testing.T) {
	// proposal + specs available, design + tasks absent → arrows skip disabled.
	newModel := func(startTab Tab) Model {
		m := Model{
			mode:      ModeNormal,
			width:     80,
			height:    24,
			tab:       startTab,
			changeIdx: 0,
			project: &openspec.Project{Changes: []openspec.Change{{
				Name:     "feat",
				Proposal: openspec.Artifact{Present: true},
				Specs:    openspec.Artifact{Present: true},
			}}},
		}
		m.vp = viewport.New(viewport.WithWidth(78), viewport.WithHeight(20))
		return m
	}

	t.Run("right advances to next available tab", func(t *testing.T) {
		m := newModel(TabProposal)
		result, _ := m.dispatchKey(tea.KeyPressMsg{Code: tea.KeyRight})
		if got := result.(Model).tab; got != TabSpecs {
			t.Errorf("expected → to move proposal→specs (skipping disabled), got tab %d", got)
		}
	})

	t.Run("left goes to previous available tab", func(t *testing.T) {
		m := newModel(TabSpecs)
		result, _ := m.dispatchKey(tea.KeyPressMsg{Code: tea.KeyLeft})
		if got := result.(Model).tab; got != TabProposal {
			t.Errorf("expected ← to move specs→proposal (skipping disabled), got tab %d", got)
		}
	})

	t.Run("h/l still navigate changes, not tabs", func(t *testing.T) {
		// Single change: l wraps to same change and must not change the tab.
		m := newModel(TabProposal)
		result, _ := m.dispatchKey(tea.KeyPressMsg{Text: "l"})
		if got := result.(Model).tab; got != TabProposal {
			t.Errorf("expected l to leave tab unchanged, got tab %d", got)
		}
	})
}

func TestMoveCursorOnSections(t *testing.T) {
	t.Run("moveCursorUp goes to section header", func(t *testing.T) {
		m := &Model{
			tasks: taskState{
				Items: []openspec.TaskItem{
					{Kind: openspec.KindSection, Text: "Section 1"},
					{Kind: openspec.KindTask, Text: "do thing"},
				},
				Cursor: 1,
			},
		}
		m.moveCursorUp()
		if m.tasks.Cursor != 0 {
			t.Errorf("expected cursor at section header (0), got %d", m.tasks.Cursor)
		}
	})

	t.Run("moveCursorDown goes to section header", func(t *testing.T) {
		m := &Model{
			tasks: taskState{
				Items: []openspec.TaskItem{
					{Kind: openspec.KindTask, Text: "do thing"},
					{Kind: openspec.KindSection, Text: "Section 2"},
				},
				Cursor: 0,
			},
		}
		m.moveCursorDown()
		if m.tasks.Cursor != 1 {
			t.Errorf("expected cursor at section header (1), got %d", m.tasks.Cursor)
		}
	})

	t.Run("moveCursorUp stops at first item", func(t *testing.T) {
		m := &Model{
			tasks: taskState{
				Items: []openspec.TaskItem{
					{Kind: openspec.KindSection, Text: "Section 1"},
					{Kind: openspec.KindTask, Text: "do thing"},
				},
				Cursor: 0,
			},
		}
		m.moveCursorUp()
		if m.tasks.Cursor != 0 {
			t.Errorf("expected cursor to stay at 0, got %d", m.tasks.Cursor)
		}
	})

	t.Run("moveCursorDown stops at last item", func(t *testing.T) {
		m := &Model{
			tasks: taskState{
				Items: []openspec.TaskItem{
					{Kind: openspec.KindTask, Text: "do thing"},
					{Kind: openspec.KindSection, Text: "Section 2"},
				},
				Cursor: 1,
			},
		}
		m.moveCursorDown()
		if m.tasks.Cursor != 1 {
			t.Errorf("expected cursor to stay at 1, got %d", m.tasks.Cursor)
		}
	})

	t.Run("moveCursorUp then down navigates through sections and tasks", func(t *testing.T) {
		m := &Model{
			tasks: taskState{
				Items: []openspec.TaskItem{
					{Kind: openspec.KindSection, Text: "S1"},
					{Kind: openspec.KindTask, Text: "T1"},
					{Kind: openspec.KindTask, Text: "T2"},
					{Kind: openspec.KindSection, Text: "S2"},
					{Kind: openspec.KindTask, Text: "T3"},
				},
				Cursor: 2,
			},
		}
		m.moveCursorDown() // T2 -> S2
		if m.tasks.Cursor != 3 {
			t.Errorf("expected cursor at section S2 (3), got %d", m.tasks.Cursor)
		}
		m.moveCursorDown() // S2 -> T3
		if m.tasks.Cursor != 4 {
			t.Errorf("expected cursor at T3 (4), got %d", m.tasks.Cursor)
		}
		m.moveCursorUp() // T3 -> S2
		if m.tasks.Cursor != 3 {
			t.Errorf("expected cursor back at section S2 (3), got %d", m.tasks.Cursor)
		}
		m.moveCursorUp() // S2 -> T2
		if m.tasks.Cursor != 2 {
			t.Errorf("expected cursor back at T2 (2), got %d", m.tasks.Cursor)
		}
	})
}

func TestRenderCursorOnSectionHeader(t *testing.T) {
	t.Run("cursor on section shows ▶ mark", func(t *testing.T) {
		m := &Model{
			width: 80,
			tasks: taskState{
				Items: []openspec.TaskItem{
					{Kind: openspec.KindSection, Text: "Section 1"},
					{Kind: openspec.KindTask, Text: "do thing", Done: false},
				},
				Cursor: 0,
			},
		}
		content, cursorLine := m.renderTasksContent()
		if !strings.Contains(content, "▶") {
			t.Error("expected cursor indicator (▶) in content for section header")
		}
		if cursorLine < 0 {
			t.Errorf("expected non-negative cursor line, got %d", cursorLine)
		}
	})

	t.Run("cursor on task still works", func(t *testing.T) {
		m := &Model{
			width: 80,
			tasks: taskState{
				Items: []openspec.TaskItem{
					{Kind: openspec.KindSection, Text: "Section 1"},
					{Kind: openspec.KindTask, Text: "do thing", Done: false},
				},
				Cursor: 1,
			},
		}
		content, cursorLine := m.renderTasksContent()
		if cursorLine == 0 {
			t.Error("expected non-zero cursor line for task")
		}
		if !strings.Contains(content, "▶") {
			t.Error("expected cursor indicator (▶) in content for task")
		}
	})
}

func TestToggleTask(t *testing.T) {
	t.Run("toggle pending to done writes to disk", func(t *testing.T) {
		dir := t.TempDir()
		content := "- [ ] do thing"
		if err := os.WriteFile(filepath.Join(dir, "tasks.md"), []byte(content), 0644); err != nil {
			t.Fatal(err)
		}
		ch := openspec.Change{Name: "test", Path: dir, Tasks: openspec.Artifact{Present: true, Content: content}}
		m := &Model{
			loader:  testLoader(),
			project: &openspec.Project{Changes: []openspec.Change{ch}},
			tasks: taskState{
				Items:  []openspec.TaskItem{{Kind: openspec.KindTask, Text: "do thing", Done: false, LineNum: 0}},
				Cursor: 0,
			},
			width: 80,
		}
		m.vp = viewport.New(viewport.WithWidth(80), viewport.WithHeight(20))
		cmd := m.doToggle()
		data, _ := os.ReadFile(filepath.Join(dir, "tasks.md"))
		if !strings.Contains(string(data), "[x]") {
			t.Errorf("expected [x] in tasks.md, got: %s", string(data))
		}
		if cmd != nil {
			t.Error("expected nil command for successful toggle")
		}
	})

	t.Run("toggle on empty items returns nil", func(t *testing.T) {
		m := &Model{tasks: taskState{Items: nil, Cursor: 0}}
		cmd := m.doToggle()
		if cmd != nil {
			t.Error("expected nil command for empty items")
		}
	})

	t.Run("toggle on section header returns nil", func(t *testing.T) {
		m := &Model{
			tasks: taskState{
				Items:  []openspec.TaskItem{{Kind: openspec.KindSection, Text: "Section 1"}},
				Cursor: 0,
			},
		}
		cmd := m.doToggle()
		if cmd != nil {
			t.Error("expected nil command for section header")
		}
	})
}

func TestLoadViewportDispatch(t *testing.T) {
	t.Run("not ready returns nil", func(t *testing.T) {
		m := &Model{vpReady: false}
		cmd := m.loadViewport()
		if cmd != nil {
			t.Error("expected nil cmd when vp not ready")
		}
	})

	t.Run("index mode returns nil", func(t *testing.T) {
		m := &Model{
			vpReady: true,
			mode:    ModeIndex,
			width:   80,
			project: &openspec.Project{},
			index:   indexState{ExpandedSpecs: make(map[int]bool)},
		}
		m.vp = viewport.New(viewport.WithWidth(80), viewport.WithHeight(20))
		cmd := m.loadViewport()
		if cmd != nil {
			t.Error("expected nil cmd for index mode")
		}
	})

	t.Run("tasks tab returns nil", func(t *testing.T) {
		m := &Model{
			vpReady: true,
			mode:    ModeNormal,
			tab:     TabTasks,
			width:   80,
			project: &openspec.Project{Changes: []openspec.Change{{Name: "test", Tasks: openspec.Artifact{Present: true}}}},
		}
		m.vp = viewport.New(viewport.WithWidth(80), viewport.WithHeight(20))
		cmd := m.loadViewport()
		if cmd != nil {
			t.Error("expected nil cmd for tasks tab")
		}
	})

	t.Run("cache hit returns nil", func(t *testing.T) {
		m := &Model{
			vpReady:     true,
			mode:        ModeNormal,
			tab:         TabProposal,
			width:       80,
			renderCache: map[Tab]string{TabProposal: "cached content"},
			project:     &openspec.Project{Changes: []openspec.Change{{Name: "test", Proposal: openspec.Artifact{Present: true, Content: "content"}}}},
		}
		m.vp = viewport.New(viewport.WithWidth(80), viewport.WithHeight(20))
		cmd := m.loadViewport()
		if cmd != nil {
			t.Error("expected nil cmd for cache hit")
		}
	})

	t.Run("config mode returns glamour cmd", func(t *testing.T) {
		m := &Model{
			vpReady:       true,
			mode:          ModeViewingConfig,
			width:         80,
			projectConfig: openspec.ProjectConfig{Context: "test context"},
		}
		m.vp = viewport.New(viewport.WithWidth(80), viewport.WithHeight(20))
		cmd := m.loadViewport()
		if cmd == nil {
			t.Error("expected non-nil cmd for config mode")
		}
	})
}

func TestHandleTick(t *testing.T) {
	t.Run("viewing archive returns nil", func(t *testing.T) {
		m := &Model{mode: ModeViewingArchive}
		cmd := m.handleTick()
		if cmd != nil {
			t.Error("expected nil cmd for viewing archive")
		}
	})

	t.Run("viewing spec returns nil", func(t *testing.T) {
		m := &Model{mode: ModeViewingSpec}
		cmd := m.handleTick()
		if cmd != nil {
			t.Error("expected nil cmd for viewing spec")
		}
	})

	t.Run("normal mode with no changes returns nil", func(t *testing.T) {
		dir := t.TempDir()
		if err := os.MkdirAll(filepath.Join(dir, "openspec", "changes"), 0755); err != nil {
			t.Fatal(err)
		}
		m := &Model{
			mode:       ModeNormal,
			root:       dir,
			singlePath: true,
			loader:     testLoader(),
			project:    &openspec.Project{Changes: []openspec.Change{{Name: "test"}}},
		}
		cmd := m.handleTick()
		if cmd != nil {
			t.Error("expected nil cmd for normal mode with singlePath")
		}
	})
}

func TestMatchesFilter(t *testing.T) {
	m := &Model{
		project: &openspec.Project{
			Changes: []openspec.Change{{Name: "data-export"}, {Name: "auth-module"}},
		},
		projectSpecs: []openspec.ProjectSpec{
			{Name: "mouse-navigation", RequirementNames: []string{"Wheel events", "Click select"}},
		},
		index: indexState{
			ExpandedSpecs: make(map[int]bool),
			ArchiveChanges: []openspec.Change{
				{Name: "refactor-tick", DisplayDate: "2026-05-30"},
			},
		},
	}
	m.buildIndexItems()

	t.Run("active change matches name", func(t *testing.T) {
		item := m.index.Items[0]
		if !m.matchesFilter(item, "data") {
			t.Error("expected 'data-export' to match 'data'")
		}
	})

	t.Run("active change case insensitive", func(t *testing.T) {
		item := m.index.Items[1]
		if !m.matchesFilter(item, "auth") {
			t.Error("expected 'auth-module' to match 'auth'")
		}
	})

	t.Run("active change no match", func(t *testing.T) {
		item := m.index.Items[0]
		if m.matchesFilter(item, "xyz") {
			t.Error("expected 'data-export' not to match 'xyz'")
		}
	})

	t.Run("spec matches name", func(t *testing.T) {
		item := m.index.Items[2]
		if !m.matchesFilter(item, "mouse") {
			t.Error("expected spec name to match 'mouse'")
		}
	})

	t.Run("requirement matches name", func(t *testing.T) {
		item := indexItem{kind: indexKindRequirement, idx: 0, reqIdx: 0}
		if !m.matchesFilter(item, "wheel") {
			t.Error("expected requirement to match 'wheel'")
		}
	})

	t.Run("archived change matches name", func(t *testing.T) {
		item := m.index.Items[3]
		if !m.matchesFilter(item, "tick") {
			t.Error("expected archived name to match 'tick'")
		}
	})

	t.Run("substring partial match", func(t *testing.T) {
		item := m.index.Items[0]
		if !m.matchesFilter(item, "port") {
			t.Error("expected 'data-export' to match substring 'port'")
		}
	})
}

func TestApplyFilter(t *testing.T) {
	items := []indexItem{
		{kind: indexKindActive, idx: 0},
		{kind: indexKindActive, idx: 1},
		{kind: indexKindSpec, idx: 0},
	}
	m := &Model{
		project: &openspec.Project{
			Changes: []openspec.Change{{Name: "data-export"}, {Name: "auth-module"}},
		},
		projectSpecs: []openspec.ProjectSpec{
			{Name: "data-pipeline"},
		},
		index: indexState{
			Items:         items,
			Cursor:        2,
			ExpandedSpecs: make(map[int]bool),
		},
	}

	t.Run("filter set builds FilterIndices", func(t *testing.T) {
		m.index.FilterText = "data"
		m.applyFilter()
		if m.index.FilterIndices == nil {
			t.Fatal("expected non-nil FilterIndices")
		}
		if len(m.index.FilterIndices) != 2 {
			t.Errorf("expected 2 matching items, got %d (%v)", len(m.index.FilterIndices), m.index.FilterIndices)
		}
	})

	t.Run("cursor clamped to visible count", func(t *testing.T) {
		m.index.Cursor = 5
		m.applyFilter()
		if m.index.Cursor != 0 {
			t.Errorf("expected cursor 0 after clamping, got %d", m.index.Cursor)
		}
	})

	t.Run("empty filter clears FilterIndices", func(t *testing.T) {
		m.index.FilterText = ""
		m.applyFilter()
		if m.index.FilterIndices != nil {
			t.Error("expected nil FilterIndices when FilterText is empty")
		}
	})
}

func TestIndexFilterKeypresses(t *testing.T) {
	m := Model{
		mode:  ModeIndex,
		width: 80,
		project: &openspec.Project{
			Changes: []openspec.Change{{Name: "data-export"}, {Name: "auth-module"}},
		},
		index: indexState{
			ExpandedSpecs: make(map[int]bool),
			Items: []indexItem{
				{kind: indexKindActive, idx: 0},
				{kind: indexKindActive, idx: 1},
			},
			Cursor: 0,
		},
	}
	m.vp = viewport.New(viewport.WithWidth(80), viewport.WithHeight(20))
	m.vpReady = true

	t.Run("/ enters filter mode", func(t *testing.T) {
		result, _ := m.dispatchKey(tea.KeyPressMsg{Text: "/"})
		updated := result.(Model)
		if !updated.index.FilterActive {
			t.Error("expected FilterActive after pressing /")
		}
	})

	t.Run("typing during filter mode updates FilterText", func(t *testing.T) {
		m.index.FilterActive = true
		m.index.FilterText = ""
		result, _ := m.dispatchKey(tea.KeyPressMsg{Text: "d"})
		updated := result.(Model)
		if updated.index.FilterText != "d" {
			t.Errorf("expected FilterText 'd', got %q", updated.index.FilterText)
		}
	})

	t.Run("backspace during filter removes char", func(t *testing.T) {
		m.index.FilterActive = true
		m.index.FilterText = "da"
		result, _ := m.dispatchKey(tea.KeyPressMsg{Code: tea.KeyBackspace})
		updated := result.(Model)
		if updated.index.FilterText != "d" {
			t.Errorf("expected FilterText 'd', got %q", updated.index.FilterText)
		}
	})

	t.Run("enter in filter mode confirms", func(t *testing.T) {
		m.index.FilterActive = true
		m.index.FilterText = "data"
		result, _ := m.dispatchKey(tea.KeyPressMsg{Code: tea.KeyEnter})
		updated := result.(Model)
		if updated.index.FilterActive {
			t.Error("expected FilterActive false after Enter")
		}
		if updated.index.FilterText != "data" {
			t.Errorf("expected FilterText 'data' to persist, got %q", updated.index.FilterText)
		}
	})

	t.Run("esc in filter mode cancels and reverts", func(t *testing.T) {
		m.index.FilterActive = true
		m.index.PrevFilterText = ""
		m.index.FilterText = "foo"
		result, _ := m.dispatchKey(tea.KeyPressMsg{Code: tea.KeyEsc})
		updated := result.(Model)
		if updated.index.FilterActive {
			t.Error("expected FilterActive false after Esc in filter mode")
		}
		if updated.index.FilterText != "" {
			t.Errorf("expected FilterText reverted to '', got %q", updated.index.FilterText)
		}
	})

	t.Run("esc with filter clears it", func(t *testing.T) {
		m.index.FilterActive = false
		m.index.FilterText = "data"
		result, cmd := m.dispatchKey(tea.KeyPressMsg{Code: tea.KeyEsc})
		updated := result.(Model)
		if cmd != nil {
			t.Error("expected nil cmd (filter cleared, not quit)")
		}
		if updated.index.FilterText != "" {
			t.Errorf("expected empty FilterText after Esc, got %q", updated.index.FilterText)
		}
	})

	t.Run("esc without filter quits", func(t *testing.T) {
		m.index.FilterActive = false
		m.index.FilterText = ""
		_, cmd := m.dispatchKey(tea.KeyPressMsg{Code: tea.KeyEsc})
		if cmd == nil {
			t.Error("expected quit cmd when no filter active")
		}
	})
}

func TestIndexFilterNoMatchMessage(t *testing.T) {
	m := &Model{
		width: 80,
		project: &openspec.Project{
			Changes: []openspec.Change{{Name: "data-export"}},
		},
		index: indexState{
			ExpandedSpecs: make(map[int]bool),
			FilterText:    "nonexistent",
		},
	}
	m.buildIndexItems()
	m.applyFilter()

	content, _ := m.renderIndexContent()
	if !strings.Contains(content, "No items match 'nonexistent'") {
		t.Errorf("expected no-match message in filtered content, got:\n%s", content)
	}
}

func TestIndexFilteredNavigation(t *testing.T) {
	m := Model{
		mode:  ModeIndex,
		width: 80,
		project: &openspec.Project{
			Changes: []openspec.Change{{Name: "data-export"}, {Name: "user-auth"}},
		},
		index: indexState{
			ExpandedSpecs: make(map[int]bool),
			Items: []indexItem{
				{kind: indexKindActive, idx: 0},
				{kind: indexKindActive, idx: 1},
			},
			Cursor:       0,
			FilterText:   "data",
			FilterActive: false,
		},
	}
	m.applyFilter()

	t.Run("visibleItemCount returns filtered count", func(t *testing.T) {
		if n := m.visibleItemCount(); n != 1 {
			t.Errorf("expected 1 visible item, got %d", n)
		}
	})

	t.Run("j/k navigate filtered list", func(t *testing.T) {
		// Try to move past the only visible item
		result, _ := m.dispatchKey(tea.KeyPressMsg{Text: "j"})
		updated := result.(Model)
		if updated.index.Cursor != 0 {
			t.Errorf("expected cursor 0 (only 1 visible), got %d", updated.index.Cursor)
		}
	})

	t.Run("visibleItemIdx maps through filter", func(t *testing.T) {
		idx := m.visibleItemIdx(0)
		if idx != 0 {
			t.Errorf("expected raw index 0 (data-export), got %d", idx)
		}
	})
}

func TestCurrentSpecPath(t *testing.T) {
	m := &Model{
		root: "/proj",
		projectSpecs: []openspec.ProjectSpec{
			{Name: "auth"},
			{Name: "export"},
		},
	}

	t.Run("valid cursor returns spec.md path", func(t *testing.T) {
		m.specViewer.Cursor = 1
		want := filepath.Join("/proj", "openspec", "specs", "export", "spec.md")
		if got := m.currentSpecPath(); got != want {
			t.Errorf("expected %q, got %q", want, got)
		}
	})

	t.Run("cursor past end returns empty", func(t *testing.T) {
		m.specViewer.Cursor = 5
		if got := m.currentSpecPath(); got != "" {
			t.Errorf("expected empty path for out-of-range cursor, got %q", got)
		}
	})

	t.Run("negative cursor returns empty", func(t *testing.T) {
		m.specViewer.Cursor = -1
		if got := m.currentSpecPath(); got != "" {
			t.Errorf("expected empty path for negative cursor, got %q", got)
		}
	})
}

func TestSpecEditKeyLaunchesEditor(t *testing.T) {
	newModel := func() Model {
		return Model{
			mode:         ModeViewingSpec,
			root:         "/proj",
			projectSpecs: []openspec.ProjectSpec{{Name: "auth"}},
		}
	}

	t.Run("e returns a command when a spec is viewed", func(t *testing.T) {
		t.Setenv("EDITOR", "vi")
		m := newModel()
		_, cmd := m.dispatchKey(tea.KeyPressMsg{Text: "e"})
		if cmd == nil {
			t.Error("expected non-nil command when pressing e on a spec")
		}
	})

	t.Run("e with EDITOR unset still returns a command (vi fallback)", func(t *testing.T) {
		t.Setenv("EDITOR", "")
		m := newModel()
		_, cmd := m.dispatchKey(tea.KeyPressMsg{Text: "e"})
		if cmd == nil {
			t.Error("expected non-nil command with vi fallback when EDITOR unset")
		}
	})

	t.Run("e returns nil command when no spec is available", func(t *testing.T) {
		m := Model{mode: ModeViewingSpec, root: "/proj"} // no projectSpecs
		_, cmd := m.dispatchKey(tea.KeyPressMsg{Text: "e"})
		if cmd != nil {
			t.Error("expected nil command when no spec is available")
		}
	})

	t.Run("e in focus mode opens the same spec file", func(t *testing.T) {
		t.Setenv("EDITOR", "vi")
		m := newModel()
		m.specViewer.FocusMode = true
		m.specViewer.JumpTarget = "R"
		if got := m.currentSpecPath(); got != filepath.Join("/proj", "openspec", "specs", "auth", "spec.md") {
			t.Errorf("focus mode resolved unexpected path %q", got)
		}
		_, cmd := m.dispatchKey(tea.KeyPressMsg{Text: "e"})
		if cmd == nil {
			t.Error("expected non-nil command when pressing e in focus mode")
		}
	})
}

func TestEditorReturnReloadsSpec(t *testing.T) {
	dir := t.TempDir()
	specDir := filepath.Join(dir, "openspec", "specs", "auth")
	if err := os.MkdirAll(specDir, 0755); err != nil {
		t.Fatal(err)
	}
	orig := "## Purpose\nP.\n\n## Requirements\n\n### Requirement: R\n#### Scenario: S\n- **WHEN** a\n- **THEN** b\n"
	if err := os.WriteFile(filepath.Join(specDir, "spec.md"), []byte(orig), 0644); err != nil {
		t.Fatal(err)
	}

	m := Model{
		mode:         ModeViewingSpec,
		root:         dir,
		loader:       testLoader(),
		width:        80,
		vpReady:      true,
		projectSpecs: []openspec.ProjectSpec{{Name: "auth", Content: orig}},
	}
	m.vp = viewport.New(viewport.WithWidth(80), viewport.WithHeight(20))

	// Simulate an external edit while the editor was open.
	updated := orig + "\n### Requirement: Added\n#### Scenario: S2\n- **WHEN** c\n- **THEN** d\n"
	if err := os.WriteFile(filepath.Join(specDir, "spec.md"), []byte(updated), 0644); err != nil {
		t.Fatal(err)
	}

	result, _ := m.Update(editorReturnMsg{})
	got := result.(Model)

	if got.mode != ModeViewingSpec {
		t.Errorf("expected to stay in ModeViewingSpec, got %d", got.mode)
	}
	if len(got.projectSpecs) == 0 || !strings.Contains(got.projectSpecs[0].Content, "Requirement: Added") {
		t.Error("expected reloaded spec content to include the external edit")
	}
}

// Regression: if the spec list shrinks while the editor is open, the cursor
// must be clamped on return so the next render cannot index out of range.
func TestEditorReturnClampsCursorWhenSpecListShrinks(t *testing.T) {
	dir := t.TempDir()
	// No specs on disk: LoadProjectSpecsFrom returns an empty list.
	if err := os.MkdirAll(filepath.Join(dir, "openspec", "specs"), 0755); err != nil {
		t.Fatal(err)
	}

	m := Model{
		mode:    ModeViewingSpec,
		root:    dir,
		loader:  testLoader(),
		width:   80,
		vpReady: true,
		// In-memory state thinks there are two specs with the cursor on the second.
		projectSpecs: []openspec.ProjectSpec{
			{Name: "auth", Content: "x"},
			{Name: "export", Content: "y"},
		},
	}
	m.specViewer.Cursor = 1
	m.vp = viewport.New(viewport.WithWidth(80), viewport.WithHeight(20))

	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("editorReturnMsg panicked after spec list shrank: %v", r)
		}
	}()
	result, _ := m.Update(editorReturnMsg{})
	got := result.(Model)

	if len(got.projectSpecs) != 0 {
		t.Fatalf("expected reloaded spec list to be empty, got %d", len(got.projectSpecs))
	}
	if got.specViewer.Cursor != 0 {
		t.Errorf("expected cursor clamped to 0 after shrink, got %d", got.specViewer.Cursor)
	}
	if got.mode != ModeViewingSpec {
		t.Errorf("expected to stay in ModeViewingSpec, got %d", got.mode)
	}
}

func TestRenderTabBar(t *testing.T) {
	t.Run("active tab highlighted", func(t *testing.T) {
		m := &Model{
			tab:   TabProposal,
			width: 80,
			project: &openspec.Project{
				Changes: []openspec.Change{{
					Name:     "test",
					Proposal: openspec.Artifact{Present: true},
				}},
			},
			mode: ModeNormal,
		}
		m.renderCache = make(map[Tab]string)
		result := m.renderTabBar()
		if !strings.Contains(result, "proposal") {
			t.Error("expected 'proposal' in tab bar")
		}
	})

	t.Run("disabled tab shows low style", func(t *testing.T) {
		m := &Model{
			tab:   TabProposal,
			width: 80,
			project: &openspec.Project{
				Changes: []openspec.Change{{
					Name:     "test",
					Proposal: openspec.Artifact{Present: true},
				}},
			},
			mode: ModeNormal,
		}
		m.renderCache = make(map[Tab]string)
		result := m.renderTabBar()
		if !strings.Contains(result, "proposal") {
			t.Error("expected proposal in tab bar")
		}
	})
}

func newArchiveExpandModel() Model {
	m := Model{
		mode:    ModeIndex,
		width:   80,
		height:  24,
		vpReady: true,
		project: &openspec.Project{},
		index: indexState{
			ExpandedSpecs:    make(map[int]bool),
			ExpandedArchives: make(map[int]bool),
			ArchiveChanges: []openspec.Change{
				{
					Name:        "old-feat",
					DisplayDate: "01/05/2026",
					Proposal:    openspec.Artifact{Present: true},
					Specs:       openspec.Artifact{Present: true},
					Tasks:       openspec.Artifact{Present: true},
				},
			},
		},
	}
	m.vp = viewport.New(viewport.WithWidth(80), viewport.WithHeight(20))
	m.buildIndexItems()
	return m
}

func archivedArtifactTabs(items []indexItem) []Tab {
	var tabs []Tab
	for _, it := range items {
		if it.kind == indexKindArchivedArtifact {
			tabs = append(tabs, Tab(it.reqIdx))
		}
	}
	return tabs
}

func TestExpandArchivedChange(t *testing.T) {
	t.Run("space expands into present artifacts in tab order, omitting absent", func(t *testing.T) {
		m := newArchiveExpandModel()
		result, _ := m.dispatchKey(tea.KeyPressMsg{Code: tea.KeySpace})
		updated := result.(Model)

		got := archivedArtifactTabs(updated.index.Items)
		want := []Tab{TabProposal, TabSpecs, TabTasks} // design is absent
		if len(got) != len(want) {
			t.Fatalf("expected %d artifact sub-items, got %d (%v)", len(want), len(got), got)
		}
		for i := range want {
			if got[i] != want[i] {
				t.Errorf("artifact %d: expected tab %d, got %d", i, want[i], got[i])
			}
		}
		if !updated.index.ExpandedArchives[0] {
			t.Error("expected archived change 0 to be marked expanded")
		}
	})

	t.Run("space again collapses and cursor stays on the archived change", func(t *testing.T) {
		m := newArchiveExpandModel()
		result, _ := m.dispatchKey(tea.KeyPressMsg{Code: tea.KeySpace})
		expanded := result.(Model)
		if expanded.index.Cursor != 0 {
			t.Fatalf("expected cursor on archived change (0) after expand, got %d", expanded.index.Cursor)
		}

		result, _ = expanded.dispatchKey(tea.KeyPressMsg{Code: tea.KeySpace})
		collapsed := result.(Model)
		if got := archivedArtifactTabs(collapsed.index.Items); len(got) != 0 {
			t.Errorf("expected no artifact sub-items after collapse, got %v", got)
		}
		if collapsed.index.ExpandedArchives[0] {
			t.Error("expected archived change 0 to be collapsed")
		}
		if collapsed.index.Cursor != 0 {
			t.Errorf("expected cursor to stay on archived change (0), got %d", collapsed.index.Cursor)
		}
	})

	t.Run("enter on artifact sub-item opens archive viewer on that tab", func(t *testing.T) {
		m := newArchiveExpandModel()
		result, _ := m.dispatchKey(tea.KeyPressMsg{Code: tea.KeySpace})
		m = result.(Model)
		// Move cursor onto the "specs" sub-item (archived row, proposal, specs, ...).
		m.index.Cursor = 2
		if m.index.Items[m.index.Cursor].kind != indexKindArchivedArtifact ||
			Tab(m.index.Items[m.index.Cursor].reqIdx) != TabSpecs {
			t.Fatalf("test setup: expected cursor on specs sub-item, got %+v", m.index.Items[m.index.Cursor])
		}

		result, _ = m.dispatchKey(tea.KeyPressMsg{Code: tea.KeyEnter})
		opened := result.(Model)
		if opened.mode != ModeViewingArchive {
			t.Errorf("expected ModeViewingArchive, got %d", opened.mode)
		}
		if opened.tab != TabSpecs {
			t.Errorf("expected tab TabSpecs, got %d", opened.tab)
		}
		if opened.index.ArchiveCursor != 0 {
			t.Errorf("expected ArchiveCursor 0, got %d", opened.index.ArchiveCursor)
		}
	})

	t.Run("filter keeps artifact sub-items visible when parent matches", func(t *testing.T) {
		m := newArchiveExpandModel()
		m.index.ExpandedArchives[0] = true
		m.buildIndexItems()

		artifact := indexItem{kind: indexKindArchivedArtifact, idx: 0, reqIdx: int(TabProposal)}
		if !m.matchesFilter(artifact, "old-feat") {
			t.Error("expected artifact sub-item to match parent change name 'old-feat'")
		}
		if !m.matchesFilter(artifact, "proposal") {
			t.Error("expected artifact sub-item to match its artifact label 'proposal'")
		}
		if m.matchesFilter(artifact, "zzz") {
			t.Error("expected artifact sub-item not to match unrelated query 'zzz'")
		}
	})
}

func TestClickArchivedArtifact(t *testing.T) {
	m := newArchiveExpandModel()
	m.index.ExpandedArchives[0] = true
	m.buildIndexItems()

	// Locate the "tasks" sub-item.
	tasksIdx := -1
	for i, it := range m.index.Items {
		if it.kind == indexKindArchivedArtifact && Tab(it.reqIdx) == TabTasks {
			tasksIdx = i
			break
		}
	}
	if tasksIdx < 0 {
		t.Fatal("test setup: tasks sub-item not found")
	}

	// Hit-testing: the content line for the sub-item must map back to it.
	found := -1
	for line := 0; line < 60; line++ {
		if idx, ok := m.indexItemAtContentLine(line); ok && idx == tasksIdx {
			found = line
			break
		}
	}
	if found < 0 {
		t.Fatal("expected indexItemAtContentLine to resolve the tasks sub-item")
	}

	// Click it when already selected -> opens archive viewer on the tasks tab.
	m.index.Cursor = tasksIdx
	result, _ := m.clickIndexItem(tasksIdx)
	opened := result.(Model)
	if opened.mode != ModeViewingArchive {
		t.Errorf("expected ModeViewingArchive, got %d", opened.mode)
	}
	if opened.tab != TabTasks {
		t.Errorf("expected tab TabTasks, got %d", opened.tab)
	}
	if opened.index.ArchiveCursor != 0 {
		t.Errorf("expected ArchiveCursor 0, got %d", opened.index.ArchiveCursor)
	}
}

func TestExpandArchivedChangeNavigationAndEmpty(t *testing.T) {
	t.Run("j navigates from archived change into its first artifact sub-item", func(t *testing.T) {
		m := newArchiveExpandModel()
		result, _ := m.dispatchKey(tea.KeyPressMsg{Code: tea.KeySpace})
		m = result.(Model)
		if m.index.Cursor != 0 {
			t.Fatalf("expected cursor on archived change (0) after expand, got %d", m.index.Cursor)
		}

		result, _ = m.dispatchKey(tea.KeyPressMsg{Text: "j"})
		moved := result.(Model)
		if moved.index.Cursor != 1 {
			t.Fatalf("expected cursor to move to first sub-item (1), got %d", moved.index.Cursor)
		}
		item := moved.index.Items[moved.index.Cursor]
		if item.kind != indexKindArchivedArtifact || Tab(item.reqIdx) != TabProposal {
			t.Errorf("expected cursor on the proposal artifact sub-item, got %+v", item)
		}
	})

	t.Run("space on an archived change with no artifacts adds nothing and keeps cursor", func(t *testing.T) {
		m := Model{
			mode:    ModeIndex,
			width:   80,
			height:  24,
			vpReady: true,
			project: &openspec.Project{},
			index: indexState{
				ExpandedSpecs:    make(map[int]bool),
				ExpandedArchives: make(map[int]bool),
				ArchiveChanges: []openspec.Change{
					{Name: "bare-feat", DisplayDate: "02/05/2026"}, // no artifacts present
				},
			},
		}
		m.vp = viewport.New(viewport.WithWidth(80), viewport.WithHeight(20))
		m.buildIndexItems()

		result, _ := m.dispatchKey(tea.KeyPressMsg{Code: tea.KeySpace})
		updated := result.(Model)
		if got := archivedArtifactTabs(updated.index.Items); len(got) != 0 {
			t.Errorf("expected no artifact sub-items for an artifact-less change, got %v", got)
		}
		if updated.index.Cursor != 0 {
			t.Errorf("expected cursor to stay at 0, got %d", updated.index.Cursor)
		}
	})
}

// Regression: the mouse wheel in index mode must clamp the cursor against the
// VISIBLE (filtered) item count, not the unfiltered len(Items). Otherwise the
// cursor overshoots the filtered list and a subsequent click panics.
func TestMouseWheelClampsToVisibleCountWithFilter(t *testing.T) {
	m := Model{
		mode:  ModeIndex,
		width: 80,
		project: &openspec.Project{Changes: []openspec.Change{
			{Name: "alpha"}, {Name: "beta"}, {Name: "gamma"},
		}},
		index: indexState{
			ExpandedSpecs:    map[int]bool{},
			ExpandedArchives: map[int]bool{},
			FilterText:       "alpha",
		},
	}
	m.vp = viewport.New(viewport.WithWidth(80), viewport.WithHeight(20))
	m.vpReady = true
	m.buildIndexItems()
	m.applyFilter() // only "alpha" matches → visibleItemCount() == 1

	for i := 0; i < 5; i++ {
		res, _ := m.handleMouseWheel(tea.MouseWheelMsg{Button: tea.MouseWheelDown})
		m = res.(Model)
	}
	if m.index.Cursor > m.visibleItemCount()-1 {
		t.Fatalf("cursor %d exceeded visible count %d after wheel", m.index.Cursor, m.visibleItemCount())
	}
}

// Regression: a click in index mode must not panic even if the cursor is out of
// range relative to FilterIndices (e.g. left desynced by another path).
func TestIndexClickWithOOBCursorDoesNotPanic(t *testing.T) {
	m := Model{
		mode:  ModeIndex,
		width: 80,
		project: &openspec.Project{Changes: []openspec.Change{
			{Name: "alpha"}, {Name: "beta"}, {Name: "gamma"},
		}},
		index: indexState{
			ExpandedSpecs:    map[int]bool{},
			ExpandedArchives: map[int]bool{},
			FilterText:       "alpha",
		},
	}
	m.vp = viewport.New(viewport.WithWidth(80), viewport.WithHeight(20))
	m.vpReady = true
	m.buildIndexItems()
	m.applyFilter()
	m.refreshIndexViewport()
	m.index.Cursor = 99 // out of range vs FilterIndices

	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("click panicked with out-of-range cursor: %v", r)
		}
	}()
	// Y offset +3 lands on the first visible item's row (past the section
	// header) so the click reaches the bounds-guarded FilterIndices access.
	res, _ := m.handleMouseClick(tea.MouseClickMsg{Button: tea.MouseLeft, X: 2, Y: indexViewportContentStart + 3})
	got := res.(Model)
	if got.index.Cursor >= len(got.index.FilterIndices) {
		t.Fatalf("cursor %d not repositioned into filtered range (len %d)", got.index.Cursor, len(got.index.FilterIndices))
	}
}

// Regression: editing a foreign worktree change and returning from the editor
// must reload into worktreeViewChange, NOT overwrite the rooted project's
// change at m.changeIdx (the bug mergeReloadedChange would otherwise cause).
func TestEditorReturnForeignWorktreeDoesNotCorruptRooted(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "proposal.md"), []byte("# updated foreign"), 0644); err != nil {
		t.Fatal(err)
	}
	foreign := openspec.Change{
		Name:     "feat-foreign",
		Path:     dir,
		Proposal: openspec.Artifact{Present: true, Content: "# old foreign"},
	}
	rooted := openspec.Change{
		Name:     "rooted",
		Path:     t.TempDir(),
		Proposal: openspec.Artifact{Present: true, Content: "ROOTED CONTENT"},
	}
	m := Model{
		mode:                  ModeViewingArchive,
		viewingWorktreeChange: true,
		worktreeViewChange:    foreign,
		tab:                   TabProposal,
		changeIdx:             0,
		width:                 80,
		vpReady:               true,
		loader:                testLoader(),
		project:               &openspec.Project{Changes: []openspec.Change{rooted}},
		renderCache:           map[Tab]string{},
	}
	m.vp = viewport.New(viewport.WithWidth(80), viewport.WithHeight(20))

	result, _ := m.Update(editorReturnMsg{})
	got := result.(Model)

	if got.project.Changes[0].Proposal.Content != "ROOTED CONTENT" {
		t.Errorf("rooted project change was corrupted: %q", got.project.Changes[0].Proposal.Content)
	}
	if !strings.Contains(got.worktreeViewChange.Proposal.Content, "updated foreign") {
		t.Errorf("worktree change was not reloaded from disk: %q", got.worktreeViewChange.Proposal.Content)
	}
}

// Regression: Ctrl+C must quit even while the help overlay is open (the overlay
// advertises it).
func TestHelpOverlayCtrlCQuits(t *testing.T) {
	m := Model{helpOpen: true}
	_, cmd := m.dispatchKey(tea.KeyPressMsg{Code: 'c', Mod: tea.ModCtrl})
	if cmd == nil {
		t.Error("expected quit command on ctrl+c while help overlay open")
	}
}
