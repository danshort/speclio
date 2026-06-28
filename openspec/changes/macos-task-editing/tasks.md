## 1. Model & parser (OpenSpecKit)

- [ ] 1.1 Extend `TaskItem` in `Models.swift` with structured fields: `sectionPrefix` (verbatim from the heading, e.g. `3`/`3b`), `ordinal`, and a bare `description` (number- and `~~`-stripped), keeping `done` and `lineNum`
- [ ] 1.2 Update `parseTasks` in `Tasks.swift` to populate the new fields: extract `<prefix>.<ordinal>` from each task line, associate each task with its current section prefix, and store the stripped `description`
- [ ] 1.3 Add a strikethrough-tolerant description extractor (handles `- [ ] ~~6.1 text~~ (skipped)`) and unit-test it against numbered, strikethrough, and letter-prefix (`3b.1`) lines
- [ ] 1.4 Refine identity matching (generalize `findCursorByText`) to match on the stripped `description`, disambiguating duplicates by section + nearest position; unit-test that a match survives renumbering

## 2. Renumber engine (OpenSpecKit)

- [ ] 2.1 Implement a pure `renumberSection` helper that, given a section's prefix and an ordered list of its task lines, rewrites each line's number to `<prefix>.<sequentialOrdinal>` while preserving checkbox state, `~~` markers, and trailing text
- [ ] 2.2 Add a surgical splice writer that re-reads the file, locates the affected section line-span(s), replaces only those lines, and preserves existing line endings (reuse the toggle's deliberate CRLF handling)
- [ ] 2.3 Add conflict detection: if the targeted task can't be located or the affected section structure changed on the pre-write re-read, return a typed "file changed" result and do not write
- [ ] 2.4 Unit-test renumber + splice: middle-insert, tail-delete, in-section reorder, `3b` prefix preservation, and "surrounding sections/prose untouched"

## 3. Edit operations (OpenSpecKit)

- [ ] 3.1 Implement `addTask(after:)` — insert a pending task after the given task (or at end of its section as fallback), then renumber that section
- [ ] 3.2 Implement `deleteTask` — remove the task, then renumber that section
- [ ] 3.3 Implement `reorderTask(within:)` — move a task to a new index in its section, then renumber that section
- [ ] 3.4 Implement `moveTask(toSection:at:)` — adopt the destination section's prefix, insert at the target index, then renumber both source and destination sections
- [ ] 3.5 Implement `editTaskText` — replace only the description, preserving checkbox state and `<prefix>.<ordinal>` number
- [ ] 3.6 Ensure every operation in this group re-reads before write and routes through the conflict-detection path (2.3)
- [ ] 3.7 Add golden tests for each operation covering the spec scenarios (cross-section move, strikethrough, conflict-abort)

## 4. macOS tasks view (LecternApp)

- [ ] 4.1 Gate all editing controls on editability (disable for foreign-worktree/read-only changes, matching the existing toggle gating)
- [ ] 4.2 Add the `+` (add after selected) and `−` (delete) controls wired to the OpenSpecKit operations
- [ ] 4.3 Add a delete confirmation step (no Undo until #103)
- [ ] 4.4 Add inline text editing: double-click a task's text to an editable field that saves via `editTaskText` on commit and cancels on escape
- [ ] 4.5 Add drag-to-reorder with a drag handle, including cross-section drops, with a visually explicit destination section/insertion point during the drag
- [ ] 4.6 On a conflict result, refresh the displayed tasks from the current file and surface a visible "file changed on disk" notice

## 5. Verification

- [ ] 5.1 `swift test` passes for OpenSpecKit (parser, identity, renumber, operations, conflict)
- [ ] 5.2 Manual check: add/delete/reorder/cross-section-move/inline-edit on a real change, confirming numbers stay contiguous and prefixes (incl. `3b`) are preserved
- [ ] 5.3 Manual check: edit a task while an external process rewrites `tasks.md`, confirming the edit aborts with the notice and no corruption
- [ ] 5.4 Manual check: confirm editing controls are absent/disabled on a foreign-worktree change
- [ ] 5.5 Update any affected docs/README notes for the macOS app's editing capability
