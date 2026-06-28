## MODIFIED Requirements

### Requirement: Toggle checkbox with Space
When pressing `Space` on a task item, the system SHALL:
1. Invert the `done` state of the item in memory
2. Re-read `tasks.md` and re-parse it, then locate the line to modify by matching the task's text against the fresh parse — never trusting a line index captured at render time
3. Modify only that line on disk, changing `[ ]` to `[x]` or vice versa, preserving the file's existing line endings
4. Update the render without reloading the entire file

If no task with the matching text is found in the fresh parse, the system SHALL make no change to the file.

#### Scenario: Mark task as completed
- **WHEN** the cursor is on `- [ ] Create structure` and the user presses `Space`
- **THEN** the line on disk becomes `- [x] Create structure` and the item shows the completed state

#### Scenario: Unmark completed task
- **WHEN** the cursor is on `- [x] Create structure` and the user presses `Space`
- **THEN** the line on disk becomes `- [ ] Create structure` and the item shows the pending state

#### Scenario: File shifted since render toggles the correct task
- **WHEN** the cursor is on `- [ ] 1.2 beta` but `tasks.md` has since gained lines above it (e.g. an agent inserted a section) so the rendered line index is now stale
- **THEN** pressing `Space` toggles the line whose text matches `1.2 beta`, not whatever line now sits at the old index

#### Scenario: Matching task no longer present
- **WHEN** the cursor task's text no longer exists in `tasks.md` at the moment of toggling
- **THEN** no write occurs and the file is left unchanged

#### Scenario: Write error
- **WHEN** `tasks.md` does not have write permissions and the user presses `Space`
- **THEN** the toggle is not applied and the TUI shows a temporary error message
