## ADDED Requirements

### Requirement: Open the worktrees view from the index
While the mode is `ModeIndex`, pressing `w` SHALL switch the mode to `ModeWorktrees`. The index helpbar SHALL advertise this with a static `w` affordance. The index SHALL NOT compute or display live cross-worktree counts, so that opening the index does not trigger any cross-worktree discovery or polling.

#### Scenario: Pressing w opens the worktrees view
- **WHEN** the mode is `ModeIndex` and the user presses `w`
- **THEN** the mode switches to `ModeWorktrees`

#### Scenario: Helpbar advertises the worktrees view
- **WHEN** the mode is `ModeIndex`
- **THEN** the helpbar includes a static `w` entry for the worktrees view and shows no live cross-worktree counts
