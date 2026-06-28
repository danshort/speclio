## MODIFIED Requirements

### Requirement: Reveal and open the selected file
The app SHALL let the user reveal the currently selected item (artifact, project spec, config, or worktree) in Finder and open it externally. Opening SHALL use the configured editor application (see "Editor application preference") when one is set and available, and otherwise the system default application for the file. The open action SHALL be available both from the file-actions menu and via a `⌘E` keyboard shortcut / menu item, and SHALL be disabled when there is no file-backed selection.

#### Scenario: Reveal in Finder
- **WHEN** the user chooses "Reveal in Finder" for a selection backed by a file or directory
- **THEN** Finder opens with that item selected

#### Scenario: Open in the system default app
- **WHEN** no editor app is configured and the user opens a file-backed selection
- **THEN** the file opens in its default application

#### Scenario: Open in the configured editor
- **WHEN** an editor app is configured and the user opens a file-backed selection
- **THEN** the file opens in that application

#### Scenario: Open via keyboard shortcut
- **WHEN** a file-backed selection is active and the user presses `⌘E`
- **THEN** the current artifact opens externally (configured app, else default)

#### Scenario: Disabled without a file
- **WHEN** the selection is not backed by a file
- **THEN** the open action and its `⌘E` shortcut are disabled

## ADDED Requirements

### Requirement: Editor application preference
The Settings window SHALL let the user choose a specific application to open artifacts with, via an application picker. The chosen application SHALL persist across launches. When no application is chosen, or the chosen application is no longer present, the app SHALL fall back to the system default application. The user SHALL be able to clear the choice (return to the system default).

#### Scenario: Choose an app
- **WHEN** the user picks an application in Settings
- **THEN** subsequent opens use that application, and the choice persists across relaunch

#### Scenario: Reset to default
- **WHEN** the user clears the chosen application
- **THEN** subsequent opens use the system default application again

#### Scenario: Missing app falls back
- **WHEN** a chosen application no longer exists at its stored location
- **THEN** opening falls back to the system default application rather than failing
