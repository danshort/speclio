## 1. Survey data

- [x] 1.1 In `AppModel`, on entering Worktrees mode / on reload, survey each non-bare worktree via `Loader.loadFrom(worktree.path)` → its active changes (current worktree first); tolerate worktrees without `openspec/` (no changes)
- [x] 1.2 Compute per-change task progress from `parseTasks(change.tasks.content)` (done / total)

## 2. Sidebar tree

- [x] 2.1 Give each worktree `SidebarNode` children = its active changes; change node subtitle shows `done/total`
- [x] 2.2 Add `Selection.worktreeChange(worktreePath, changeName)`; map change nodes to it

## 3. Detail

- [x] 3.1 Render a selected worktree change **read-only** (reuse artifact rendering; no task toggling); show its progress
- [x] 3.2 Worktree node detail still shows metadata (path/branch/HEAD/flags)

## 4. Robustness

- [x] 4.1 Refresh re-surveys; bare worktrees skipped; missing-git still shows the graceful unavailable state
- [x] 4.2 Confirm the OutlineGroup expand/collapse stays glitch-free with the deeper tree

## 5. Spec + verification

- [x] 5.1 Delta spec: MODIFIED `macos-app` "Worktrees overview" requirement (per-worktree changes + progress + read-only open)
- [x] 5.2 `swift build` green; manual QA against the sample worktrees (current + the lectern-wt-* set)
- [x] 5.3 Confirm no domain/golden changes (reused `loadFrom`/`parseTasks`); Go + Swift lanes unaffected

## 6. Parity polish (folded in from #62)

- [x] 6.1 Graphical progress bar on worktree change rows in the sidebar (alongside/instead of the textual done/total)
- [x] 6.2 "(no active changes)" affordance for a non-bare worktree that has an `openspec/` project but zero active changes (distinct from bare / no-project)
- [x] 6.3 Surface worktree state in the sidebar label: `locked`/`prunable` flags, and a short HEAD SHA for detached worktrees
- [x] 6.4 `swift build` green; manual QA against the sample worktrees
