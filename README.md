# lectern

<p align="center">
  <a href="https://github.com/danshort/lectern/actions/workflows/ci.yml"><img src="https://img.shields.io/github/actions/workflow/status/danshort/lectern/ci.yml?branch=main&label=CI" alt="CI status" /></a>
  <a href="https://github.com/danshort/lectern/releases"><img src="https://img.shields.io/github/v/release/danshort/lectern" alt="Latest release" /></a>
  <a href="go.mod"><img src="https://img.shields.io/github/go-mod/go-version/danshort/lectern" alt="Go version" /></a>
  <a href="https://goreportcard.com/report/github.com/danshort/lectern"><img src="https://goreportcard.com/badge/github.com/danshort/lectern" alt="Go Report Card" /></a>
  <a href="LICENSE"><img src="https://img.shields.io/badge/license-MIT-blue" alt="License: MIT" /></a>
  <a href="https://github.com/danshort/lectern/issues"><img src="https://img.shields.io/github/issues/danshort/lectern" alt="Open issues" /></a>
  <a href="https://github.com/openspec"><img src="https://img.shields.io/badge/built%20with-OpenSpec-6f42c1" alt="Built with OpenSpec" /></a>
  <a href="https://www.buymeacoffee.com/danshort"><img src="https://img.shields.io/badge/Buy%20Me%20a%20Coffee-support-FFDD00?logo=buymeacoffee&logoColor=black" alt="Buy Me a Coffee" /></a>
</p>

