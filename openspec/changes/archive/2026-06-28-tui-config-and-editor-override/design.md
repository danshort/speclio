## Context

The TUI's `openInEditor` (`internal/ui/viewer.go`) hardcodes `$EDITOR` (→ `vi`) and launches via `tea.ExecProcess`, which yields the terminal and blocks until the editor exits — correct for a terminal editor, the only kind it supports. There is no per-user config; the `cfg` threaded through `main.go` today is the *OpenSpec project* config (`openspec/config.yaml`), unrelated to user preferences. The project's only config dependency is `gopkg.in/yaml.v3`.

This change introduces a per-user config file and uses it to let `e` open artifacts in something other than a terminal editor — most usefully, the user's default GUI markdown app — while keeping `$EDITOR` the default so the TUI stays usable over SSH/headless.

## Goals / Non-Goals

**Goals:**
- A per-user config file with clear location, precedence, and resilient parsing.
- `e` honors an `editor.open_with` preference with three modes; launch mode implied by the value.
- Keep `$EDITOR` the default; never block launch on a bad config.
- Surface launch failures instead of swallowing them.

**Non-Goals:**
- macOS override/shortcut (#110); other config keys; fsnotify (#90); repo-local config; live config reload; a `detach` flag.

## Decisions

### D1 — TOML at the XDG path
Config lives at `$XDG_CONFIG_HOME/lectern/config.toml`, falling back to `~/.config/lectern/config.toml`, on **all** platforms (it's a CLI; terminal users expect `~/.config`). Format is **TOML** via a new dependency, `github.com/BurntSushi/toml`.
- *Why TOML over the project's YAML:* chosen for hand-edited ergonomics. Trade-off accepted: one new dependency (well-maintained, `govulncheck`-clean).
- *Alternative:* `os.UserConfigDir()` (macOS → `~/Library/Application Support`) — rejected as non-idiomatic for a CLI.

### D2 — Resilient, forward-compatible parsing
Missing file → all defaults, no error. Malformed file (TOML syntax error) → print a warning to stderr and continue with defaults; do **not** exit. Unknown keys are ignored (BurntSushi reports them via `MetaData.Undecoded`; we don't treat them as errors), so future keys and typos degrade gracefully.
- *Why:* an editor preference is cosmetic; a typo must not stop the user from reading specs. Mirrors the #92 ethos (don't fail silently *or* catastrophically).

### D3 — `editor.open_with`: one setting, three modes, mode implied by value
```toml
[editor]
open_with = "system"   # unset/"$EDITOR" (default) | "system" | a command e.g. "nvim"
```
Resolution at open time:
- **unset / `"$EDITOR"`** → `$EDITOR` split into fields, else `vi`. **Terminal** mode: `tea.ExecProcess` (yield + wait), then reload on return (today's path).
- **`"system"`** → platform handler: `open` (macOS), `xdg-open` (Linux), `start` (Windows). **Detached** mode: launch without yielding the terminal; the existing poll picks up the save.
- **any other string** → that command (field-split). **Terminal** mode via `ExecProcess`.
- *Why no flag:* the default and custom-command cases are terminal (the safe assumption for a CLI); `"system"` is the one inherently-detached case and is a known sentinel. A `detach` flag is only needed for a custom *GUI* command, which `"system"` already covers — deferred.

### D4 — Threading + launch hardening
`main.go` loads the user config alongside the project config and passes the resolved editor preference into the model. `openInEditor` branches on mode: terminal → `ExecProcess` (as today); detached → resolve the platform handler, verify it exists on `PATH`, `Start()` it, and return immediately. In both modes, a failure to launch (opener missing, exec error) sets `m.errMsg` instead of being dropped — fixing `viewer.go:170`, which currently returns `editorReturnMsg{}` and ignores the error.

### D5 — Surface a rejected config in the TUI (not just stderr)
A malformed config (e.g. a value typed with curly "smart" quotes — a real failure hit during QA) returns a parse error from `config.Load`. `main.go` warns on stderr *and* threads the message into the model's `errMsg` (a `startupWarn` constructor argument), because the TUI's alternate screen hides stderr during a session — otherwise the rejection reads as a silent fallback to `$EDITOR`.

### D6 — Open the config from the TUI
A `c` keybinding (index + change-viewer modes) opens the config file in the resolved editor, via `config.EnsureFile`, which creates a documented, all-defaults starter file when none exists. Changes apply on next launch (consistent with "read once at startup"); a live-applying in-TUI settings screen is deferred (#111).

## Risks / Trade-offs

- **New dependency (`BurntSushi/toml`)** → standard, stable, vuln-clean; acceptable.
- **`"system"` over SSH does nothing useful** → that's why it's opt-in and `$EDITOR` is the default; documented.
- **Detached launch + 500 ms poll latency** → the edited file reloads within a tick, same as today's external-edit path; acceptable.
- **`"$EDITOR"` as a literal config value vs the env var** → treat the exact string `"$EDITOR"` (and unset) as "use the `$EDITOR` env var"; documented so it isn't mistaken for shell expansion.

## Migration Plan

Additive and TUI-only. With no config file, behavior is byte-for-byte identical to today (`$EDITOR`→`vi`, `ExecProcess`). Rollback is reverting the change; no user files are written. No macOS impact.

## Open Questions

- None outstanding; all four config decisions are settled (format, location, scope, error handling).
