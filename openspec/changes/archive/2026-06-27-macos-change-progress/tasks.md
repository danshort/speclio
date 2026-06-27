## 1. Current-change progress

- [x] 1.1 `AppModel`: resolve the change behind the current selection (`.artifact` → active/archived; `.worktreeArtifact` → worktree change; else none) and expose its `ChangeProgress?`
- [x] 1.2 Reuse `progressOf` (parseTasks → done/total); nil when the change has no tasks

## 2. Persistent progress bar

- [x] 2.1 `DetailView`: render a thin progress bar pinned at the top of the detail pane (above the scrolling content), shown only when a current change with tasks exists; hidden for project specs, config, worktree metadata, and task-less changes

## 3. Per-section task progress

- [x] 3.1 `TasksView`: for each section heading, compute that section's completed/total (tasks until the next section) and render it to the right of the heading — for both the interactive and read-only paths

## 4. Sidebar progress for all changes

- [x] 4.1 Give active and archived change nodes a `progress` (like worktree-change nodes), so `SidebarRow` shows the bar + `done/total` everywhere

## 5. Spec + verification

- [x] 5.1 Delta spec: ADD requirements for persistent change progress, per-section task progress, and sidebar change progress (modifying `macos-app`)
- [x] 5.2 `swift build` green; manual QA (progress consistent across sidebar / persistent bar / Tasks; hidden for non-change views)
- [x] 5.3 Confirm no domain/golden changes (reused `parseTasks`); Go + Swift lanes unaffected
