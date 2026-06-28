## MODIFIED Requirements

### Requirement: Task identity for safe writes ignores the number and tolerates strikethrough
When locating a task on disk before writing, the system SHALL match on the task's description text with the leading `<prefix>.<ordinal>` number removed, so that renumbering does not break the match. Matching SHALL also tolerate descriptions wrapped in `~~…~~` (skipped tasks).

Task descriptions SHALL be unique within a section (a duplicate is a malformed `tasks.md`, since OpenSpec tasks are numbered). When more than one task in the target section matches the identity, the system SHALL treat it as a conflict and make no change, surfacing a message that prompts the user to disambiguate — it SHALL NOT silently act on the first match.

#### Scenario: Match survives renumbering
- **WHEN** a task displayed as `1.3 Add frontend component` is to be modified but the file now numbers it `1.2`
- **THEN** the task is still located by the description "Add frontend component"

#### Scenario: Match a strikethrough task
- **WHEN** the on-disk task is `- [ ] ~~6.1 Drop the cache~~ (skipped)`
- **THEN** the task is located by the description "Drop the cache (skipped)" — the leading `<prefix>.<ordinal>` number and the `~~` strikethrough markers are removed, while text outside the markers is kept

#### Scenario: Ambiguous match is a conflict, not a silent first-match
- **WHEN** an edit (delete/move/inline-edit) targets a description that matches two tasks in the same section
- **THEN** no write occurs and the user is shown a message to rename one of them, rather than the first match being changed

### Requirement: Add a task after the selected task
The system SHALL insert a new pending task (`- [ ]`) immediately after the currently selected task, within the same section, and SHALL assign it the next sequential ordinal so the section's ordinals remain contiguous. The new task's placeholder text SHALL be unique within its section (e.g. `New task`, then `New task 2`, …) so that adding several tasks does not create colliding identities.

#### Scenario: Insert into the middle of a section
- **WHEN** the selected task is `1.2` in a section containing `1.1`, `1.2`, `1.3`
- **THEN** a new task is inserted as `1.3` and the former `1.3` becomes `1.4`

#### Scenario: Insert after the last task in a section
- **WHEN** the selected task is the last task `2.4` of section 2
- **THEN** a new task is appended as `2.5`

#### Scenario: Repeated adds get distinct placeholders
- **WHEN** the user adds two tasks in the same section without renaming the first
- **THEN** the placeholders differ (e.g. `New task` and `New task 2`), so the two tasks have distinct identities
