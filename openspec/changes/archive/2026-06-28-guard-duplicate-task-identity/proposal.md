## Why

Both task-editing layers locate a task by its identity (section prefix + number-stripped, `~~`-stripped description) and act on the **first** match. Two tasks with the same description in one section therefore alias each other: editing/deleting/toggling the second silently hits the first, and selection/focus jump (#115). OpenSpec `tasks.md` is numbered and a duplicate description within a section is a malformed file, not a supported case — so the right fix is to make the precondition explicit, treat a duplicate as a conflict instead of silently picking one, stop the editing UI from *creating* duplicates, and key the macOS task list on stable identity now that identities are unique.

## What Changes

- **Precondition documented:** task descriptions are unique within a section. Stated in the `tasks-toggle` (Go) and `macos-task-editing` capabilities and the code that relies on it.
- **Guard the write paths (treat multiple matches as a conflict):**
  - Swift `findTaskLine` (`TaskEditing.swift`) throws a new `TaskEditError.ambiguous` when more than one task in the section matches the identity; `editTaskText`/`deleteTask`/`moveTask` surface it (no write). The macOS UI shows a clear "rename to disambiguate" notice.
  - Go `ToggleTaskByText` returns a new sentinel `ErrAmbiguousTask` when the cursor's text matches more than one task; the TUI surfaces it instead of toggling the wrong line. `FindCursorByText` (used for benign cursor restore) is unchanged.
- **`performAdd` uses a unique placeholder (macOS):** instead of the literal `"New task"`, generate the first un-used `"New task"`, `"New task 2"`, … within the section, so adding several tasks never creates colliding identities.
- **Stable list identity (macOS):** key the tasks `ForEach` on the per-task stable identity (and a kind-tagged id for section rows) rather than array offset, so edit-focus and selection track the task, not the slot, across reorder/add/delete.

## Capabilities

### Modified Capabilities
- `macos-task-editing`: the safe-write identity requirement gains the uniqueness precondition and an ambiguity-is-a-conflict rule; add gains a unique-placeholder guarantee.
- `tasks-toggle`: the toggle requirement gains the same ambiguity-is-a-conflict guard for the by-text write.

## Impact

- **Code:** `macos/OpenSpecKit/Sources/OpenSpecKit/TaskEditing.swift` (ambiguity guard, `TaskEditError.ambiguous`), `macos/LecternApp/Sources/LecternApp/ContentView.swift` (unique placeholder, `ForEach` identity, surface `.ambiguous`), `internal/openspec/tasks.go` (`ErrAmbiguousTask` guard in `ToggleTaskByText`), `internal/ui/tasks.go` (surface it).
- **Tests:** OpenSpecKit (ambiguous edit → conflict; unique placeholder), Go (`ToggleTaskByText` ambiguous → error), and the existing suites stay green.
- No new dependency; no cross-language contract change.

## Non-goals

- Reworking identity to a parse-time stable token / UUID — unnecessary once duplicates are disallowed and guarded.
- Any change to the renumbering or signature-gating behavior.
