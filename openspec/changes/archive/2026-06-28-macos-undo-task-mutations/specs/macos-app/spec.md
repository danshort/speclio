# macos-app spec delta

## MODIFIED Requirements

### Requirement: Task toggling that preserves line endings
The app SHALL let the user toggle a task checkbox in `tasks.md`, writing the change to disk while preserving the file's existing line endings. Because the app reloads on disk changes, it SHALL re-read and re-parse `tasks.md` immediately before writing so a stale line index cannot toggle the wrong line. If the toggled task can no longer be found in the re-read file (it was changed or removed externally), the app SHALL surface a transient notice and refresh from disk rather than silently doing nothing. Each toggle SHALL be undoable: it SHALL register an undo action with the window's undo manager so the user can revert the change with Undo (`⌘Z`) and re-apply it with Redo (`⌘⇧Z`), as specified by "Task mutations are undoable" in the macos-task-editing capability. The Edit menu SHALL name the action "Toggle Task".

#### Scenario: Toggle a task
- **WHEN** the user toggles a task in the app
- **THEN** the corresponding `- [ ]`/`- [x]` marker is flipped in `tasks.md` and the file's original line endings (LF or CRLF) are preserved

#### Scenario: File changed on disk before toggle
- **WHEN** `tasks.md` was modified by another process after it was rendered and the user then toggles a task
- **THEN** the app re-reads the current file before writing, so the intended task is toggled rather than a stale line

#### Scenario: Toggled task no longer present
- **WHEN** the toggled task's text no longer exists in `tasks.md` at the moment of toggling
- **THEN** no write occurs, the app shows a transient notice that the task could not be found, and it refreshes from disk

#### Scenario: Undo a toggle
- **WHEN** the user toggles a task and then chooses Undo (`⌘Z`)
- **THEN** the task's marker is flipped back to its prior state on disk and the change is reflected in the view, and the Edit menu labels the action "Toggle Task"
