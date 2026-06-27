## Why

The macOS app's Worktrees mode shows only git metadata (path, branch, HEAD, flags). The TUI does more — it surveys every worktree with its **active changes and live task progress**, which is the actual value of the view: seeing work-in-flight across worktrees at a glance. This change brings the app to parity (the deferred enhancement from the `macos-app` change, tracked in #62).

## What Changes

- In the **Worktrees** mode, each non-bare worktree becomes expandable to list its **active changes** (loaded from that worktree's `openspec/changes/`), each with a **task-progress** indicator (done / total).
- Selecting a change under a worktree opens it **read-only** in the detail pane (its artifacts render like any change).
- The **current** worktree is listed first and marked; bare worktrees and worktrees without an `openspec/` project degrade gracefully (no children, not an error).
- **Refresh** re-surveys the worktrees.

## Non-goals

- Toggling/editing tasks in foreign worktrees — they are read-only here (writes stay scoped to the opened project).
- Live FSEvents-watching of every worktree — v1 surveys on entering the mode and on refresh; per-worktree live progress is a later enhancement.

## Capabilities

### Modified Capabilities

- `macos-app`: the "Worktrees overview" requirement gains per-worktree active changes + task progress and read-only open.

## Impact

- `macos/LecternApp` — Worktrees mode: sidebar worktree nodes gain change children (with progress subtitles); detail renders a selected foreign change read-only; a new selection case for worktree changes.
- `macos/OpenSpecKit` — reuses existing `Loader.loadFrom` + `parseTasks`; no new domain behavior, so **no corpus/golden change**.
- No Go / TUI changes.
