package openspec

import (
	"context"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

// Worktree describes a single git worktree as reported by
// `git worktree list --porcelain`.
type Worktree struct {
	Path      string `json:"path"`
	Branch    string `json:"branch"` // short branch name (refs/heads/ stripped); empty when detached or bare
	Head      string `json:"head"`   // full HEAD SHA
	IsCurrent bool   `json:"is_current"`
	Detached  bool   `json:"detached"`
	Bare      bool   `json:"bare"`
	Locked    bool   `json:"locked"`
	Prunable  bool   `json:"prunable"`
}

// ListWorktrees enumerates the git worktrees sharing root's repository by
// invoking `git worktree list --porcelain`. It returns an error when git is
// not on PATH or root is not inside a git working tree, so callers can render a
// graceful "unavailable" state rather than failing. The worktree matching
// root's toplevel is marked IsCurrent.
func ListWorktrees(root string) ([]Worktree, error) {
	out, err := runGit(root, "worktree", "list", "--porcelain")
	if err != nil {
		return nil, err
	}
	wts := parseWorktreeList(out)
	if top, err := runGit(root, "rev-parse", "--show-toplevel"); err == nil {
		markCurrentWorktree(wts, strings.TrimSpace(string(top)))
	}
	return wts, nil
}

// #nosec G204 -- by design: invokes the git binary against the user's own
// project directory. No shell is involved (exec.Command does not interpret
// shell metacharacters), and the arguments are fixed literals.
func runGit(dir string, args ...string) ([]byte, error) {
	// Bound the call so a hung filesystem, lock, or pathological repo cannot
	// freeze the TUI (runGit is invoked synchronously on the update goroutine).
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, "git", append([]string{"-C", dir}, args...)...)
	return cmd.Output()
}

// parseWorktreeList parses the porcelain output of `git worktree list`. Each
// worktree is a block of `key value` lines terminated by a blank line; valueless
// keys (detached, bare, locked, prunable) appear without a trailing value.
func parseWorktreeList(out []byte) []Worktree {
	var wts []Worktree
	var cur *Worktree
	flush := func() {
		if cur != nil {
			wts = append(wts, *cur)
			cur = nil
		}
	}
	for _, line := range strings.Split(string(out), "\n") {
		line = strings.TrimRight(line, "\r")
		if line == "" {
			flush()
			continue
		}
		key, val, _ := strings.Cut(line, " ")
		switch key {
		case "worktree":
			flush()
			cur = &Worktree{Path: val}
		case "HEAD":
			if cur != nil {
				cur.Head = val
			}
		case "branch":
			if cur != nil {
				cur.Branch = strings.TrimPrefix(val, "refs/heads/")
			}
		case "detached":
			if cur != nil {
				cur.Detached = true
			}
		case "bare":
			if cur != nil {
				cur.Bare = true
			}
		case "locked":
			if cur != nil {
				cur.Locked = true
			}
		case "prunable":
			if cur != nil {
				cur.Prunable = true
			}
		}
	}
	flush()
	return wts
}

func markCurrentWorktree(wts []Worktree, toplevel string) {
	if toplevel == "" {
		return
	}
	target := normalizePath(toplevel)
	for i := range wts {
		if normalizePath(wts[i].Path) == target {
			wts[i].IsCurrent = true
			return
		}
	}
}

// normalizePath resolves symlinks so a worktree path and root toplevel compare
// equal even when one side traverses a symlink (e.g. macOS /var → /private/var).
func normalizePath(p string) string {
	if resolved, err := filepath.EvalSymlinks(p); err == nil {
		return resolved
	}
	return filepath.Clean(p)
}
