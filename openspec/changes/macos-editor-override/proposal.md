## Why

The macOS app opens an artifact in the system default `.md` app via `NSWorkspace.open`, reachable only through a toolbar flyout — there's no way to pick a *specific* editor, and no keyboard shortcut (#110). Users who prefer a particular Markdown app (or want a quick keystroke) have to change their system default or hunt the flyout. This adds an editor-app override in Settings and a `⌘E` shortcut. It's also the first real preference beyond font scale (the concrete first slice of #100).

## What Changes

- **Editor override (Settings):** an "Opening artifacts" preference to choose a specific application via an app picker (`NSOpenPanel`, filtered to applications). The chosen app's path is stored in `@AppStorage`; "Use Default" clears it. When unset (or the chosen app is missing), opening falls back to `NSWorkspace.open` (today's default-app behavior).
- **Open action honors the override:** opening the current artifact uses the configured app (`NSWorkspace.open(_:withApplicationAt:configuration:)`) when set, else the system default. The logic moves to `AppModel.openCurrentArtifactExternally()` so the toolbar action and a menu command share it.
- **Keyboard shortcut + menu item:** a File-menu "Open in Editor" command bound to `⌘E` (routed to the focused window), and the toolbar flyout item relabeled "Open in Editor". Disabled when there's no file-backed selection.
- Cross-worktree (read-only) artifacts still open externally — read-only only blocks in-app editing.

## Capabilities

### Modified Capabilities
- `macos-app`: "Reveal and open the selected file" gains the editor-app override and the `⌘E` shortcut; a new requirement specifies the editor-app preference (picker, persistence, default fallback).

## Impact

- **Code:** `macos/LecternApp/Sources/LecternApp/Preferences.swift` (an `EditorPref` key + an "Opening artifacts" Settings section with the app picker), `AppModel.swift` (`openCurrentArtifactExternally()` honoring the override), `ContentView.swift` (flyout calls the model method, relabeled), `App.swift` (the `⌘E` File-menu command via the focused model).
- No new dependency (AppKit `NSWorkspace`/`NSOpenPanel`, `UniformTypeIdentifiers`). No OpenSpecKit/Go change.
- **Tests:** AppKit/Settings UI isn't headlessly testable; covered by `swift build` + manual QA.

## Non-goals

- Broader Preferences expansion beyond the editor setting — that's the rest of #100.
- In-app editing (#97, shipped) and default-open (already works).
- A command-string editor (TUI's model, #95) — macOS picks an app, not a shell command.
