## Why

`$EDITOR` is the TUI's only way to open an artifact, and it's a pain to configure correctly. For OpenSpec — markdown that agents generate and humans review and lightly tweak — many users want the file to open in the GUI markdown app they already use (Obsidian/Typora/iA/VS Code), not `vi`. There is no per-user config file to express that preference. This change adds one, with the editor override as its first real setting (#94 + #95, bundled — a config file is hard to motivate without a consumer).

## What Changes

- Add a per-user TOML config at `$XDG_CONFIG_HOME/lectern/config.toml` (falling back to `~/.config/lectern/config.toml`) on all platforms. Missing file → all defaults; a malformed file → warn to stderr and continue with defaults (never block launch). Unknown keys are ignored (forward-compatible).
- v1 contains exactly one setting, `[editor] open_with`, controlling how `e` opens the active artifact:
  - unset or `"$EDITOR"` → `$EDITOR` then `vi` — a **terminal** editor run via `tea.ExecProcess` (today's behavior, the default). SSH/headless-safe.
  - `"system"` → the OS default handler (`open` / `xdg-open` / `start`), launched **detached** (the TUI does not yield the terminal; the live-reload poll catches the save when you return).
  - any other value → treated as a **terminal** editor command (e.g. `"nvim"`, `"code --wait"`) run via `tea.ExecProcess`.
- The launch mode (terminal vs detached) is **implied by the value** — no separate flag.
- Harden the launch: verify the resolved opener exists before launching; surface launch/exit errors instead of swallowing them (`internal/ui/viewer.go:170` currently always returns `editorReturnMsg{}` and ignores the error).
- Keep the `e` keybinding, and add a `c` keybinding (index + change-viewer) that opens the config file in the editor, creating a documented starter file if none exists.
- Surface a rejected (malformed) config in the TUI status line, not just stderr — the alt-screen hides stderr during a session, so a swallowed warning would otherwise read as a silent fallback.

## Capabilities

### New Capabilities
- `config-file`: a per-user TOML configuration file for the TUI — its location/precedence, missing-vs-malformed handling, forward-compatible parsing, and the `[editor] open_with` setting.

### Modified Capabilities
- `editor-launch`: `e` no longer always runs `$EDITOR`; it honors `editor.open_with`, supporting a detached system-handler handoff and an explicit terminal-command override, and surfaces launch errors.

## Impact

- **Code:** new config loader in `internal/openspec` (or a new `internal/config`) using a new dependency `github.com/BurntSushi/toml`; `cmd/lectern/main.go` to load the user config and thread the editor preference into the model; `internal/ui/viewer.go` `openInEditor` to branch on the resolved mode (terminal `ExecProcess` vs detached launch) and to verify/​surface errors.
- **Dependency:** adds `github.com/BurntSushi/toml` (passes `govulncheck`).
- **Tests:** config parsing (missing/malformed/unknown-keys/precedence), opener resolution (default → `$EDITOR` → `vi`; `"system"` → platform handler; command), and launch-mode selection.
- **No macOS change** — macOS already defaults to the system app and gets its override separately (#110).

## Non-goals

- macOS editor override / keyboard shortcut — separate issue #110.
- Other config keys (theme, keybindings, watch interval) — the loader is forward-compatible, but v1 ships only `[editor]`.
- Replacing 500 ms polling with fsnotify — #90.
- Repo-local / per-project config — the editor preference is a user/machine setting; global only.
- Live-reload of the config file — read once at startup.
- A `detach = true` flag for custom GUI commands — use `"system"`; the flag can be added later if a real need appears.
