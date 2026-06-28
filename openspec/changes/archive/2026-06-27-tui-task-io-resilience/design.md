## Context

Two TUI behaviors fail silently:

- `Loader.ToggleTask(path, items, idx)` (`internal/openspec/tasks.go`) re-reads `tasks.md` but edits the line at `items[idx].LineNum` — a line index captured when the view last rendered. If the file shifted since (agent/editor edit), it flips the wrong line. The macOS port already solved this with `toggleTaskByText` (re-read → re-parse → `findCursorByText` → toggle the fresh line); the Go side has `FindCursorByText` but the toggle never uses it.
- The `ModeIndex`/normal-mode poll helpers (`pollIndexMode`, `pollNormalModeChanges` in `internal/ui/index.go`) `return nil` on disk read errors and swallow a failed `LoadFrom` (`if … err == nil`), so a mid-session failure is invisible. The UI already has an `m.errMsg` status line used elsewhere.

## Goals / Non-Goals

**Goals:**
- A `Space` toggle always flips the task the cursor is on, even if the file shifted since render (or no-ops if that task is gone).
- Reload/poll errors and malformed change configs are visible in the status line.
- Preserve CRLF endings and existing behavior for the happy path.

**Non-Goals:**
- fsnotify (replacing 500 ms polling) — #90.
- Atomic temp-file writes.
- Number-stripped identity — not needed; the TUI never renumbers, so full-text matching is exact.

## Decisions

### D1 — Add `ToggleTaskByText`, mirroring the macOS fix
Add `(l *Loader) ToggleTaskByText(path, text string) error` (and a package wrapper) that: re-reads the file, `ParseTasks` the fresh content, `FindCursorByText` for `text`, and — only if a task with that exact text is found — flips that fresh line (raw `\n` split to preserve CRLF) and writes. The tasks view (`internal/ui/tasks.go`) calls it with the cursor task's `Text` instead of `(items, idx)`.
- *Why full text, not number-stripped identity:* the TUI never renumbers tasks, so the rendered `Text` is a stable key; this matches the existing Swift `toggleTaskByText`. Keeping it text-based also keeps the golden/`TaskItem` contract untouched.
- *Disposition of the old `ToggleTask(items, idx)`:* retained for now (still covered by tests) but no longer on the UI path; the by-text entry point is the safe one.

### D2 — Surface poll/reload errors via `m.errMsg`
In the poll helpers, on a disk read error (the `ListChangeNamesFrom`/`ListArchiveNamesFrom`/`ListSpecNamesFrom` reads) and on a swallowed `LoadFrom` failure, set `m.errMsg` to a descriptive message instead of returning `nil` with no trace. The tick still returns without applying a partial reload (current data is retained), but the failure is now visible. `m.errMsg` is already cleared on the next keypress (`update.go:120`).

### D3 — Leave malformed `.openspec.yaml` as documented (non-goal)
`loadChangeFromDir` deliberately ignores the change-meta parse error (only the optional `Created` date is affected). Surfacing it would need either a new field on `Change` (which would break the cross-language golden contract for `tasks`/`Change`) or a per-tick `errMsg`. Not worth it; left as-is. See proposal Non-goals.

## Risks / Trade-offs

- **A persistent disk error re-sets `errMsg` every tick (≤500 ms)** → acceptable: it reflects current reality and clears on keypress. We do not spam logs or stack messages.
- **Behavior change for `tasks.md` with two identical task lines** → `FindCursorByText` matches the first; this matches the existing cursor-restore behavior (`Restore cursor by text after reload`) and the macOS toggle, so it is consistent, not new.
- **Retaining the old `ToggleTask`** → minor dead-path risk; mitigated by routing the UI through the by-text version and keeping tests on both.

## Migration Plan

Additive and Go-TUI-only. The happy path (cursor on the right line) is unchanged byte-for-byte. No macOS impact (already correct), no change to polling cadence or specs beyond the two modified requirements.

## Open Questions

- None outstanding; both fixes have a clear, tested shape.
