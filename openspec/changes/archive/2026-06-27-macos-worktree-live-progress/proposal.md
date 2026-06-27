## Why

The Worktrees mode surveys each worktree's changes only on enter, on manual reload, or when the *opened* project's FSEvents fire — the watcher covers only the opened project's `openspec/`, not the other worktrees. So when an agent edits a foreign worktree's `tasks.md`, its progress doesn't move until you reload. The TUI polls (`pollWorktrees`) so the bars track live. This is the last open item for worktree/TUI parity (#62) — the survey and all the sidebar polish (graphical progress bar, "no active changes" affordance, locked/prunable + detached-SHA labels) already shipped.

## What Changes

- **Poll worktree changes while in Worktrees mode.** On a timer, re-read the already-surveyed worktree changes from disk (via `OpenSpecKit.reloadChange`, which re-reads artifacts only — no `git` re-enumeration), mirroring the TUI's `pollWorktrees`. When a change's content differs, update it so the sidebar progress bars and any open read-only worktree artifact reflect it live.
- **Scoped and cheap.** Polling runs only while the Worktrees mode is active and stops otherwise (and on window close); the sidebar is rebuilt only when something actually changed, preserving the current selection. The worktree *set* is not re-enumerated on the tick (no `git` subprocess) — that still happens on enter/reload.

## Non-goals

- Re-enumerating worktrees (add/remove) on the tick — that stays on enter / manual reload, matching the TUI.
- Live updates outside Worktrees mode, or FSEvents watchers per worktree — polling matches the TUI and avoids N-watcher lifecycle complexity.
- Any change to the opened project's existing FSEvents live reload.

## Capabilities

### Modified Capabilities

- `macos-app`: adds live (polled) task-progress updates for worktree changes while in Worktrees mode.

## Impact

- `macos/LecternApp` — `AppModel` only (a poll timer gated on Worktrees mode + a `reloadChange`-based diff that rebuilds the sidebar on change). No `OpenSpecKit` / `internal/openspec` / corpus changes (`reloadChange` already exists).
