## ADDED Requirements

### Requirement: Multiple project windows
The app SHALL allow more than one project to be open at the same time, each in its own window (and macOS window tab), with independent navigation state (mode, selection, sidebar, and live reload). Opening a project SHALL NOT close or replace any already-open project.

#### Scenario: Open a second project without closing the first
- **WHEN** a project is open and the user opens another project
- **THEN** the second project opens in its own window and the first remains open and unchanged

#### Scenario: Independent navigation per window
- **WHEN** two project windows are open
- **THEN** changing the mode or selection in one window does not affect the other

#### Scenario: Window tabs hold different projects
- **WHEN** the user creates a new tab and opens a project in it
- **THEN** that tab shows the newly opened project, independent of the other tabs

### Requirement: Open into an empty window or a new one
The app SHALL open a chosen project into the current window when that window has no project loaded, and otherwise SHALL open it in a new window, so that opening a project never evicts another. Each project SHALL have at most one window: opening a project that is already open SHALL focus its existing window rather than creating a duplicate.

#### Scenario: Reuse an empty window
- **WHEN** the focused window has no project loaded and the user opens a project
- **THEN** the project loads into that window and no stray empty window remains

#### Scenario: Open in a new window when the current one is in use
- **WHEN** the focused window already has a project loaded and the user opens another project
- **THEN** the project opens in a new window and the focused window is unchanged

#### Scenario: Opening an already-open project focuses it
- **WHEN** the user opens a project that is already open in another window
- **THEN** that existing window is brought to the front and no duplicate window is created

### Requirement: Open Recent
The app SHALL provide a File ▸ Open Recent menu listing recently opened projects, most-recent first. Selecting an entry SHALL open the project (or focus it if already open). The menu SHALL include a Clear Menu action.

#### Scenario: Reopen a recent project
- **WHEN** the user chooses a project from Open Recent
- **THEN** the project opens in a window, or its existing window is focused if it is already open

#### Scenario: Recents update on open
- **WHEN** the user opens a project
- **THEN** it appears at the top of the Open Recent menu, without duplicate entries

#### Scenario: Clear the menu
- **WHEN** the user chooses Clear Menu
- **THEN** the Open Recent list is emptied, and any currently-open project windows are unaffected

### Requirement: Reopen open projects on launch
The app SHALL remember the set of projects that were open when it last quit and reopen a window for each on the next launch.

#### Scenario: Restore multiple projects
- **WHEN** the user quits with two or more projects open and relaunches the app
- **THEN** a window reopens for each previously-open project

#### Scenario: Unresolvable project on restore
- **WHEN** a remembered project can no longer be resolved (moved, deleted, or permission revoked)
- **THEN** the app still launches and that window shows the empty/error state instead of failing

### Requirement: Per-window commands
Menu and toolbar commands that act on a project SHALL act on the focused window's project, so commands invoked in one window do not affect another. The content text-size preference SHALL remain global across all windows.

#### Scenario: Reload affects only the focused window
- **WHEN** the user invokes Reload (or a file action) with multiple windows open
- **THEN** only the focused window's project is reloaded / acted upon

#### Scenario: Text size applies everywhere
- **WHEN** the user changes the content text size
- **THEN** the change applies to rendered content in all open windows
