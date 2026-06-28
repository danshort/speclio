# macos-task-editing Specification

## Purpose
Structured editing of `tasks.md` from the macOS app's Tasks view — add, delete, drag-reorder (within and across sections), and inline text-edit of tasks — with positional renumbering (recompute ordinals, preserve section prefixes including non-integer ones like `3b`), number-stripped and strikethrough-tolerant task identity, surgical re-read-before-write persistence that leaves untouched content byte-for-byte intact, and conflict-abort when the file changed underneath the edit. Cross-worktree changes remain read-only.

## Requirements

### Requirement: Structured task editing is available for editable changes
The macOS tasks view SHALL expose add, delete, reorder, and inline-edit controls for a change in the current worktree, revealed per row on hover or selection. For read-only (cross-worktree) changes these controls SHALL be disabled, matching the existing checkbox-toggle behavior. Editing targets the OpenSpec single-line numbered task format (`## <prefix>.` headings with `- [ ] <prefix>.<ordinal> <text>` tasks); nested/indented and unnumbered tasks are out of scope.

#### Scenario: Editing enabled for a local numbered change
- **WHEN** the user views the tasks of a change in the current worktree whose `tasks.md` uses numbered sections
- **THEN** the add, delete, reorder, and inline-edit affordances are available

#### Scenario: Editing disabled for a foreign worktree
- **WHEN** the user views the tasks of a change opened from another worktree (read-only)
- **THEN** the editing affordances are disabled and only viewing is possible

### Requirement: Add a task after the selected task
The system SHALL insert a new pending task (`- [ ]`) immediately after the currently selected task, within the same section, and SHALL assign it the next sequential ordinal so the section's ordinals remain contiguous.

#### Scenario: Insert into the middle of a section
- **WHEN** the selected task is `1.2` in a section containing `1.1`, `1.2`, `1.3`
- **THEN** a new task is inserted as `1.3` and the former `1.3` becomes `1.4`

#### Scenario: Insert after the last task in a section
- **WHEN** the selected task is the last task `2.4` of section 2
- **THEN** a new task is appended as `2.5`

### Requirement: Delete a task with confirmation
The system SHALL remove a task only after an explicit confirmation step, and SHALL renumber the remaining tasks in that section so ordinals stay contiguous with no gaps.

#### Scenario: Confirmed delete renumbers the tail
- **WHEN** the user deletes `1.2` from a section containing `1.1`, `1.2`, `1.3` and confirms
- **THEN** `1.2` is removed, the former `1.3` becomes `1.2`, and the file is written

#### Scenario: Cancelled delete makes no change
- **WHEN** the user initiates a delete but cancels the confirmation
- **THEN** the task remains and the file is not modified

### Requirement: Reorder tasks within a section
The system SHALL let the user drag a task to a new position within its section and SHALL recompute the ordinals of that section so they reflect the new visual order.

#### Scenario: Drag a task upward
- **WHEN** the user drags `1.3` above `1.1` in a section ordered `1.1`, `1.2`, `1.3`
- **THEN** the dragged task becomes `1.1`, the former `1.1` becomes `1.2`, and the former `1.2` becomes `1.3`

### Requirement: Move a task across sections
The system SHALL let the user drag a task into a different section. The moved task SHALL adopt the destination section's prefix, and the ordinals of both the source and destination sections SHALL be recomputed to remain contiguous.

#### Scenario: Drag a task from section 1 into section 3
- **WHEN** the user drags `1.2` into section `3` (which had `3.1`, `3.2`) at the end
- **THEN** the moved task becomes `3.3`, section 1's remaining tasks renumber to close the gap, and section 3 stays contiguous

#### Scenario: Destination prefix is preserved verbatim
- **WHEN** the user drags a task into a section whose heading is `## 3b. …`
- **THEN** the moved task adopts the `3b` prefix (e.g. becomes `3b.<n>`), not a normalized integer prefix

#### Scenario: Drop at the end of a section
- **WHEN** the user drops a task on a section's end-of-section drop zone
- **THEN** the task is appended as the last task of that section, and both affected sections are renumbered to stay contiguous

