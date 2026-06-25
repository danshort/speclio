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

func tasksArtifact(content string) openspec.Artifact {
	return openspec.Artifact{Content: content, Present: true}
}

// worktreesModel builds a Model already in ModeWorktrees with the given entries,
// with a ready viewport and items built for navigation.
func worktreesModel(entries []worktreeEntry) Model {
	m := Model{
		mode:    ModeWorktrees,
		width:   80,
		height:  24,
		project: &openspec.Project{Name: "proj"},
		loader:  testLoader(),
	}
	m.worktrees.Available = true
	m.worktrees.Entries = entries
	m.vp = viewport.New(viewport.WithWidth(78), viewport.WithHeight(20))
	m.vpReady = true
	m.buildWorktreeItems()
	return m
}

func TestSortWorktreeEntriesCurrentFirst(t *testing.T) {
	entries := []worktreeEntry{
		{wt: openspec.Worktree{Path: "/repo/feature", Branch: "feature"}},
		{wt: openspec.Worktree{Path: "/repo/main", Branch: "main", IsCurrent: true}},
		{wt: openspec.Worktree{Path: "/repo/other", Branch: "other"}},
	}
	sortWorktreeEntriesCurrentFirst(entries)
	if !entries[0].wt.IsCurrent {
		t.Fatalf("expected current worktree first, got %q", entries[0].wt.Branch)
	}
	if entries[1].wt.Branch != "feature" || entries[2].wt.Branch != "other" {
		t.Errorf("expected stable order for the rest, got %q, %q", entries[1].wt.Branch, entries[2].wt.Branch)
	}
}

func TestRenderWorktreesContent(t *testing.T) {
	entries := []worktreeEntry{
		{
			wt: openspec.Worktree{Path: "/repo/main", Branch: "main", IsCurrent: true},
			changes: []openspec.Change{
				{Name: "current-change", Tasks: tasksArtifact("- [ ] 1.1 a\n")},
			},
		},
		{
			wt:      openspec.Worktree{Path: "/repo/empty", Branch: "empty-branch"},
			changes: nil,
		},
		{
			wt: openspec.Worktree{Path: "/repo/feat", Branch: "feature"},
			changes: []openspec.Change{
				{Name: "foreign-change", Tasks: tasksArtifact("- [x] 1.1 a\n- [ ] 1.2 b\n")},
			},
		},
	}
	m := worktreesModel(entries)
	out, _ := m.renderWorktreesContent()

	t.Run("current worktree is first and badged", func(t *testing.T) {
		if !strings.Contains(out, "(current)") {
			t.Error("expected a (current) badge")
		}
		mainIdx := strings.Index(out, "main")
		featIdx := strings.Index(out, "feature")
		if mainIdx < 0 || featIdx < 0 || mainIdx > featIdx {
			t.Errorf("expected current worktree (main) before feature; mainIdx=%d featIdx=%d", mainIdx, featIdx)
		}
	})

	t.Run("empty worktree shows (no active changes)", func(t *testing.T) {
		if !strings.Contains(out, "(no active changes)") {
			t.Error("expected (no active changes) for the empty worktree")
		}
	})

	t.Run("foreign change shows progress bar from its tasks", func(t *testing.T) {
		if !strings.Contains(out, "1/2") {
			t.Errorf("expected a 1/2 progress count for the foreign change, got:\n%s", out)
		}
	})
}

func TestRenderWorktreesDetachedAndAnnotations(t *testing.T) {
	entries := []worktreeEntry{
		{wt: openspec.Worktree{Path: "/repo/det", Head: "abcdef1234567890", Detached: true}},
		{wt: openspec.Worktree{Path: "/repo/lk", Branch: "lk", Locked: true, Prunable: true}},
	}
	m := worktreesModel(entries)
	out, _ := m.renderWorktreesContent()

	if !strings.Contains(out, "abcdef1") {
		t.Error("expected detached worktree labelled by short SHA")
	}
	if strings.Contains(out, "abcdef1234567890") {
		t.Error("expected the SHA to be shortened, not full")
	}
	if !strings.Contains(out, "locked") || !strings.Contains(out, "prunable") {
		t.Errorf("expected locked/prunable annotations, got:\n%s", out)
	}
}

