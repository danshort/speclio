# Tasks

## 1. Poll worktree changes
- [x] 1.1 Add `pollWorktreeChanges()` that reloads each surveyed worktree change via `OpenSpecKit.reloadChange`, compares to the current value, and (only if any changed) replaces `worktreeChanges` and rebuilds the sidebar, preserving `selectedNodeID`
- [x] 1.2 Do not re-enumerate worktrees on the tick (no `git`) — set changes stay on enter/reload

## 2. Timer gated on Worktrees mode
- [x] 2.1 Start a repeating timer (~1.5 s) when `mode` becomes `.worktrees`; invalidate it when leaving the mode and in `teardown()`
- [x] 2.2 Ensure the timer is main-actor and holds a weak self

## 3. Verify
- [x] 3.1 `swift build` the LecternApp package cleanly
- [x] 3.2 Manual: in Worktrees mode, toggle a task in another worktree's `tasks.md` externally → its sidebar progress (and open read-only artifact) updates within ~1.5 s without reload
- [x] 3.3 Manual: leaving Worktrees mode / closing the window stops polling; selection is preserved across a live update
