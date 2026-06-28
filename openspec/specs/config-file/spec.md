# config-file Specification

## Purpose
A per-user TOML configuration file for the TUI at `$XDG_CONFIG_HOME/lectern/config.toml` (falling back to `~/.config/lectern/config.toml`). The file is optional and resilient: a missing file yields defaults, a malformed file warns (in the status line and on stderr) and falls back to defaults without blocking launch, and unknown keys are ignored for forward compatibility. v1 exposes the `editor.open_with` setting and a `c` keybinding to open the file in the editor.

## Requirements

### Requirement: Per-user TOML configuration file
The TUI SHALL read a per-user configuration file in TOML format at `$XDG_CONFIG_HOME/lectern/config.toml`, falling back to `~/.config/lectern/config.toml` when `$XDG_CONFIG_HOME` is unset, on all platforms. The file is optional.

#### Scenario: XDG_CONFIG_HOME honored
- **WHEN** `$XDG_CONFIG_HOME` is set to `/cfg`
- **THEN** the TUI reads `/cfg/lectern/config.toml`

#### Scenario: Default to ~/.config
- **WHEN** `$XDG_CONFIG_HOME` is unset
- **THEN** the TUI reads `~/.config/lectern/config.toml`

### Requirement: Missing config file yields defaults
When the configuration file does not exist, the TUI SHALL use built-in defaults for every setting and SHALL NOT report an error.

#### Scenario: No config file present
- **WHEN** no `config.toml` exists at the resolved path
- **THEN** the TUI starts normally with all settings at their defaults

### Requirement: Malformed config warns and falls back to defaults
When the configuration file exists but cannot be parsed as TOML, the TUI SHALL surface a warning identifying the problem and SHALL continue with built-in defaults rather than exiting. Because the TUI's alternate screen hides stderr during a session, the warning SHALL be shown in the in-app status line (in addition to stderr).

#### Scenario: Syntax error does not block launch
- **WHEN** `config.toml` contains invalid TOML
- **THEN** the TUI launches with default settings and shows a warning in the status line (and on stderr)

#### Scenario: Smart-quoted value is reported, not silently ignored
- **WHEN** `config.toml` contains a value typed with curly quotes (e.g. `open_with = “system"`), which is invalid TOML
- **THEN** the TUI does not silently fall back; it shows the parse warning in the status line so the user can see the config was rejected

### Requirement: Unknown keys are ignored
The configuration loader SHALL ignore keys it does not recognize, so that forward-compatible additions and minor typos degrade gracefully instead of failing.

#### Scenario: Unrecognized key tolerated
- **WHEN** `config.toml` contains a key the current version does not define
- **THEN** the unknown key is ignored and recognized settings still load

### Requirement: Editor open-with setting
The configuration SHALL expose an `editor.open_with` string controlling how the active artifact is opened. Recognized values: unset or `"$EDITOR"` (use the `$EDITOR` environment variable, falling back to `vi`), `"system"` (use the operating system's default handler), or any other string (treated as an editor command). The default when unset is the `$EDITOR`/`vi` behavior.

#### Scenario: Default when unset
- **WHEN** `config.toml` has no `[editor]` table or no `open_with` key
- **THEN** `editor.open_with` resolves to the `$EDITOR`/`vi` default

#### Scenario: System handler selected
- **WHEN** `config.toml` sets `editor.open_with = "system"`
- **THEN** the configured opener is the operating system's default handler

#### Scenario: Custom command selected
- **WHEN** `config.toml` sets `editor.open_with = "nvim"`
- **THEN** the configured opener is the `nvim` command

### Requirement: Open the config file from the TUI
The TUI SHALL provide a keybinding (`c`, available in index and change-viewer modes) that opens the user config file in the configured editor. When the file does not yet exist, the system SHALL create it with a documented, all-defaults starter template before opening. Changes take effect on the next launch.

#### Scenario: Open an existing config
- **WHEN** the user presses `c` and `config.toml` exists
- **THEN** the TUI opens it using the resolved editor opener

#### Scenario: Create then open when missing
- **WHEN** the user presses `c` and no `config.toml` exists
- **THEN** the TUI creates the parent directory and a commented starter file, then opens it
