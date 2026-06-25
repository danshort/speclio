package openspec

import "testing"

func TestParseWorktreeList(t *testing.T) {
	out := `worktree /repo/main
HEAD 1111111111111111111111111111111111111111
branch refs/heads/main

worktree /repo/.worktrees/feature
HEAD 2222222222222222222222222222222222222222
branch refs/heads/feature/cool-thing

worktree /repo/.worktrees/detached
HEAD 3333333333333333333333333333333333333333
detached

worktree /repo/.worktrees/locked
HEAD 4444444444444444444444444444444444444444
branch refs/heads/locked-branch
locked needs review

worktree /repo/.worktrees/prunable
HEAD 5555555555555555555555555555555555555555
branch refs/heads/gone
prunable gitdir file points to non-existent location

worktree /repo/bare
bare
`

	wts := parseWorktreeList([]byte(out))
	if len(wts) != 6 {
		t.Fatalf("expected 6 worktrees, got %d", len(wts))
	}

	t.Run("branch entry", func(t *testing.T) {
		w := wts[0]
		if w.Path != "/repo/main" {
			t.Errorf("path = %q", w.Path)
		}
		if w.Branch != "main" {
			t.Errorf("branch = %q, want main", w.Branch)
		}
		if w.Head != "1111111111111111111111111111111111111111" {
			t.Errorf("head = %q", w.Head)
		}
		if w.Detached || w.Bare || w.Locked || w.Prunable {
			t.Errorf("expected no flags, got %+v", w)
		}
	})

	t.Run("branch with slashes is stripped of refs/heads/", func(t *testing.T) {
		if wts[1].Branch != "feature/cool-thing" {
			t.Errorf("branch = %q, want feature/cool-thing", wts[1].Branch)
		}
	})

	t.Run("detached HEAD", func(t *testing.T) {
		w := wts[2]
		if !w.Detached {
			t.Error("expected Detached")
		}
		if w.Branch != "" {
			t.Errorf("branch = %q, want empty for detached", w.Branch)
		}
	})

	t.Run("locked entry", func(t *testing.T) {
		w := wts[3]
		if !w.Locked {
			t.Error("expected Locked")
		}
		if w.Branch != "locked-branch" {
			t.Errorf("branch = %q", w.Branch)
		}
	})

	t.Run("prunable entry", func(t *testing.T) {
		if !wts[4].Prunable {
			t.Error("expected Prunable")
		}
	})

	t.Run("bare entry", func(t *testing.T) {
		w := wts[5]
		if !w.Bare {
			t.Error("expected Bare")
		}
		if w.Branch != "" {
			t.Errorf("branch = %q, want empty for bare", w.Branch)
		}
	})
}

func TestMarkCurrentWorktree(t *testing.T) {
	wts := []Worktree{
		{Path: "/repo/main"},
		{Path: "/repo/.worktrees/feature"},
	}
	markCurrentWorktree(wts, "/repo/.worktrees/feature")
	if wts[0].IsCurrent {
		t.Error("main should not be current")
	}
	if !wts[1].IsCurrent {
		t.Error("feature should be current")
	}
}

func TestMarkCurrentWorktreeEmptyToplevel(t *testing.T) {
	wts := []Worktree{{Path: "/repo/main"}}
	markCurrentWorktree(wts, "")
	if wts[0].IsCurrent {
		t.Error("no worktree should be marked current when toplevel is empty")
	}
}
