package ui

import (
	"path/filepath"

	"github.com/danshort/lectern/internal/openspec"
)

// freshness caches per-path file signatures so the periodic poll can skip
// re-reading and re-parsing a change whose tasks.md hasn't changed. The
// signature itself comes from the caller (Loader.Signature); this type is just
// the compare-and-update logic, which keeps it trivially testable.
//
// Index mode and worktrees mode keep SEPARATE freshness caches (they hold
// independent in-memory copies of the current worktree's changes), so a reload
// gated for one never masks a needed reload for the other.
type freshness struct {
	sigs map[string]openspec.FileSig
}

func newFreshness() *freshness { return &freshness{sigs: map[string]openspec.FileSig{}} }

// changed records cur as the latest signature for path and reports whether it
// differs from the previously seen one. A first-seen path counts as changed.
func (f *freshness) changed(path string, cur openspec.FileSig) bool {
	prev, ok := f.sigs[path]
	f.sigs[path] = cur
	return !ok || cur != prev
}

// taskChanged reports whether ch's tasks.md changed since last seen in cache c,
// updating the cache. Gates the per-change ReloadChange in the index and
// worktrees polls so unchanged changes are not re-read/re-parsed.
func (m *Model) taskChanged(c *freshness, ch openspec.Change) bool {
	p := filepath.Join(ch.Path, openspec.FileTasks)
	return c.changed(p, m.loader.Signature(p))
}
