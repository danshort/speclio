## ADDED Requirements

### Requirement: Resolve the opener from the editor configuration
The opener used by `e` SHALL be resolved from `editor.open_with` (see the `config-file` capability) into one of two launch modes:
- **Terminal mode** — when `open_with` is unset or `"$EDITOR"` (the editor is `$EDITOR` split into fields, else `vi`), or when `open_with` is any other command string (the command split into fields). The TUI SHALL launch via `tea.ExecProcess`, yielding the terminal and resuming on exit.
- **Detached mode** — when `open_with` is `"system"`. The opener SHALL be the operating system's default handler (`open` on macOS, `xdg-open` on Linux, `start` on Windows). The TUI SHALL launch it without yielding the terminal and SHALL continue running; the edited file is picked up by the normal reload path.

If the resolved opener cannot be launched (the executable is not found, or the launch fails), the TUI SHALL surface the error rather than silently ignoring it.

#### Scenario: Default resolves to a terminal editor
- **WHEN** `editor.open_with` is unset
- **THEN** the opener is `$EDITOR` (or `vi`) in terminal mode

#### Scenario: System handler resolves to detached mode
- **WHEN** `editor.open_with` is `"system"`
- **THEN** the opener is the OS default handler in detached mode, and pressing `e` does not yield the terminal

#### Scenario: Command resolves to a terminal editor
- **WHEN** `editor.open_with` is `"nvim"`
- **THEN** the opener is `nvim` in terminal mode

#### Scenario: Launch failure is surfaced
- **WHEN** the resolved opener cannot be found or fails to launch
- **THEN** the TUI shows an error message instead of silently doing nothing

## MODIFIED Requirements

### Requirement: Open the active artifact in the external editor
The TUI SHALL allow the user to open the file of the active tab's artifact by pressing `e`. The opener and its launch mode SHALL be resolved from configuration (see "Resolve the opener from the editor configuration"). In terminal mode the TUI SHALL suspend its control of the terminal before launching using `tea.ExecProcess` and resume it upon exit; in detached mode it SHALL launch the opener without yielding the terminal. After returning from (or launching) the editor, mouse tracking SHALL still be functional because mouse mode is declared in `View()` and re-applied on every render frame.

#### Scenario: Open proposal in editor
- **WHEN** the active tab is `proposal`, `editor.open_with` is at its default, and the user presses `e`
- **THEN** the TUI yields the terminal and opens `$EDITOR proposal.md`; when the editor is closed the TUI resumes with functional mouse tracking

#### Scenario: Open tasks in editor
- **WHEN** the active tab is `tasks`, the opener is a terminal editor, and the user presses `e`
- **THEN** the TUI yields the terminal and opens the editor on `tasks.md`; when the editor is closed the TUI resumes

#### Scenario: Fallback to vi when $EDITOR is not defined
- **WHEN** `editor.open_with` is at its default, `$EDITOR` is not defined, and the user presses `e`
- **THEN** the TUI launches `vi` with the path of the active artifact

#### Scenario: Open in the system default app
- **WHEN** `editor.open_with` is `"system"` and the user presses `e`
- **THEN** the TUI launches the OS default handler on the active artifact without yielding the terminal, and the change is reflected after the file is saved via the normal reload path

#### Scenario: Key e disabled on tab
- **WHEN** the user presses `e` and the active tab has no available artifact (`Present == false`)
- **THEN** nothing happens

#### Scenario: Mouse wheel works after editor return
- **WHEN** the user returns from a terminal editor
- **THEN** the user can immediately scroll the viewport with the mouse wheel without needing to restart the TUI

### Requirement: Open the current spec in the external editor

The TUI SHALL allow the user to open the spec being viewed by pressing `e` while in `ModeViewingSpec`. The opener and launch mode SHALL be resolved from configuration (see "Resolve the opener from the editor configuration"). The file opened SHALL be `openspec/specs/<name>/spec.md` for the spec currently being viewed. This SHALL apply both when the full spec is rendered and when a single requirement is focused, since requirements are sections within the same `spec.md` file. In terminal mode the TUI SHALL suspend control of the terminal using `tea.ExecProcess` and resume on exit, with mouse tracking still functional after returning; in detached mode it SHALL launch without yielding the terminal.

#### Scenario: Open full spec in editor

- **WHEN** the mode is `ModeViewingSpec` showing a full spec and the user presses `e`
- **THEN** the TUI opens the resolved opener on that spec's `openspec/specs/<name>/spec.md`; for a terminal editor it yields and resumes with functional mouse tracking

#### Scenario: Open spec in editor while a requirement is focused

- **WHEN** the mode is `ModeViewingSpec` focused on a single requirement and the user presses `e`
- **THEN** the TUI opens the resolved opener on the same spec's `openspec/specs/<name>/spec.md` (the file containing that requirement)

#### Scenario: Fallback to vi when $EDITOR is not defined

- **WHEN** the mode is `ModeViewingSpec`, `editor.open_with` is at its default, `$EDITOR` is not defined, and the user presses `e`
- **THEN** the TUI launches `vi` with the path of the spec being viewed

#### Scenario: Help bar advertises the edit shortcut in spec view

- **WHEN** the mode is `ModeViewingSpec`
- **THEN** the help bar includes `e: edit`
