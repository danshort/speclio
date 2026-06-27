## Why

You can only view one project at a time — opening another replaces the first (#69). The whole app is built around a single app-level `AppModel` (one `rootPath`, one `project`, one security-scoped bookmark, one FSEvents watcher), so there's no way to compare two repos' OpenSpec data side by side. The single shared model is also why macOS's built-in "Show Tab Bar" / New Tab is dead — every tab clones the same project.

## What Changes

- **One project per window.** Move `AppModel` from a single app-level `@StateObject` to one instance **per window** (value-based `WindowGroup`), so each window has independent navigation state (mode, selection, sidebar, watcher). The per-window project label (sidebar header + title) from #70 already fits this exactly.
- **Open without closing, one window per project.** "Open Project…" opens a project without disturbing others. Each project has a **single window** — opening a project that's already open focuses its existing window instead of duplicating it (a natural consequence of value-based windows keyed by project). Opening into an empty (no-project) window reuses it; otherwise a new window opens. Native window tabbing is kept on, so New Tab / the Tab Bar now genuinely opens *different* projects (resolving the dead menu item).
- **Reopen all on launch.** The app remembers the set of open projects and reopens a window for each on next launch (replacing the single-bookmark restore).
- **Open Recent.** A standard File ▸ Open Recent submenu lists recently opened projects; selecting one opens (or focuses) it. Backed by the same per-project bookmark store, with a Clear Menu item.
- **Per-window commands.** Menu commands (Open, Reload, Reveal/Open file) act on the focused window's project via SwiftUI focused values. The content text-size preference stays global (`@AppStorage`).

## Non-goals

- Combined multi-root sidebar (all projects in one window) or an in-window project switcher — explicitly chose separate windows/tabs (Option A).
- Cross-project features (global search, comparing/diffing across projects) — each window stays independent for now.
- Sandboxing changes — still non-sandboxed but bookmark-based (Option C), now with one security-scoped bookmark per open project.

## Capabilities

### Modified Capabilities

- `macos-app`: adds multiple simultaneous project windows, open-without-closing, and reopen-all-on-launch restore.

## Impact

- `macos/LecternApp` — `App.swift` (value-based `WindowGroup`, focused-value command routing, multi-bookmark store + restore), a new per-window root that owns `AppModel`, `AppModel` open/restore/bookmark logic generalized from one project to a keyed set. No `OpenSpecKit` / `internal/openspec` / corpus changes — the domain loader already loads an arbitrary project path.
