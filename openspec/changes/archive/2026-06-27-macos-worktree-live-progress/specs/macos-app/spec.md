## ADDED Requirements

### Requirement: Live worktree progress
While the Worktrees mode is active, the app SHALL periodically re-read the surveyed worktree changes and reflect any change in their task progress — in the sidebar and in an open read-only worktree artifact — without a manual reload. Polling SHALL run only while the Worktrees mode is active and SHALL stop otherwise, and SHALL not re-enumerate the worktree set on each tick.

#### Scenario: Foreign worktree progress updates live
- **WHEN** the user is in the Worktrees mode and a task in another worktree's change is completed externally
- **THEN** that change's progress updates in the sidebar (and in the open read-only artifact, if shown) within a short interval, without a manual reload

#### Scenario: Polling is scoped to the mode
- **WHEN** the user leaves the Worktrees mode (or closes the window)
- **THEN** worktree polling stops

#### Scenario: No spurious updates
- **WHEN** a poll finds no change in the surveyed worktree changes
- **THEN** the sidebar is not rebuilt and the current selection is preserved
