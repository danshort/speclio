# path-arg Specification

## Purpose
Allows launching `lectern` with an explicit path to a specific change, bypassing the scan of `openspec/changes/` and limiting polling to that single change.

## Requirements


### Requirement: Invocation with explicit path
The binary SHALL accept an optional positional argument: the path to a change directory. If provided, it SHALL load only that change and display it in the TUI without scanning `openspec/changes/`. If not provided, the behavior SHALL be identical to the current behavior.

#### Scenario: Invocation without arguments
- **WHEN** the user runs `./lectern` with no arguments
- **THEN** the TUI starts in normal mode loading all active changes from `openspec/changes/`

#### Scenario: Invocation with path to an active change
- **WHEN** the user runs `./lectern ./openspec/changes/my-change`
- **THEN** the TUI shows only the artifacts of `my-change`

#### Scenario: Invocation with path to an archived change
- **WHEN** the user runs `./lectern ./openspec/changes/archive/2026-05-02-my-change`
- **THEN** the TUI shows only the artifacts of the archived change, with tab navigation the same as in normal mode

### Requirement: Path validation
If a path is provided and does not correspond to a valid change, the binary SHALL print a descriptive error message to stderr and exit with exit code 1, without opening the TUI.

#### Scenario: Nonexistent path
- **WHEN** the user runs `./lectern ./path/that/does/not/exist`
- **THEN** the binary prints `"error: path not found: ./path/that/does/not/exist"` and exits with code 1

#### Scenario: Path without .openspec.yaml
- **WHEN** the user runs `./lectern ./some/directory` and that directory does not contain `.openspec.yaml`
- **THEN** the binary prints `"error: not a valid change directory (missing .openspec.yaml)"` and exits with code 1

### Requirement: Stable polling in path mode
When the TUI starts with an explicit path, the polling cycle SHALL be limited to reloading the artifacts of the specified change. It SHALL NOT attempt to reload the change list or detect new changes in `openspec/changes/`.

#### Scenario: Tick in path mode
- **WHEN** the TUI is open with an explicit path and the tick fires
- **THEN** only `ReloadChange` is called for the loaded change; `ListChangeNames` and `Load` are not called
