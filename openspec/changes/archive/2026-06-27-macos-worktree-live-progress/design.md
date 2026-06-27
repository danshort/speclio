## Context

`AppModel.loadWorktrees(path)` enumerates worktrees with `ProcessGitService` and surveys each non-bare worktree's changes via `loadFrom(wt.path)` into `worktreeChanges`. It's called from `refreshData`, which fires on enter, manual reload, and the opened project's FSEvents. Foreign worktrees aren't watched, so their progress is stale until a reload. The TUI's `pollWorktrees` re-reads the loaded changes' content on a 500 ms tick (without re-enumerating the worktree set); `OpenSpecKit.reloadChange` is the direct Swift equivalent and re-reads a change's artifacts by path (disk only, no `git`).

## Decisions

### Poll, mirroring the TUI — not per-worktree FSEvents

A repeating timer re-reads the already-surveyed worktree changes and updates them when their content changed. Chosen over FSEvents watchers on each worktree's `openspec/` because it matches the TUI, avoids the lifecycle complexity of a watcher set that changes as worktrees come and go, and the work is cheap (artifact reads for a handful of changes).

### Gated on the Worktrees mode

The timer starts when `mode` becomes `.worktrees` and is invalidated otherwise and on `teardown()` (window close) — cross-worktree progress is only visible in that mode, so there's no reason to poll elsewhere. Started/stopped from `mode`'s `didSet` (and started in `load` if the window opens directly into Worktrees mode is not a case today, since Active is the default — but the didSet covers any later default change).

### Re-read content, not the worktree set

`pollWorktreeChanges()` walks `worktreeChanges`, calls `reloadChange` per change, and compares the result to the current value. The worktree *set* (`git worktree list`) is not re-run on the tick — add/remove still happens via `loadWorktrees` on enter/reload, exactly as the TUI captures the set on entry. This keeps the tick `git`-free.

### Rebuild only on change, preserve selection

If any change differs, replace `worktreeChanges` and rebuild the sidebar, restoring `selectedNodeID` when it still resolves (the same preserve-selection pattern `refreshData` uses). Because the detail pane resolves a `.worktreeArtifact` selection *through* `worktreeChanges`, refreshing that dictionary also updates the open read-only artifact and the persistent progress bar — covering the TUI's `pollWorktreeChange` in the same pass. When nothing changed, do nothing (no view churn).

### Interval

~1.5 s — responsive enough to "track agents live" without the 500 ms TUI cadence's overhead in a desktop app. Single timer, main-actor, comparing `Change` values (`Equatable`).

## Risks / Trade-offs

- **Polling cost.** Re-reading artifacts every ~1.5 s while in Worktrees mode. Bounded (a few worktrees × few changes, disk-only) and only while that mode is foregrounded; stops otherwise.
- **Selection churn.** Mitigated by rebuilding only on actual change and restoring the selection by id.
- **Not event-exact.** Up to one interval of latency, and no live add/remove of worktrees — both deliberate, matching the TUI.
- **Background polling.** The timer is mode-gated but not app-active-gated; pausing when the app is inactive is a possible later refinement (kept out to avoid threading `scenePhase` per window).
