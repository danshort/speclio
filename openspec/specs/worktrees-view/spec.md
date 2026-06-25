# worktrees-view Specification

## Purpose
Provides a read-only worktrees view (`ModeWorktrees`), opened with `w` from the index, that lists the current repository's git worktrees, their active changes, and per-change task progress, with navigation into a read-only artifact viewer for a foreign change. It gives a single place to see what every agent working in a sibling worktree is doing and how far along each change is.

## Requirements
### Requirement: Discover the repository's git worktrees on entry
On entering `ModeWorktrees`, lectern SHALL enumerate the git worktrees of the current repository by invoking `git worktree list --porcelain` once and SHALL load each worktree's active changes through the existing loader. When `git` is unavailable or the project is not inside a git working tree, the view SHALL render a single explanatory line instead of erroring.

#### Scenario: Worktrees are enumerated on entry
- **WHEN** the user opens the worktrees view from the index
- **THEN** lectern runs `git worktree list` once and lists every worktree of the current repository together with its loaded active changes

#### Scenario: Git unavailable degrades gracefully
- **WHEN** the user opens the worktrees view and `git` is not available or the project is not inside a git working tree
- **THEN** the view shows an explanatory line that worktrees are unavailable and does not crash

### Requirement: Display worktrees with their active changes and progress
The worktrees view SHALL group active changes by worktree. Each worktree SHALL be labelled by its branch name, or by a short HEAD SHA when its HEAD is detached. Each active change SHALL be rendered nested under its worktree with a task-progress bar showing completed and total tasks. A worktree with no active changes SHALL render as empty. Bare worktrees SHALL be omitted, and locked or prunable worktrees SHALL be annotated as such.

#### Scenario: A worktree shows its active changes with progress
- **WHEN** the worktrees view is open and a worktree has an active change with some completed tasks
- **THEN** that change is shown nested under its worktree with a progress bar reflecting its completed and total task counts

#### Scenario: A worktree with no active changes shows empty
- **WHEN** the worktrees view is open and a worktree has no active changes
- **THEN** that worktree is shown with a "(no active changes)" indication rather than being omitted

#### Scenario: A detached-HEAD worktree is labelled by SHA
- **WHEN** the worktrees view is open and a worktree's HEAD is detached
- **THEN** that worktree is labelled by a short HEAD SHA instead of a branch name

### Requirement: Identify the current worktree
The worktrees view SHALL list the current worktree first and SHALL badge it as the current worktree. The current worktree's changes are surfaced on the index already, so the view SHALL NOT offer to open them from here.

#### Scenario: Current worktree is first and badged
- **WHEN** the worktrees view is open
- **THEN** the worktree lectern is running in appears first in the list with a "(current)" badge

### Requirement: Navigate the worktrees view and return to the index
While in `ModeWorktrees`, `j` and `k` SHALL move the cursor down and up across worktrees and their changes, and `esc` SHALL return to `ModeIndex`.

#### Scenario: Navigate between items
- **WHEN** the worktrees view is open and the user presses `j`
- **THEN** the cursor moves to the next navigable item

#### Scenario: Escape returns to the index
- **WHEN** the worktrees view is open and the user presses `esc`
- **THEN** the mode returns to `ModeIndex`

### Requirement: Open a foreign change read-only
Pressing `Enter` on an active change belonging to another worktree SHALL open that change in the read-only archive viewing path with an available artifact tab selected. In that mode, task checkboxes SHALL NOT be toggleable and artifacts SHALL NOT be edited in place; opening the artifact in `$EDITOR` SHALL remain available.

#### Scenario: Enter opens a foreign change read-only
- **WHEN** the worktrees view is open, the cursor is on another worktree's active change, and the user presses `Enter`
- **THEN** the change opens in a read-only viewer with an available artifact tab shown

#### Scenario: Foreign tasks cannot be toggled
- **WHEN** a foreign change is open in the read-only viewer on its tasks artifact and the user presses the toggle key
- **THEN** no task checkbox state is changed and nothing is written to that worktree

### Requirement: Poll only while the worktrees view is active
While `ModeWorktrees` (or a foreign change opened from it) is the active mode, lectern SHALL refresh the displayed changes' content on the existing polling tick so task progress tracks live edits. The set of worktrees SHALL be captured once on entry and not re-enumerated on every tick. When the user is not in the worktrees view, lectern SHALL NOT poll other worktrees.

#### Scenario: Progress updates live while viewing
- **WHEN** the worktrees view is open and a task is completed in another worktree's change on disk
- **THEN** that change's progress bar updates on the next poll while the view remains open

#### Scenario: No cross-worktree polling outside the view
- **WHEN** the user is on the index or viewing a change in the current worktree
- **THEN** lectern does not enumerate or reload other worktrees
