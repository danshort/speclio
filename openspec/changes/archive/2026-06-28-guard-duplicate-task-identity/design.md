## Context

A code review (#115) found that task identity (`sectionPrefix` + number-stripped, `~~`-stripped description) isn't unique, and both edit layers act on the first match — so duplicate descriptions in a section alias each other. OpenSpec `tasks.md` is numbered, so a within-section duplicate is malformed; the resolution is to disallow-and-guard rather than support it. The macOS `ForEach(id: \.offset)` is the intertwined half: it keys rows by position, so once identities are unique we can key by identity and make focus/selection track tasks.

## Goals / Non-Goals

**Goals:** make the uniqueness precondition explicit; turn a duplicate match into a visible conflict (no silent first-match write) on every identity-based write path; stop the UI from creating duplicates; key the macOS list on stable identity.

**Non-Goals:** parse-time UUID tokens (unnecessary once duplicates are disallowed); any renumber/signature change.

## Decisions

### D1 — Ambiguity is a conflict, not a first-match
- **Swift:** `findTaskLine` returns the single match, throws `TaskEditError.ambiguous` on >1, returns nil on none. `editTaskText`/`deleteTask`/`moveTask` already `try` it, so they surface the conflict with no write. A new `.ambiguous` case (distinct from `.fileChanged`) lets the UI show an accurate "rename to disambiguate" message.
- **Go:** `ToggleTaskByText` counts matches for the cursor's text; >1 → return a sentinel `ErrAmbiguousTask` (no write). `FindCursorByText` is left first-match because it's also used for benign cursor restoration after reload, where first-match is the desired, non-destructive behavior.

### D2 — Unique add placeholder (macOS)
`performAdd` computes the first un-used name in the sequence `New task`, `New task 2`, `New task 3`, … among the target section's existing task descriptions, so repeated adds never collide. The new task is still selected and opened for inline editing.

### D3 — Stable list identity (macOS)
Replace `ForEach(Array(items.enumerated()), id: \.offset)` with rows carrying an explicit, kind-tagged stable id: `"t\u{1}" + sectionPrefix + "\u{1}" + taskDescription` for tasks and `"s\u{1}" + section text` for section headers. With task descriptions unique per section (D1/D2) these ids are unique, so SwiftUI tracks each row to its task — edit-focus and selection no longer follow the slot across reorder/add/delete. The existing string identity used for selection/drag/edit is reused, so the two identity systems converge.

## Risks / Trade-offs

- **`.ambiguous` only fires on malformed files** → that's intended; it converts a silent corruption into a visible, actionable message.
- **Section-header id collisions** → two headers with identical text are themselves malformed; tagging by kind plus the prefix makes practical collisions impossible.
- **Go sentinel vs Swift enum** → different idioms per language, but both mean "ambiguous: refuse and surface."

## Migration Plan

Additive guards + a UI identity change; no data migration, no contract change. Well-formed files (the norm) behave exactly as before. Rollback is reverting the change.

## Open Questions

- None; #115's chosen resolution is fully specified.
