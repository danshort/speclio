## MODIFIED Requirements

### Requirement: Real-time index updates
While the mode is `ModeIndex`, the TUI SHALL detect on each tick (≤ 500 ms) whether the list of active changes, the list of archived changes, or the list of project specs has changed on disk. If any structural change is detected, the index SHALL reload all three lists, rebuild the navigable items, and refresh the viewport without the user having to leave and re-enter `ModeIndex`. Additionally, when no structural change is detected, the TUI SHALL check each active change's `tasks.md` signature (modification time and size) and re-read and re-parse only the changes whose `tasks.md` has changed (see the `reload-freshness` capability); unchanged changes SHALL NOT be re-read or re-parsed. If any task content has changed, the TUI SHALL rebuild the index items and refresh the viewport so that progress bars reflect the latest task completion state. The cursor SHALL be preserved if the resulting index has at least as many items as the current position; otherwise it SHALL move to the last available item.

If a required disk read fails while loading or reloading project data on a tick, the TUI SHALL surface the error in the status line rather than silently skipping the update. A periodic reload error SHALL NOT crash the TUI or discard the currently displayed data.

#### Scenario: New spec appears on disk while the index is open
- **WHEN** the mode is `ModeIndex` and a new directory is created in `openspec/specs/`
- **THEN** within a maximum of 500 ms the index shows the new spec in the "Specifications" section without user intervention

#### Scenario: Spec disappears from specs while the index is open
- **WHEN** the mode is `ModeIndex` and a directory is deleted from `openspec/specs/`
- **THEN** within a maximum of 500 ms the spec disappears from the "Specifications" section

#### Scenario: New archived change while the index is open
- **WHEN** the mode is `ModeIndex` and a change is moved to `openspec/changes/archive/`
- **THEN** within a maximum of 500 ms the change appears in the "Archived Changes" section

#### Scenario: New active change while the index is open
- **WHEN** the mode is `ModeIndex` and a new change is created in `openspec/changes/`
- **THEN** within a maximum of 500 ms the change appears in the "Active Changes" section

#### Scenario: Cursor preserved when the item still exists
- **WHEN** the index reloads and the number of items does not decrease below the cursor position
- **THEN** the cursor stays at the same numeric position

#### Scenario: Cursor readjusted when the item disappears
- **WHEN** the index reloads and the number of items is less than the current cursor position
- **THEN** the cursor moves to the last available item

#### Scenario: Tasks updated on disk while the index is open
- **WHEN** the mode is `ModeIndex` and the `tasks.md` file of an active change is externally modified (e.g., a checkbox is toggled)
- **THEN** within a maximum of 500 ms the progress bar for that change in the index reflects the updated `done/total` count without user intervention

#### Scenario: Unchanged changes are not re-read on a tick
- **WHEN** the mode is `ModeIndex` and no active change's `tasks.md` has changed since the previous tick
- **THEN** the tick performs no task-content re-read or re-parse for those changes (only cheap signature checks)

#### Scenario: Reload error is surfaced, not swallowed
- **WHEN** the mode is `ModeIndex` and a disk read required for the periodic reload fails
- **THEN** the error is shown in the status line, the previously displayed index data is retained, and the TUI keeps running
