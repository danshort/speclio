## ADDED Requirements

### Requirement: Periodic reloads are gated on a file signature
On each periodic tick, before re-reading and re-parsing a change, the TUI SHALL compare the current signature of its `tasks.md` — modification time and size — against the signature last seen for that path. The TUI SHALL re-read and re-parse the change only when the signature has changed; when it is unchanged the change SHALL NOT be re-read or re-parsed. This gating governs the per-change reloads whose cost scales with the number of changes: the index task-progress reload and the worktrees per-change reload.

#### Scenario: Unchanged file is not re-read on a tick
- **WHEN** a periodic tick occurs and a change's `tasks.md` signature is identical to the signature last seen
- **THEN** the file is not re-read or re-parsed and no rebuild/refresh is triggered for it

#### Scenario: Changed file is reloaded once
- **WHEN** a change's `tasks.md` is modified so its signature differs from the last seen signature
- **THEN** the TUI re-reads and re-parses that file on the next tick, updates the stored signature, and refreshes the affected view

### Requirement: First-seen files are treated as changed
When the TUI has no recorded signature for a path (e.g. on the first tick, or for a newly appeared change), it SHALL treat the file as changed and load it, then record its signature.

#### Scenario: First tick loads then quiesces
- **WHEN** the first periodic tick runs and no signatures have been recorded yet
- **THEN** the relevant files are read once and their signatures recorded, and subsequent ticks with no on-disk change perform no further reads

### Requirement: A missing file counts as a change
When a previously present file can no longer be stat'd (e.g. it was deleted), the TUI SHALL treat this as a change (present → absent) and update its state and stored signature accordingly, rather than retaining stale content.

#### Scenario: Deleted tasks.md is detected
- **WHEN** a change's `tasks.md` existed on a prior tick but is absent on the current tick
- **THEN** the TUI detects the transition and updates the change's task state to reflect the absence

### Requirement: Explicit reloads bypass gating
Reloads triggered by an explicit user action — notably returning from the external editor — SHALL re-read the affected artifact unconditionally, without consulting the signature cache, so a just-saved edit is always reflected immediately.

#### Scenario: Editor-return reloads regardless of signature
- **WHEN** the user edits an artifact in the external editor and returns
- **THEN** the TUI re-reads and re-renders that artifact even if its cached signature would otherwise suggest no change
