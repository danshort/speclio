## Why

Two TUI behaviors silently do the wrong thing, eroding trust in a tool people leave running while agents edit the same files:

- **Toggling the wrong task (#91).** `ToggleTask` re-reads `tasks.md` but flips the line at the caller's *stale* `LineNum`. If the file shifted between render and keypress (an agent appended a section, the user edited in `$EDITOR`), `Space` silently toggles the wrong checkbox. The macOS port already fixed this (`toggleTaskByText`); the Go TUI never got the fix.
- **Invisible reload failures (#92).** The index/normal-mode poll loop `return nil` on disk errors, so a mid-session read failure just stops updating with no indication. A malformed `.openspec.yaml` is silently discarded (`loader.go:295`), leaving fields blank with no hint why.

## What Changes

- Add `ToggleTaskByText(path, text)` that re-reads, re-parses with `ParseTasks`, locates the task by its rendered `Text` (via `FindCursorByText`), and toggles that fresh line — so a shifted file can never flip the wrong checkbox. The tasks view calls it with the cursor task's `Text` instead of `(items, idx)`. CRLF endings are preserved (write path splits on raw `\n`). Matching on full `Text` is correct because the TUI never renumbers.
- Surface reload/poll errors through the existing `m.errMsg` status line instead of returning `nil` silently (the `ListChangeNamesFrom`/`ListArchiveNamesFrom`/`ListSpecNamesFrom` reads and the swallowed `LoadFrom` in `pollIndexMode` and `pollNormalModeChanges`).

## Capabilities

### New Capabilities
<!-- none -->

### Modified Capabilities
- `tasks-toggle`: the "Toggle checkbox with Space" requirement gains a safe-write guarantee — the on-disk line to flip is re-derived by matching the task text against a fresh re-parse, not a stale line index.
- `change-index`: the "Real-time index updates" requirement gains an error-surfacing guarantee — a reload/poll error is shown in the status line rather than silently swallowed.

## Impact

- **Code:** `internal/openspec/tasks.go` (add `ToggleTaskByText` + package wrapper), `internal/ui/tasks.go` (call the by-text entry point), `internal/ui/index.go` (surface poll errors via `m.errMsg`).
- **Tests:** `internal/openspec/tasks_test.go` (toggle-after-shift hits the right task; CRLF preserved; unknown text no-ops).
- **No change** to the macOS app (its `toggleTaskByText` already behaves this way) or to the 500 ms polling cadence.

## Non-goals

- Replacing 500 ms polling with filesystem watching — that is **#90**.
- Atomic temp-file-and-rename writes — separate hardening, not required to fix the wrong-line bug.
- Number-stripped task identity — only needed where renumbering occurs (the macOS editor, #97); the TUI matches on full text.
- Surfacing a malformed `.openspec.yaml`: the loader deliberately treats it as optional/non-fatal (only the optional `Created` date is affected). Surfacing it would risk a per-tick error message or a golden-contract change to `Change`; left as documented.
