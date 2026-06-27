## Why

The TUI surfaces change progress pervasively; the macOS app only shows a progress bar at the top of the Tasks view. Progress should be visible (a) for the **current change regardless of which artifact** you're viewing, (b) **per Tasks section**, and (c) in the **sidebar for every change** — we already do the sidebar treatment for worktree changes, but not for active/archived ones (#65).

## What Changes

- **Persistent change progress bar:** a thin bar across the top of the detail content area (below the toolbar/mode switcher) showing the current change's overall task completion, visible no matter which of its artifacts is open. Shown only for change-backed selections (active, archived, or worktree change); hidden for project specs, config, and worktree metadata.
- **Per-section task progress:** each section heading in the Tasks view shows that section's completed/total to its right (interactive and read-only Tasks alike).
- **Sidebar change progress everywhere:** active and archived change rows show a progress bar + `done/total`, matching what worktree-change rows already show.

## Non-goals

- Progress for project specs or config (they aren't changes).
- Changing how progress is computed — reuse `OpenSpecKit.parseTasks`. No new domain behavior, so no corpus/golden change.

## Capabilities

### Modified Capabilities

- `macos-app`: adds persistent per-change progress in the detail pane, per-section task progress, and sidebar progress for all changes.

## Impact

- `macos/LecternApp` — `AppModel` (resolve the current change for any selection; progress on active/archived change nodes); `ContentView` (persistent bar in the detail pane; per-section progress in `TasksView`).
- No `OpenSpecKit` / `internal/openspec` / corpus changes.