### Requirement: Inline-edit task text preserves the number
The system SHALL let the user edit a task's description text in place using a multi-line (wrapping) editor. The edit SHALL change only the description; the task's `- [ ]`/`- [x]` state and its `<prefix>.<ordinal>` number SHALL be preserved. The task SHALL remain single-line: any newline entered in the editor SHALL be collapsed to a single space on save.

#### Scenario: Fix a word in a task
- **WHEN** the user edits the text of `2.1` from "Implment API" to "Implement API"
- **THEN** the line becomes `- [ ] 2.1 Implement API` with the checkbox state and number unchanged

#### Scenario: Multi-line input is flattened
- **WHEN** the user enters text containing line breaks in the editor and saves
- **THEN** the task is written as a single line, with each newline replaced by a space

### Requirement: Inline editor commit and cancel
The inline editor SHALL commit the edit on Cmd-Return or when focus leaves the field (clicking another task, toggling a task, starting to edit a different task, or clicking empty space), and SHALL cancel without saving on Esc.

#### Scenario: Commit on Cmd-Return
- **WHEN** the user presses Cmd-Return while editing a task
- **THEN** the edited text is saved

#### Scenario: Commit on click-away
- **WHEN** the user clicks another task or empty space while editing
- **THEN** the in-progress edit is saved

#### Scenario: Cancel on Escape
- **WHEN** the user presses Esc while editing
- **THEN** the editor closes and the task text is left unchanged

### Requirement: Positional renumbering recomputes ordinals and preserves section prefixes
On any structural edit (add, delete, reorder, cross-section move) the system SHALL recompute task **ordinals** sequentially from 1 within each affected section, and SHALL preserve each section's **prefix** exactly as it appears in the `##` heading, including non-integer prefixes such as `3b`. The system SHALL NOT generate, reorder, or renumber sections.

#### Scenario: Ordinals recomputed, prefix untouched
- **WHEN** tasks in a section headed `## 3b.` are reordered
- **THEN** their numbers become `3b.1`, `3b.2`, `3b.3`, … and the `## 3b.` heading is left unchanged

### Requirement: Task identity for safe writes ignores the number and tolerates strikethrough
When locating a task on disk before writing, the system SHALL match on the task's description text with the leading `<prefix>.<ordinal>` number removed, so that renumbering does not break the match. Matching SHALL also tolerate descriptions wrapped in `~~…~~` (skipped tasks).

#### Scenario: Match survives renumbering
- **WHEN** a task displayed as `1.3 Add frontend component` is to be modified but the file now numbers it `1.2`
- **THEN** the task is still located by the description "Add frontend component"

#### Scenario: Match a strikethrough task
- **WHEN** the on-disk task is `- [ ] ~~6.1 Drop the cache~~ (skipped)`
- **THEN** the task is located by the description "Drop the cache (skipped)" — the leading `<prefix>.<ordinal>` number and the `~~` strikethrough markers are removed, while text outside the markers is kept

### Requirement: Edits are surgical and re-read the file before writing
Every structural edit SHALL re-read `tasks.md` immediately before writing, derive the change against the current file contents, and rewrite only the minimal span of lines affected (the touched section, or both sections for a cross-section move), leaving all other lines — including prose between tasks — byte-for-byte unchanged.

#### Scenario: Surrounding content is preserved
- **WHEN** a task in section 2 is deleted
- **THEN** only section 2's task lines are rewritten and all other sections and any interspersed prose are unchanged

### Requirement: Abort the edit with a visible notice when the file changed underneath
If, on the pre-write re-read, the targeted task can no longer be located or the relevant section structure has changed since the edit began, the system SHALL NOT write, SHALL re-read and refresh the displayed tasks, and SHALL surface a visible notice that the file changed on disk.

#### Scenario: Concurrent agent edit detected
- **WHEN** the user reorders tasks but an agent rewrote the section between render and save such that the dragged task is no longer found
- **THEN** no write occurs, the view refreshes from the current file, and a notice informs the user the file changed on disk