func TestWorktreesUnavailableRendersLine(t *testing.T) {
	m := Model{mode: ModeWorktrees, width: 80, height: 24}
	m.worktrees.Available = false
	m.worktrees.Message = "Worktrees unavailable: git is not on PATH or this is not a git working tree."

	// Must not panic and must surface the explanatory message.
	out, cursorLine := m.renderWorktreesContent()
	if cursorLine != 0 {
		t.Errorf("expected cursorLine 0 for unavailable view, got %d", cursorLine)
	}
	if !strings.Contains(out, "unavailable") {
		t.Errorf("expected an explanatory unavailable line, got:\n%s", out)
	}
}

func TestEnterWorktreesNonGitDirDegrades(t *testing.T) {
	dir := t.TempDir() // a fresh temp dir is not inside a git working tree
	m := Model{width: 80, height: 24, root: dir, project: &openspec.Project{}, loader: testLoader()}
	m.vp = viewport.New(viewport.WithWidth(78), viewport.WithHeight(20))
	m.vpReady = true

	m.enterWorktrees()

	if m.mode != ModeWorktrees {
		t.Fatalf("expected ModeWorktrees, got %d", m.mode)
	}
	if m.worktrees.Available {
		t.Error("expected worktrees to be unavailable outside a git working tree")
	}
	if m.worktrees.Message == "" {
		t.Error("expected an explanatory message when worktrees are unavailable")
	}
}

func TestWorktreesNavigationSkipsCurrent(t *testing.T) {
	entries := []worktreeEntry{
		{
			wt:      openspec.Worktree{Path: "/repo/main", Branch: "main", IsCurrent: true},
			changes: []openspec.Change{{Name: "current-change"}},
		},
		{
			wt:      openspec.Worktree{Path: "/repo/feat", Branch: "feature"},
			changes: []openspec.Change{{Name: "foreign-a"}, {Name: "foreign-b"}},
		},
	}
	m := worktreesModel(entries)

	// Only the two foreign changes are navigable; the current worktree's
	// change is excluded.
	if len(m.worktrees.Items) != 2 {
		t.Fatalf("expected 2 navigable items, got %d", len(m.worktrees.Items))
	}
	for _, it := range m.worktrees.Items {
		if m.worktrees.Entries[it.wtIdx].wt.IsCurrent {
			t.Error("current worktree change should not be navigable")
		}
	}

	res, _ := m.dispatchKey(tea.KeyPressMsg{Text: "j"})
	if got := res.(Model).worktrees.Cursor; got != 1 {
		t.Errorf("expected cursor 1 after j, got %d", got)
	}
}

func TestWorktreesEscReturnsToIndex(t *testing.T) {
	m := worktreesModel([]worktreeEntry{
		{wt: openspec.Worktree{Path: "/repo/main", Branch: "main", IsCurrent: true}},
	})
	res, _ := m.dispatchKey(tea.KeyPressMsg{Code: tea.KeyEsc})
	if got := res.(Model).mode; got != ModeIndex {
		t.Errorf("expected ModeIndex after esc, got %d", got)
	}
}

