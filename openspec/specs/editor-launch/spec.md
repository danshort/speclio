# editor-launch Specification

## Purpose
Allows editing the active artifact of a change in the system editor (`$EDITOR`) by pressing `e`, and automatically reloads the content when the editor is closed without needing to restart the TUI.

## Requirements


### Requirement: Open the active artifact in the external editor
The TUI SHALL allow the user to open the file of the active tab's artifact in the system editor by pressing `e`. The editor SHALL be the value of the `$EDITOR` environment variable; if it is not defined, `vi` SHALL be used as a fallback. The TUI SHALL correctly suspend its control of the terminal before launching the editor using `tea.ExecProcess` and resume it upon exit. After returning from the editor, mouse tracking SHALL still be functional because mouse mode is declared in `View()` and re-applied on every render frame.

#### Scenario: Open proposal in editor
- **WHEN** the active tab is `proposal` and the user presses `e`
- **THEN** the TUI yields the terminal and opens `$EDITOR proposal.md`; when the editor is closed the TUI resumes with functional mouse tracking

#### Scenario: Open tasks in editor
- **WHEN** the active tab is `tasks` and the user presses `e`
- **THEN** the TUI yields the terminal and opens `$EDITOR tasks.md`; when the editor is closed the TUI resumes

#### Scenario: Fallback to vi when $EDITOR is not defined
- **WHEN** `$EDITOR` is not defined in the environment and the user presses `e`
- **THEN** the TUI launches `vi` with the path of the active artifact

#### Scenario: Key e disabled on tab
- **WHEN** the user presses `e` and the active tab has no available artifact (`Present == false`)
- **THEN** nothing happens

#### Scenario: Mouse wheel works after editor return
- **WHEN** the user returns from the external editor
- **THEN** the user can immediately scroll the viewport with the mouse wheel without needing to restart the TUI

### Requirement: Immediate reload after closing the editor
The TUI SHALL reload the content of the edited artifact immediately upon returning from the editor, without waiting for the next polling cycle.

#### Scenario: Reload tasks after editing
- **WHEN** the user edits `tasks.md` in the editor and closes the editor
- **THEN** the TUI shows the updated tasks content instantly, with the cursor restored by text

#### Scenario: Reload markdown artifact after editing
- **WHEN** the user edits `proposal.md`, `design.md`, or a `spec.md` and closes the editor
- **THEN** the TUI invalidates the render cache for that tab and re-renders with the new content

### Requirement: Open the current spec in the external editor

The TUI SHALL allow the user to open the spec being viewed in the system editor by pressing `e` while in `ModeViewingSpec`. The editor SHALL be the value of the `$EDITOR` environment variable; if it is not defined, `vi` SHALL be used as a fallback. The file opened SHALL be `openspec/specs/<name>/spec.md` for the spec currently being viewed. This SHALL apply both when the full spec is rendered and when a single requirement is focused, since requirements are sections within the same `spec.md` file. The TUI SHALL suspend its control of the terminal before launching the editor using `tea.ExecProcess` and resume it on exit, with mouse tracking still functional after returning.

#### Scenario: Open full spec in editor

- **WHEN** the mode is `ModeViewingSpec` showing a full spec and the user presses `e`
- **THEN** the TUI yields the terminal and opens `$EDITOR` on that spec's `openspec/specs/<name>/spec.md`; when the editor is closed the TUI resumes with functional mouse tracking

#### Scenario: Open spec in editor while a requirement is focused

- **WHEN** the mode is `ModeViewingSpec` focused on a single requirement and the user presses `e`
- **THEN** the TUI opens `$EDITOR` on the same spec's `openspec/specs/<name>/spec.md` (the file containing that requirement)

#### Scenario: Fallback to vi when $EDITOR is not defined

- **WHEN** the mode is `ModeViewingSpec`, `$EDITOR` is not defined in the environment, and the user presses `e`
- **THEN** the TUI launches `vi` with the path of the spec being viewed

#### Scenario: Help bar advertises the edit shortcut in spec view

- **WHEN** the mode is `ModeViewingSpec`
- **THEN** the help bar includes `e: edit`

### Requirement: Reload spec content after editing

The TUI SHALL reload the content of the edited spec immediately upon returning from the editor, so that changes made externally are reflected without restarting the TUI. The view SHALL remain in `ModeViewingSpec`, preserving full-spec or requirement-focus state.

#### Scenario: Reload spec after editing

- **WHEN** the user edits a spec's `spec.md` in the editor and closes it while in `ModeViewingSpec`
- **THEN** the TUI re-renders the spec view with the updated content, staying in `ModeViewingSpec`

#### Scenario: Focused requirement reflects edits

- **WHEN** a single requirement is focused and the user edits that requirement's text in the editor and closes it
- **THEN** the focused requirement view shows the updated requirement content
