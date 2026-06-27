## ADDED Requirements

### Requirement: Persistent change progress bar
When the current selection belongs to a change (an active, archived, or worktree change) that has tasks, the detail pane SHALL show a persistent progress bar at the top reflecting that change's overall task completion, visible regardless of which of the change's artifacts is open. The bar SHALL be hidden for selections that are not changes (project specs, project config, worktree metadata) and for changes with no tasks.

#### Scenario: Progress visible across a change's artifacts
- **WHEN** the user views any artifact (proposal, design, a spec, or tasks) of a change that has tasks
- **THEN** a progress bar at the top of the detail pane shows that change's overall completed/total

#### Scenario: Hidden for non-change views
- **WHEN** the selection is a project spec, the project config, or worktree metadata (or a change with no tasks)
- **THEN** no persistent change progress bar is shown

### Requirement: Per-section task progress
In the Tasks view, each section heading SHALL show that section's task completion (completed/total of the tasks under it) alongside the heading. The Tasks view SHALL NOT show a separate overall progress bar — the change's overall progress is the persistent bar at the top of the detail pane.

#### Scenario: Section progress shown
- **WHEN** a tasks file has sections with tasks
- **THEN** each section heading displays the completed/total for the tasks in that section

#### Scenario: No overall bar in the Tasks view
- **WHEN** the Tasks view is shown
- **THEN** it shows per-section progress only, not a separate overall progress bar (overall progress is the persistent detail-pane bar)

### Requirement: Change progress in the sidebar
Every change row in the sidebar — active, archived, and worktree changes — SHALL show its task progress (completed/total), so progress is visible without opening the change.

#### Scenario: Active and archived changes show progress
- **WHEN** the sidebar lists active or archived changes that have tasks
- **THEN** each change row shows its completed/total task progress, consistent with worktree-change rows