A keyboard-driven terminal UI for reading and navigating [OpenSpec](https://github.com/openspec) project artifacts — proposals, designs, specs, and tasks.

> **lectern is a fork of [dossier](https://github.com/fselich/dossier) by [fselich](https://github.com/fselich).** It was renamed as it diverges and is maintained independently. All credit for the original tool goes to the upstream author; see [LICENSE](LICENSE).

> Built with OpenSpec. This repository contains all project-level spec files and archived changes that document the complete development history of the tool.

<p align="center">
  <img src="docs/lectern.gif" alt="lectern demo" />
</p>

---

## Features

- Navigates all active changes and their artifacts from a single interface
- Renders Markdown with full syntax highlighting
- Toggles task checkboxes (`- [ ]` / `- [x]`) in-place, writing directly to `tasks.md`
- Live-reloads on disk changes (500 ms polling)
- Opens any artifact in `$EDITOR`
- Shows a keyboard-shortcut overlay from any screen with `?`
- Surveys every git worktree of the repository in one read-only view, with each worktree's active changes and live task progress
- Accepts a path argument to view a single change directory without a full project

---

## Installation

**Requirements:** terminal with ANSI color support. Go 1.25 or later if building from source.

```bash
# Homebrew
brew tap danshort/tap
brew trust danshort/tap   # one-time: Homebrew requires trusting third-party taps
brew install lectern

# From source
git clone https://github.com/danshort/lectern
cd lectern
make build    # produces ./lectern
make install  # installs via go install

# Using go install
go install github.com/danshort/lectern/cmd/lectern@latest
```

Hacking on lectern itself? See [DEVELOPING.md](DEVELOPING.md) for building and
running a dev version alongside your installed copy.

### macOS app (preview)

A native SwiftUI reader is in development under [`macos/`](macos/) — it reuses the
same domain logic as the CLI (the `OpenSpecKit` Swift port of `internal/openspec`,
kept byte-for-byte in sync via a shared golden corpus). It presents changes,
specs, worktrees, and config with native markdown rendering and interactive task
editing.

In the Tasks view of a change in the current worktree you can edit `tasks.md`
directly — hover a task to reveal its controls:

- **Toggle** a checkbox to mark a task done.
- **Add** (`+`) a task after the selected one, **delete** (`−`) with confirmation,
  and **edit** (✎) the text inline (a three-line wrapping box; **⌘-Return** or
  clicking away saves, **Esc** cancels).
- **Drag** the handle (`☰`) to reorder within a section or move a task to another
  section, including a drop zone at the end of each section. Task numbers
  (`1.1`, `1.2`, …) renumber automatically; section prefixes (e.g. `3b`) are
  preserved.

Every task change — toggle, edit, add, delete, reorder, move — is **undoable**
with **⌘Z** (and **⌘⇧Z** to redo); the Edit menu names each action. Undo restores
`tasks.md` to its exact prior bytes and is guarded so it won't clobber an edit
made on disk in the meantime.

Edits are written surgically and re-read just before saving, so a change made by
an agent working the same file won't be clobbered; if the file moved underneath
an edit, the app refreshes and tells you. Changes opened from **other worktrees**
are read-only.

**Opening artifacts in an editor.** Press **⌘E** (or use the toolbar flyout /
File ▸ Open in Editor) to open the selected artifact externally. By default it
uses the system handler for the file; in **Settings (⌘,) → Opening artifacts**
you can pick a specific app to open artifacts with (and reset to the default).
Other shortcuts: **⌘O** open a project, **⌘+/⌘-/⌘0** adjust content text size.

Install via Homebrew (preview) — the app ships as a **cask** in the *same*
`danshort/tap` as the CLI formula (one tap, two install targets):

```bash
brew tap danshort/tap         # if you haven't already
brew install --cask danshort/tap/lectern-app
```

Or build and run it from source (requires Xcode):

```bash
cd macos/LecternApp
swift run            # opens the app; ⌘O to choose a project folder
```

Package a distributable `.app` yourself:

```bash
macos/LecternApp/scripts/package.sh 0.1.0   # → macos/LecternApp/dist/Lectern.app + .zip
```

> **First launch (preview builds):** preview builds are ad-hoc signed but not yet
> notarized, so macOS blocks the first open. Easiest fix (any macOS version):
> `xattr -dr com.apple.quarantine /Applications/Lectern.app`, then open it.
> Or use **System Settings → Privacy & Security → Open Anyway**. On **macOS 15
> (Sequoia) and later the old right-click → Open no longer bypasses Gatekeeper**,
> so use one of those. This step goes away once notarized builds ship.

---

## Usage

Run from the root of an OpenSpec project:

```bash
lectern
```

View a single change directory by path:

```bash
lectern /path/to/openspec/changes/my-change
```

### Keyboard reference

Press `?` on any screen to open an in-app overlay listing these shortcuts, grouped by screen. Press `?`, `Esc`, or `q` to close it.

#### Normal mode (viewing a change)

| Key | Action |
|---|---|
| `h` / `l` | Previous / next change |
| `1` | Proposal tab |
| `2` | Design tab |
| `3` | Specs tab |
| `4` | Tasks tab |
| `Tab` / `Shift+Tab` | Next / previous tab |
| `←` / `→` | Previous / next tab (mirrors `Shift+Tab` / `Tab`) |
| `[` / `]` | Previous / next spec (specs tab, when the change has multiple specs; or click a spec chip) |
| `j` / `down` | Scroll down (or move task cursor down) |
| `k` / `up` | Scroll up (or move task cursor up) |
| `Space` | Toggle task under cursor (tasks tab only) |
| `e` | Open artifact in `$EDITOR` |
| `i` | Open project config view |
| `a` / `Esc` | Enter index mode |
| `?` | Toggle keyboard-shortcut help |
| `q` / `Ctrl+C` | Quit |

#### Index mode (change and spec navigator)

| Key | Action |
|---|---|
| `j` / `down` | Move cursor down |
| `k` / `up` | Move cursor up |
| `Enter` | Open selected change, spec, or archived change |
| `Space` | Expand / collapse a project spec |
| `/` | Filter the list (type to narrow; `Enter` confirms, `Esc` cancels) |
| `s` | Toggle spec sort (by name / by suffix) |
| `w` | Open the worktrees view |
| `i` | Open project config view |
| `?` | Toggle keyboard-shortcut help |
| `Esc` | Clear active filter, otherwise quit |
| `q` / `Ctrl+C` | Quit |

#### Worktrees mode (cross-worktree overview)

Surveys the git worktrees of the current repository — useful when several agents
work in sibling worktrees at once. Each worktree is listed with its branch (or a
short HEAD SHA when detached) and its active changes with live task progress; the
current worktree is shown first and badged `(current)`. Foreign changes open
**read-only** (task toggling and in-place edits are disabled); `e` still opens the
artifact in `$EDITOR`. Requires `git` on `PATH` — otherwise the view shows a single
"unavailable" line.

| Key | Action |
|---|---|
| `j` / `k` | Move cursor down / up |
| `Enter` | Open the selected change read-only |
| `e` | Open artifact in `$EDITOR` |
| `?` | Toggle keyboard-shortcut help |
| `a` / `Esc` | Return to index |
| `q` / `Ctrl+C` | Quit |

#### Archive mode (viewing an archived change)

| Key | Action |
|---|---|
| `1`–`4` | Switch artifact tab |
| `Tab` / `Shift+Tab` / `←` / `→` | Cycle artifact tabs |
| `j` / `k` | Scroll |
| `e` | Open artifact in `$EDITOR` |
| `i` | Open project config view |
| `a` / `Esc` | Return to index |
| `?` | Toggle keyboard-shortcut help |
| `q` / `Ctrl+C` | Quit |

#### Spec viewer mode

| Key | Action |
|---|---|
| `j` / `k` | Scroll |
| `e` | Open spec in `$EDITOR` |
| `Esc` | Return to index |
| `?` | Toggle keyboard-shortcut help |
| `q` / `Ctrl+C` | Quit |

In requirement focus mode:

| Key | Action |
|---|---|
| `h` / `l` | Previous / next requirement |
| `j` / `k` | Scroll |
| `e` | Open spec in `$EDITOR` |
| `Esc` | Return to index |
| `?` | Toggle keyboard-shortcut help |
| `q` / `Ctrl+C` | Quit |

#### Project config view

| Key | Action |
|---|---|
| `j` / `k` | Scroll |
| `i` / `Esc` | Return to the previous screen |
| `?` | Toggle keyboard-shortcut help |
| `q` / `Ctrl+C` | Quit |

---

## Configuration

lectern reads an optional per-user config file at `$XDG_CONFIG_HOME/lectern/config.toml`
(falling back to `~/.config/lectern/config.toml`). It is optional — a missing file uses
defaults, and a malformed one prints a warning and falls back to defaults rather than
failing to launch.

Today it controls how `e` opens an artifact:

```toml
[editor]
# How to open the active artifact with `e`:
#   "$EDITOR"  (default) — your $EDITOR, falling back to vi, in the terminal
#   "system"            — the OS default app for the file (open / xdg-open / start), detached
#   "nvim", "code --wait", … — any editor command, run in the terminal
open_with = "system"
```

With `open_with = "system"`, pressing `e` hands the file to your default Markdown app and
lectern keeps running — your edit shows up on the next reload. The default stays `$EDITOR`,
so terminal-only and SSH sessions are unaffected.

Press `c` inside the TUI to open this config file in your editor (it's created with a
commented template if it doesn't exist yet); changes take effect on the next launch.

---

## Project structure

lectern expects an `openspec/` directory at the project root:

```
openspec/
├── changes/
│   ├── <change-name>/
│   │   ├── .openspec.yaml   # Required: identifies the directory as a change
│   │   ├── proposal.md
│   │   ├── design.md
│   │   ├── tasks.md         # GFM checkbox syntax: - [ ] / - [x]
│   │   └── specs/
│   │       └── <spec-name>/
│   │           └── spec.md
│   └── archive/
│       └── YYYY-MM-DD-<name>/
└── specs/
    └── <spec-name>/
        └── spec.md          # Requirements parsed from: ### Requirement: <name>
```
