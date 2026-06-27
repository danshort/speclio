## Context

The pieces already exist: `OpenSpecKit.Loader.loadFrom(path)` returns a worktree's active changes, `parseTasks(content)` yields done/total for progress, and the Worktrees mode already renders worktree metadata in the `OutlineGroup` sidebar. This change is mostly a UI expansion reusing the domain layer — no new parsing behavior, so the cross-language golden corpus is untouched.

This mirrors the TUI (`internal/ui/worktrees.go`): discover worktrees → `LoadFrom(wt.Path)` per non-bare worktree → render each with its active changes + task progress, current-first, opening a foreign change read-only.

## Decisions

- **Reuse the domain layer, no new contract surface.** Survey = `loadFrom(worktree.path)` per worktree; progress = `parseTasks(change.tasks.content)` → `done/total`. Nothing new in `OpenSpecKit` to golden.
- **Tree shape.** Each worktree `SidebarNode` gains children = its active changes; each change node's subtitle shows `n/total`. Fits the existing `OutlineGroup` (which already fixed the disclosure glitches). Selecting a change node resolves to a read-only artifact view.
- **Selection.** Add `Selection.worktreeChange(worktreePath, changeName)` so the detail can resolve a foreign change by (path, name) without colliding with the current project's `.artifact`.
- **Survey timing.** Survey on entering the Worktrees mode and on Reload — not live-watched per worktree (the FSEvents watcher covers only the opened project's `openspec/`). Per-worktree live progress is deferred (note it; the TUI polls).
- **Read-only.** Foreign-worktree changes render their artifacts but tasks are not toggleable — writes stay scoped to the opened project, which keeps the security-scoped-bookmark/file-access story simple and avoids accidental cross-worktree edits.
- **Graceful degradation.** Skip bare worktrees; a worktree without `openspec/` lists no children (not an error); current worktree first and marked.

## Risks / Trade-offs

- **[Low] Cost.** `loadFrom` per worktree on enter/refresh; cheap for typical worktree counts. If it ever matters, survey lazily on expand.
- **[Low] Staleness.** Other worktrees' progress updates only on Reload in v1 (documented; live polling is a later enhancement).
- **[Low] Read-only asymmetry.** Tasks are interactive in the current project but not under Worktrees — intentional; signalled by the read-only rendering (no checkboxes).
