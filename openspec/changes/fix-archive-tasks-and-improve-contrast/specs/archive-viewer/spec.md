## MODIFIED Requirements

### Requirement: Read-only in archive mode
In `ViewingArchive` mode, keys `e` (open editor) and `Space` (task toggle) SHALL be silently ignored. The Tasks tab SHALL be rendered as read-only content: `j`/`k` and the down/up arrows SHALL scroll the viewport and SHALL NOT move a task cursor.

#### Scenario: 'e' ignored in archive mode
- **WHEN** the mode is `ViewingArchive` and the user presses `e`
- **THEN** no editor is opened and the state does not change

#### Scenario: 'Space' ignored in archive mode
- **WHEN** the mode is `ViewingArchive` and the user presses `Space`
- **THEN** no task changes its state

#### Scenario: Arrow keys scroll the Tasks tab in archive mode
- **WHEN** the mode is `ViewingArchive`, the active tab is `tasks`, and the user presses `j` or `k`
- **THEN** the viewport scrolls and the task cursor position does not change
- **AND** the rendered task list remains visible
