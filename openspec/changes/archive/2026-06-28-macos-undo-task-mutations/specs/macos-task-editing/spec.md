# macos-task-editing spec delta

## ADDED Requirements

### Requirement: Task mutations are undoable
Every task mutation performed in the interactive Tasks view — inline-editing a task's text, adding a task, deleting a task, and reordering or moving a task — SHALL be undoable through the window's undo manager. Undo (`⌘Z`) SHALL revert the mutation on disk and reflect it in the view; Redo (`⌘⇧Z`) SHALL re-apply it. The Edit menu SHALL name each action ("Edit Task", "Add Task", "Delete Task", "Move Task"; toggling is named "Toggle Task" per the macos-app capability).

Undo and redo SHALL restore `tasks.md` to a byte-exact prior state, so task numbers, ordering, and the file's existing line endings (LF or CRLF) are preserved exactly. Undo/redo SHALL be conflict-guarded: before restoring, the app SHALL verify that `tasks.md` on disk still matches the content the mutation produced; if another process changed the file in between, the app SHALL skip the restore (making no write), surface a transient notice, and refresh from disk rather than clobbering the external edit.

Read-only task views (a foreign worktree's tasks) SHALL NOT register undo actions, since they do not mutate.

#### Scenario: Undo a delete restores the task exactly
- **WHEN** the user deletes a task and then chooses Undo
- **THEN** the deleted task is restored to its original position with its original number and the file's line endings preserved

#### Scenario: Undo a move restores the original order
- **WHEN** the user reorders or moves a task and then chooses Undo
- **THEN** the task returns to its original section and position

#### Scenario: Redo re-applies a mutation
- **WHEN** the user undoes a task mutation and then chooses Redo
- **THEN** the mutation is re-applied on disk and reflected in the view

#### Scenario: Undo is guarded against external edits
- **WHEN** the user performs a mutation, `tasks.md` is then changed by another process, and the user chooses Undo
- **THEN** the app does not overwrite the external change; it shows a transient notice and refreshes from disk

#### Scenario: Read-only worktree tasks register no undo
- **WHEN** a foreign worktree's tasks are displayed read-only
- **THEN** no task-mutation undo actions are available for them
