## ADDED Requirements

### Requirement: Persistent project label
The app SHALL keep the open project's name visible across all navigation states, so the user always knows which project they are viewing. The project name SHALL be shown as a header at the top of the sidebar AND as the window title, with the current location shown as the window subtitle.

#### Scenario: Project name in the sidebar header
- **WHEN** a project is open
- **THEN** the project's name is shown as a prominent header at the top of the sidebar, above the navigation list, regardless of the current mode or selection

#### Scenario: Project name in the window title
- **WHEN** a project is open
- **THEN** the window title shows the project's name, and the window subtitle shows the current location (mode, and the selected change/spec/artifact where applicable)

#### Scenario: Persists across navigation
- **WHEN** the user switches modes (Active, Archived, Specs, Worktrees) or selects different items
- **THEN** the project name remains visible in both the sidebar header and the window title; only the window subtitle changes to reflect the new location

#### Scenario: No project open
- **WHEN** no project is open
- **THEN** no project header or project title is shown, and the empty state communicates that no project is open