func TestEnterOpensForeignChangeReadOnly(t *testing.T) {
	// Build a real change dir so ReloadChange and the read-only toggle guard can
	// be exercised against disk.
	dir := t.TempDir()
	tasksPath := filepath.Join(dir, "tasks.md")
	original := "## 1. Section\n\n- [ ] 1.1 first task\n"
	if err := os.WriteFile(tasksPath, []byte(original), 0o644); err != nil {
		t.Fatal(err)
	}

	entries := []worktreeEntry{
		{
			wt: openspec.Worktree{Path: "/repo/feat", Branch: "feature"},
			changes: []openspec.Change{
				{Name: "foreign", Path: dir, Tasks: tasksArtifact(original)},
			},
		},
	}
	m := worktreesModel(entries)

	res, _ := m.dispatchKey(tea.KeyPressMsg{Code: tea.KeyEnter})
	opened := res.(Model)

	t.Run("opens via read-only archive path", func(t *testing.T) {
		if opened.mode != ModeViewingArchive {
			t.Fatalf("expected ModeViewingArchive, got %d", opened.mode)
		}
		if !opened.viewingWorktreeChange {
			t.Error("expected viewingWorktreeChange to be set")
		}
		ch := opened.current()
		if ch == nil || ch.Name != "foreign" {
			t.Fatalf("expected current() to be the foreign change, got %v", ch)
		}
		if opened.tab != TabTasks {
			t.Errorf("expected the tasks tab (only available artifact), got %d", opened.tab)
		}
	})

	t.Run("space does not toggle a foreign task", func(t *testing.T) {
		res2, _ := opened.dispatchKey(tea.KeyPressMsg{Text: " "})
		_ = res2
		got, err := os.ReadFile(tasksPath)
		if err != nil {
			t.Fatal(err)
		}
		if string(got) != original {
			t.Errorf("foreign tasks.md was modified; read-only guard failed:\n%s", string(got))
		}
	})

	t.Run("esc returns to the worktrees view", func(t *testing.T) {
		res2, _ := opened.dispatchKey(tea.KeyPressMsg{Code: tea.KeyEsc})
		back := res2.(Model)
		if back.mode != ModeWorktrees {
			t.Errorf("expected to return to ModeWorktrees, got %d", back.mode)
		}
		if back.viewingWorktreeChange {
			t.Error("expected viewingWorktreeChange cleared on return")
		}
	})
}

func TestPollWorktreesPicksUpProgress(t *testing.T) {
	dir := t.TempDir()
	tasksPath := filepath.Join(dir, "tasks.md")
	if err := os.WriteFile(tasksPath, []byte("- [ ] 1.1 a\n- [ ] 1.2 b\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	m := worktreesModel([]worktreeEntry{
		{
			wt:      openspec.Worktree{Path: "/repo/feat", Branch: "feature"},
			changes: []openspec.Change{{Name: "foreign", Path: dir, Tasks: tasksArtifact("- [ ] 1.1 a\n- [ ] 1.2 b\n")}},
		},
	})

	// Simulate an agent completing a task in the foreign worktree on disk.
	if err := os.WriteFile(tasksPath, []byte("- [x] 1.1 a\n- [ ] 1.2 b\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	m.pollWorktrees()

	got := m.worktrees.Entries[0].changes[0].Tasks.Content
	if !strings.Contains(got, "- [x] 1.1 a") {
		t.Errorf("expected pollWorktrees to refresh foreign tasks content, got:\n%s", got)
	}
	if done, total := taskCounts(m.worktrees.Entries[0].changes[0]); done != 1 || total != 2 {
		t.Errorf("expected progress 1/2 after poll, got %d/%d", done, total)
	}
}

func TestForeignChangeEditLaunchesEditor(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "tasks.md"), []byte("- [ ] 1.1 t\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	t.Setenv("EDITOR", "true") // a no-op editor that exits 0

	m := worktreesModel([]worktreeEntry{
		{
			wt:      openspec.Worktree{Path: "/repo/feat", Branch: "feature"},
			changes: []openspec.Change{{Name: "foreign", Path: dir, Tasks: tasksArtifact("- [ ] 1.1 t\n")}},
		},
	})
	res, _ := m.dispatchKey(tea.KeyPressMsg{Code: tea.KeyEnter})
	opened := res.(Model)

	_, cmd := opened.dispatchKey(tea.KeyPressMsg{Text: "e"})
	if cmd == nil {
		t.Error("expected pressing e on a foreign change to return an editor command")
	}
}
