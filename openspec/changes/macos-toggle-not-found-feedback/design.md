## Context

`toggleTaskByText` re-reads + re-parses, finds the task by text, and toggles it. On no match it `return items` (unchanged) — so the UI's `toggle` can't tell "toggled" from "not found," and shows nothing (#101). The structured edit ops (`editTaskText`/`deleteTask`/`moveTask`) already throw `TaskEditError.fileChanged` on a missing target, and the editing UI's `run` surfaces that with a notice + disk refresh. The toggle simply predates that pattern.

## Goals / Non-Goals

**Goals:** a not-found toggle produces a visible notice and a refresh, consistent with the edit ops. No change to the found/write-error paths.

**Non-Goals:** the Go TUI toggle; duplicate-text ambiguity (#115).

## Decisions

### D1 — Not-found throws `.fileChanged`
`toggleTaskByText` throws `TaskEditError.fileChanged` instead of returning the list unchanged when the guard fails (no task with the given text). This reuses the existing conflict type and the UI's existing conflict handling, so the toggle and the edit ops behave identically on a vanished target. The found path and the write-error path (`toggleTask` propagating a write failure) are untouched.

### D2 — UI surfaces it on the toggle path
`toggle` catches `TaskEditError.fileChanged` → sets a transient notice ("Couldn't find that task — it may have changed on disk; refreshed.") and calls `refreshFromDisk()`; other errors keep the existing "Couldn't write tasks.md" message. The toggle handler already runs `commitActiveEdit()` first, unchanged.

## Risks / Trade-offs

- **Test update:** `ToggleTests` "unknown task no change" asserted a silent no-op; it now asserts the thrown conflict + unchanged file. Intended.
- **No false positives:** the throw only fires when the guard already fails (truly not found), so well-formed toggles are unaffected.

## Migration Plan

macOS-only, additive behavior on a previously-silent path. No data or contract change.

## Open Questions

- None.
